package html

import (
	"strings"

	"golang.org/x/net/html"
)

const (
	gDirs = "g-"

	gOn    = gDirs + "on:"
	gOnAlt = "@"

	gBind    = gDirs + "bind:"
	gBindAlt = ":"

	gRef = gDirs + "ref"

	gHTML = gDirs + "html"

	gFor = gDirs + "for"

	gIf     = gDirs + "if"
	gElseIf = gDirs + "else-if"
	gElse   = gDirs + "else"

	gCase        = gDirs + "case"
	gCaseDefault = gDirs + "default"

	gIsPointer = gDirs + "pointer"
)

func getXAttr(attrs []html.Attribute, x string) string {
	for _, a := range attrs {
		if a.Key == x {
			return a.Val
		}
	}
	return ""
}

type ElementInfo struct {
	Tag string

	IsComment bool

	Attrs map[string]string

	Handlers map[string]string
	Binds    map[string]string

	Ref string

	HTMLRender string

	ForData string

	IfData     string
	ElseData   bool
	ElseIfData string

	SwitchData      string
	CaseData        string
	CaseDefaultData bool

	IsPointer bool
}

// BuildBody return string with gas element strucutre
func (info *ElementInfo) BuildBody() string {
	return `&gas.E{Tag:"` + info.Tag + `", ` + info.RenderHandlers() + info.RenderAttrs() + info.RenderHTMLDir() + info.RenderRef() + info.RenderIsPointer() + `},`
}

// RenderRef return RefName for Element
func (info *ElementInfo) RenderRef() string {
	if len(info.Ref) == 0 {
		return ""
	}

	return `RefName: "` + info.Ref + `",`
}

// RenderHandlers generate Handlers for Element
func (info *ElementInfo) RenderHandlers() string {
	if len(info.Handlers) == 0 {
		return ""
	}

	var out string
	for key, val := range info.Handlers {
		out += `"` + key + `": func(e gas.Event) {` + val + `},`
	}

	return `Handlers: map[string]gas.Handler{` + out + `},`
}

// RenderAttrs generate Attrs for Element
func (info *ElementInfo) RenderAttrs() string {
	if len(info.Attrs) == 0 && len(info.Binds) == 0 {
		return ""
	}

	var out string
	for key, val := range info.Attrs {
		out += `"` + key + `": "` + val + `",`
	}

	for key, val := range info.Binds {
		out += `"` + key + `": ` + val + `,`
	}

	return `Attrs: func() gas.Map { return gas.Map{` + out + `} },`
}

// RenderHTMLDir generate HTML directive for Element
func (info *ElementInfo) RenderHTMLDir() string {
	if len(info.HTMLRender) == 0 {
		return ""
	}

	return `HTML: gas.HTMLDirective{Render: func() string { return ` + info.HTMLRender + `},},`
}

// RenderIsPointer generate IsPointer for Element
func (info *ElementInfo) RenderIsPointer() string {
	if !info.IsPointer {
		return ""
	}

	return `IsPointer: true,`
}

// GetElementInfo generate ElementInfo from html.Element.Attr
func GetElementInfo(tag string, attrs []html.Attribute, handler HTMLHandler) *ElementInfo {
	info := &ElementInfo{
		Tag:      tag,
		Handlers: make(map[string]string),
		Binds:    make(map[string]string),
		Attrs:    make(map[string]string),
	}

	for _, attr := range attrs {
		aKey := attr.Key
		aVal := attr.Val

		handler.runOnAttribute(aKey, aVal)

		if hasPrefix(aKey, gOn) || hasPrefix(aKey, gOnAlt) {
			var current string
			if strings.HasPrefix(aKey, gOn) {
				current = gOn
			} else {
				current = gOnAlt
			}

			info.Handlers[strings.TrimPrefix(aKey, current)] = aVal
			continue
		}

		if hasPrefix(aKey, gBind) || hasPrefix(aKey, gBindAlt) {
			var current string
			if strings.HasPrefix(aKey, gBind) {
				current = gBind
			} else {
				current = gBindAlt
			}

			info.Binds[strings.TrimPrefix(aKey, current)] = aVal
			continue
		}

		switch aKey {
		case gRef:
			info.Ref = aVal
		case gFor:
			if !strings.Contains(aVal, "=") {
				aVal = "key, val := " + aVal
			}
			info.ForData = aVal
		case gIf:
			info.IfData = aVal
		case gElseIf:
			info.ElseIfData = aVal
		case gElse:
			info.ElseData = true
		case gHTML:
			info.HTMLRender = aVal
		case gCase:
			info.CaseData = aVal
		case gCaseDefault:
			info.CaseDefaultData = true
		case gIsPointer:
			info.IsPointer = true
		default:
			info.Attrs[aKey] = aVal
		}
	}

	return info
}
