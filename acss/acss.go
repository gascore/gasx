package acss

import (
	"regexp"
	"strings"
	"github.com/gascore/gasx"
)

var styleRgxp = regexp.MustCompile(`([a-zA-Z]*)\((.*?)\)(:[a-z]|)(--([a-z]*)|)`)

type Generator struct {
	LockFile *gasx.LockFile
	Styles   string

	Exceptions  []string
	BreakPoints map[string]string
	Custom      map[string]string
}

func (g *Generator) OnAttribute() func(string, string, *gasx.BlockInfo) {
	if g.LockFile.BuildExternal {
		g.LockFile.Body["acss"] = ""
	}

	generated := make(map[string]bool)

	return func(key, val string, info *gasx.BlockInfo) {
		if key != "class" {
			return
		}

		for _, class := range styleRgxp.FindAllString(val, -1) {
			if generated[class] {
				continue
			}

			generated[class] = true

			if inArray(class, g.Exceptions) {
				continue
			}

			classOut := g.buildClass(class)
			if classOut == "" {
				continue
			}

			g.Styles += classOut
			if info.FileInfo.IsExternal {
				g.LockFile.Body["acss"] += classOut
			}
		}
	}
}

func (g *Generator) buildClass(class string) string {
	rawClass := class

	var pOperator string
	if strings.Contains(class, ":") {
		pIndex := strings.Index(class, ":")
		pOperator = class[1+pIndex : 2+pIndex]
		class = class[:pIndex] + class[2+pIndex:]

		switch pOperator {
		case "a":
			pOperator = "active"
			break
		case "c":
			pOperator = "checked"
			break
		case "f":
			pOperator = "focus"
			break
		case "h":
			pOperator = "hover"
			break
		case "d":
			pOperator = "disabled"
			break
		}
	}

	var breakPoint string
	if strings.Contains(class, "--") {
		breakPoint = g.BreakPoints[class[strings.Index(class, "--")+len("--"):]]
		class = class[:strings.Index(class, "--")]
	}

	styleValue := GenerateStyleForClass(class, g.Custom)
	if len(styleValue) == 0 {
		return ""
	}

	styleOut := "{\n" + styleValue + "\n}\n"
	styleClass := "." + ClearClass(rawClass)
	if len(pOperator) == 0 {
		styleOut = styleClass + styleOut
	} else {
		styleOut = styleClass + ":" + pOperator + styleOut
	}

	if len(breakPoint) != 0 {
		styleOut = breakPoint + " {\n" + styleOut + "}\n"
	}

	return styleOut
}

func (g *Generator) GetStyles() string {
	if !g.LockFile.BuildExternal {
		return g.Styles + "\n" + g.LockFile.Body["acss"]
	}

	return g.Styles
}

func ClearClass(class string) string {
	class = strings.Replace(class, "(", "\\(", -1)
	class = strings.Replace(class, ")", "\\)", -1)
	class = strings.Replace(class, ",", "\\,", -1)
	class = strings.Replace(class, ":", "\\:", -1)
	class = strings.Replace(class, "#", "\\#", -1)
	class = strings.Replace(class, ".", "\\.", -1)
	class = strings.Replace(class, "%", "\\%", -1)
	class = strings.Replace(class, "!", "\\!", -1)

	return class
}

func inArray(el string, arr []string) bool {
	for _, x := range arr {
		if x == el {
			return true
		}
	}

	return false
}
