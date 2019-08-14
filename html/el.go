package html

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

var trim = strings.TrimSpace
var hasPrefix = strings.HasPrefix

func executeEl(t *html.Node, handler HTMLHandler) (*ElementInfo, string, error) {
	if t.Type == html.CommentNode {
		return &ElementInfo{IsComment: true}, "/*" + t.Data + "*/", nil
	}

	if t.Type != html.ElementNode {
		return nil, returnOutText(t.Data), nil
	}

	switch t.Data {
	// e, g-slot, g-body -- tags for components-based programming
	case "e":
		runAttribute := getXAttr(t.Attr, "run")
		if len(runAttribute) == 0 {
			return nil, "", fmt.Errorf("invalid external component: no 'run' attribute")
		}

		out := runAttribute

		var slots, templates string
		beforeHandler := func(c *html.Node) (bool, error) {
			if c.Type != html.ElementNode {
				return true, nil
			}

			if c.Data == "template" {
				tName := getXAttr(c.Attr, "name")
				if len(tName) == 0 {
					return false, errors.New("ERROR: invalid template syntax: 'name' attribute undefined")
				}

				templateBody, err := genChildes(getElChildes(c), nil, handler)
				if err != nil {
					return false, err
				}

				tTypes := getXAttr(c.Attr, "types")
				if len(tName) != 0 {
					tTypes = "\n" + tTypes + "\n"
				}

				templates = templates + `"` + tName + `": func(values ...interface{}) []interface{} {` + tTypes + `return []interface{} {` + templateBody + `} },`
				return false, nil
			}

			slotAttr := getXAttr(c.Attr, "slot")
			if len(slotAttr) != 0 {
				_, cOut, err := executeEl(c, handler)
				if err != nil {
					return false, err
				}

				if len(cOut) == 0 {
					cOut = "``"
				}

				slots += `"` + slotAttr + `":` + cOut + ","
				return false, nil
			}

			return true, nil
		}

		body, err := genChildes(getElChildes(t), beforeHandler, handler)
		if err != nil {
			return nil, "", err
		}

		var external string

		if len(body) != 0 {
			external += "Body: []interface{}{" + body + "},"
		}

		if len(slots) != 0 {
			external += "Slots: map[string]interface{}{" + slots + "},"
		}

		if len(templates) != 0 {
			external += "Templates: map[string]gas.Template{" + templates + "},"
		}

		if len(t.Attr) > 1 { // not only "run" attribute
			var attrsString string
			for _, attr := range t.Attr {
				if attr.Key == "run" {
					continue
				}

				attrsString += `"` + attr.Key + `": "` + attr.Val + `",`
			}
			external += "Attrs: func() gas.Map { return gas.Map {" + attrsString + "} },"
		}

		if len(external) > 0 {
			external = "gas.External{" + external
			if out[len(out)-2] != '(' {
				external = ", " + external
			}

			out = out[:len(out)-1] + external + "})"
		}

		return &ElementInfo{Tag: "e"}, out, nil
	case "g-slot":
		name := getXAttr(t.Attr, "name")
		if len(name) == 0 {
			return nil, "", errors.New("slot name is undefined")
		}
		return &ElementInfo{Tag: "g-slot"}, `e.Slots["` + name + `"]`, nil
	case "g-body":
		return &ElementInfo{Tag: "g-body"}, `gas.NE(&gas.E{}, e.Compile)`, nil
	case "g-switch":
		runAttribute := getXAttr(t.Attr, "run")
		if len(runAttribute) == 0 {
			return nil, "", errors.New("invalid g-switch: \"run\" attribute is undefined")
		}

		var switchOut string
		beforeHandler := func(c *html.Node) (bool, error) {
			attrs, cOut, err := executeEl(c, handler)
			if err != nil {
				return false, err
			}

			if attrs == nil {
				return false, nil
			}

			if len(attrs.CaseData) == 0 && !attrs.CaseDefaultData {
				return false, errors.New("invalid g-switch child")
			}

			if len(attrs.CaseData) != 0 {
				switchOut += "case " + attrs.CaseData + ": \n\treturn " + cOut + "\n"
			}

			if attrs.CaseDefaultData {
				switchOut += "default: \n\treturn " + cOut + "\n"
			}

			return false, nil
		}

		_, err := genChildes(getElChildes(t), beforeHandler, handler)
		if err != nil {
			return nil, "", err
		}

		return &ElementInfo{Tag: "g-switch"}, "func()interface{}{\n\tswitch " + runAttribute + " {\n" + switchOut + "}\n\treturn nil\n}()", nil
	}

	elementInfo := GetElementInfo(t.Data, t.Attr, handler)
	tBody := elementInfo.BuildBody()

	childesOut, err := genChildes(getElChildes(t), nil, handler)
	if err != nil {
		return nil, "", err
	}

	elOut := returnOutElement(tBody + childesOut)

	if len(elementInfo.ForData) != 0 {
		return elementInfo, returnOutFOR(elementInfo.ForData, elOut), nil
	}

	return elementInfo, elOut, nil
}

