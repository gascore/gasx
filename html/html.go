package html

import (
	"bytes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type HTMLCompiler struct {
	onAttribute []func(key,val string)
	// TODO: Add onElement, on......	
}

func NewCompiler() *HTMLCompiler {
	return &HTMLCompiler{}
}

func (c *HTMLCompiler) AddOnAttribute(f func(string,string)) {
	c.onAttribute = append(c.onAttribute, f)
}

func (c *HTMLCompiler) runOnAttribute(key,val string) {
	for _, f := rnage c.onAttribute {
		f(key,val)
	}
}

func (c *HTMLCompiler) Block() builder.BlockCompiler {
	return func(info *BlockInfo) (string, error) {
		if info.Name != "html" && info.Name != "htmlF" {
			return info.FileBytes, nil
		}

		buf := bytes.NewBufferString(info.FileBytes)
		nodes, err := html.ParseFragment(reader, &html.Node{
			Type:     html.ElementNode,
			Data:     "div",
			DataAtom: atom.Div,
		})
		if err != nil {
			return "", fmt.Errorf("error while parsing html block: %s", err.Error())
		}

		var out string
		for _, node := range nodes {
			_, compiledNode, err := executeEl(node)
			if err != nil {
				return "", fmt.Errorf("error while compiling html node: %s", err.Error())
			}

			out += compiledNode + ", "
		}
		out = "gas.CL("+out+")"

		if info.Name == "htmlF" {
			return "func() []interface{return "+out+"}"
		}

		return out
	}
}