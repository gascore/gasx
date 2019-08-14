package html

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func returnOutStatic(out string) string {
	return "`" + strings.Replace(out, "\n", "", -1) + "`"
}

func returnOutText(out string) string {
	if len(out) == 0 || len(trim(out)) == 0 {
		return ""
	}

	i := strings.Index(out, "{{")
	if i != -1 {
		return parseText(out, i, "")
	}

	return returnOutStatic(out)
}

func returnOutElement(out string) string {
	return "gas.NE(" + out + ")"
}

func returnOutFOR(data, element string) string {
	rand.Seed(time.Now().UnixNano())
	c := fmt.Sprintf("c%d", rand.Int())
	return fmt.Sprintf("func()[]interface{}{var %s []interface{}; for %s { %s = append(%s, %s) }; return %s}()", c, data, c, c, element, c)
}