func getElChildes(t *html.Node) []*html.Node {
	childes := []*html.Node{}
	for c := t.FirstChild; c != nil; c = c.NextSibling {
		childes = append(childes, c)
		t.FirstChild = c.NextSibling
	}
	return childes
}

func genChildes(childes []*html.Node, beforeHandler func(*html.Node) (bool, error), handler HTMLHandler) (string, error) {
	var (
		haveIf     bool
		logicBlock string // if, else, else if
		mainBlock  string
	)

	closeLogicBlock := func() {
		logicBlock = " func()interface{} { " + logicBlock + "; return nil }(),"

		mainBlock = logicBlock + mainBlock
		logicBlock = ""
		haveIf = false
	}

	for _, c := range childes {
		if beforeHandler != nil {
			needContinue, err := beforeHandler(c)
			if err != nil {
				return "", err
			}

			if !needContinue {
				continue
			}
		}

		info, cOut, err := executeEl(c, handler)
		if err != nil {
			return "", err
		}

		if len(cOut) == 0 {
			continue
		}

		needComma := true

		if info == nil {
			cOutTrimmed := trim(cOut)
			if cOutTrimmed[len(cOutTrimmed)-1] == ',' {
				needComma = false
			}
		} else {
			needComma = !info.IsComment

			switch {
			case len(info.IfData) != 0:
				if haveIf {
					closeLogicBlock()
				}

				logicBlock = "if " + info.IfData + " { return " + cOut + " }"
				haveIf = true
				continue
			case len(info.ElseIfData) != 0:
				if !haveIf {
					return "", errors.New("invalid g-else-if: no g-if before")
				}

				logicBlock += " else if " + info.ElseIfData + " { return " + cOut + " }"
				continue
			case info.ElseData:
				if !haveIf {
					return "", errors.New("invalid g-else: no g-if before")
				}

				logicBlock += " else { return " + cOut + " }"

				closeLogicBlock()
				continue
			}
		}

		mainBlock += cOut
		if needComma {
			mainBlock += ","
		}

		if haveIf {
			closeLogicBlock()
		}
	}

	if len(logicBlock) != 0 && haveIf {
		logicBlock = "func()interface{} { " + logicBlock + "; return nil }(),"
		mainBlock = mainBlock + logicBlock
	}

	return mainBlock, nil
}

func indexOfLastByteInString(a string, x rune) int {
	lastIndex := -1
	for i, el := range a {
		if el == x {
			lastIndex = i
		}
	}
	return lastIndex
}

func parseText(in string, i int, buf string) string {
	if i > 0 && in[i-1] == '\\' {
		return buf + returnOutStatic(removeInArr([]byte(in), i-1))
	}

	closeI := strings.Index(in, "}}")
	if closeI == -1 {
		return buf + returnOutStatic(in)
	}

	s1 := in[:i]        // was edit in past
	s3 := in[closeI+2:] // will be edited

	s2 := in[i : closeI+2] // editing now
	s2 = s2[2 : len(s2)-2] // remove first and last 2 chars "{{" and "}}"

	buf = buf + "`" + trim(s1) + "`," + s2 + ", "

	startI := strings.Index(s3, "{{")

	switch {
	case startI != -1:
		return parseText(s3, startI, buf)
	case startI == -1 && stringNotEmpty(s3):
		return buf + "`" + trim(s3) + "`"
	default:
		return buf
	}
}

func stringNotEmpty(a string) bool {
	cleared := trim(a)
	return len(cleared) != 0
}

func removeInArr(a []byte, i int) string {
	copy(a[i:], a[i+1:]) // Shift a[i+1:] left one index
	a[len(a)-1] = 0      // Erase last element (write zero value)
	a = a[:len(a)-1]     // Truncate slice
	return string(a)
}
