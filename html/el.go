package html

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func executeEl(t *html.Node) (*ElementInfo, string, error) {
	if t.Type == html.CommentNode {
		return &ElementInfo{IsComment:true}, "/*" + t.Data + "*/", nil
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

				templateBody, err := genChildes(c, nil)
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
				_, cOut, err := executeEl(c)
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

		body, err := genChildes(t, beforeHandler)
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

				attrsString += `"`+attr.Key+`": "`+attr.Val+`",`
			}
			external += "Attrs: map[string]string{"+attrsString+"},"
		}

		if len(external) > 0 {
			external = "gas.External{" + external
			if out[len(out)-2] != '(' {
				external = ", " + external
			}

			out = out[:len(out)-1] + external + "})"
		}

		return &ElementInfo{Tag:"e"}, out, nil
	case "g-slot":
		name := getXAttr(t.Attr, "name")
		if len(name) == 0 {
			return nil, "", errors.New("slot name is undefined")
		}
		return &ElementInfo{Tag:"g-slot"}, `e.Slots["` + name + `"]`, nil
	case "g-body":
		return &ElementInfo{Tag:"g-body"}, `gas.NE(&gas.E{}, e.Compile)`, nil
	case "g-switch":
		runAttribute := getXAttr(t.Attr, "run")
		if len(runAttribute) == 0 {
			return nil, "", errors.New("invalid g-switch: \"run\" attribute is undefined")
		}

		var switchOut string
		beforeHandler := func(c *html.Node) (bool, error) {
			attrs, cOut, err := executeEl(c)
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
				switchOut += "case "+attrs.CaseData+": \n\treturn "+cOut+"\n"
			}

			if attrs.CaseDefaultData {
				switchOut += "default: \n\treturn "+cOut+"\n"
			}

			return false, nil
		}

		_, err := genChildes(t, beforeHandler)
		if err != nil {
			return nil, "", err
		}

		return &ElementInfo{Tag:"g-switch"}, "func()interface{}{\n\tswitch "+runAttribute+" {\n"+switchOut+"}\n\treturn nil\n}()", nil
	}

	elementInfo := GetElementInfo(t.Data, t.Attr)
	tBody := elementInfo.BuildBody()

	childesOut, err := genChildes(t, nil)
	if err != nil {
		return nil, "", err
	}

	elOut := returnOutElement(tBody + childesOut)

	if len(elementInfo.ForData) != 0 {
		return elementInfo, returnOutFOR(elementInfo.ForData, elOut), nil
	}

	return elementInfo, elOut, nil
}

func genChildes(t *html.Node, beforeHandler func(*html.Node) (bool, error)) (string, error) {
	var haveIf bool
	
	var logicBlock string // if, switch
	var mainBlock string
	
	closeLogicBlock := func() {
		logicBlock = "func()interface{} {\n" + logicBlock + "\nreturn nil\n}(),"
		mainBlock = logicBlock + mainBlock
		
		logicBlock = ""
		haveIf = false
	}

	for c := t.FirstChild; c != nil; c = c.NextSibling {
		if beforeHandler != nil {
			needContinue, err := beforeHandler(c)
			if err != nil {
				return "", err
			}

			if !needContinue {
				continue
			}
		}

		info, cOut, err := executeEl(c)
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

			switch{
			case len(info.IfData) != 0:
				if haveIf {
					closeLogicBlock()
				}

				logicBlock = "if " + info.IfData + " {\n\treturn " + cOut + "\n}"
				haveIf = true
				continue
			case len(info.ElseIfData) != 0:
				if !haveIf {
					return "", errors.New("invalid g-else-if: no g-if before")
				}

				logicBlock += " else if " + info.ElseIfData + " {\n\treturn " + cOut + "\n}"
				continue
			case info.ElseData:
				if !haveIf {
					return "", errors.New("invalid g-else: no g-if before")
				}

				logicBlock += " else {\n\treturn " + cOut + "\n}"

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

		t.FirstChild = c.NextSibling
	}

	if len(logicBlock) != 0 && haveIf {
		logicBlock = "func()interface{} {\n" + logicBlock + "\nreturn nil\n}(),"
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
	s2 = s2[2:len(s2)-2]   // remove first and last 2 chars "{{" and "}}"

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
