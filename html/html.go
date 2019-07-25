package html

import (
	"bytes"
	"fmt"

	"github.com/gascore/gasx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type HTMLCompiler struct {
	onAttribute []func(key, val string, info *gasx.BlockInfo)
	// TODO: Add onElement, on......
}

func NewCompiler() *HTMLCompiler {
	return &HTMLCompiler{}
}

func (c *HTMLCompiler) AddOnAttribute(f func(string, string, *gasx.BlockInfo)) {
	c.onAttribute = append(c.onAttribute, f)
}

type HTMLHandler struct {
	info *gasx.BlockInfo
	c *HTMLCompiler
}

func (handler *HTMLHandler) runOnAttribute(key, val string) {
	for _, f := range handler.c.onAttribute {
		f(key, val, handler.info)
	}
}

func (c *HTMLCompiler) Block() gasx.BlockCompiler {
	return func(info *gasx.BlockInfo) (string, error) {
		if info.Name != "html" && info.Name != "htmlF" && info.Name != "htmlEl" {
			return info.Value, nil
		}

		nodes, err := html.ParseFragment(bytes.NewBufferString(info.Value), &html.Node{
			Type:     html.ElementNode,
			Data:     "div",
			DataAtom: atom.Div,
		})
		if err != nil {
			return "", fmt.Errorf("error while parsing html block: %s", err.Error())
		}

		handler := HTMLHandler{
			info: info,
			c: c,
		}

		if info.Name == "htmlEl" {
			_, compiledNode, err := executeEl(nodes[0], handler)
			if err != nil {
				return "", fmt.Errorf("error while compiling html node: %s", err.Error())
			}

			return compiledNode, nil
		}

		out, err := genChildes(nodes, nil, handler)
		if err != nil {
			return "", fmt.Errorf("error while compiling html nodes")
		}

		out = "gas.CL(" + out + ")"

		if info.Name == "htmlF" {
			return "func() []interface{} {return " + out + "}", nil
		}

		return out, nil
	}
}
