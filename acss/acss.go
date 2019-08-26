package acss

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/gascore/gasx/html"
)

var styleRgxp = regexp.MustCompile(`([a-zA-Z]*){(.*?)}(:[a-z]*|)(@([a-z]*)|)`)

type Generator struct {
	Styles strings.Builder

	Exceptions  []string
	BreakPoints map[string]string
	Custom      map[string]string
}

func (g *Generator) Init() {
	if g.BreakPoints == nil {
		g.BreakPoints = make(map[string]string)
	}

	if g.Custom == nil {
		g.Custom = make(map[string]string)
	}
}

type acssStyle struct {
	body        string
	media       string
	pseudoClass string
}

func (g *Generator) OnElementInfo() func(*html.ElementInfo) {
	rand.Seed(time.Now().UnixNano())
	return func(info *html.ElementInfo) {
		acssAttr := info.Attrs["acss"]
		if acssAttr == "" {
			return
		}

		classID   := "A" + randID(8)
		outStyles := g.GenCSS(classID, acssAttr)
		
		delete(info.Attrs, "acss")
		info.Attrs["data-acss"] = acssAttr
		info.Attrs["data-acss-id"] = classID
		info.Attrs["class"] += " " + classID

		g.Styles.WriteString(outStyles)
	}
}

func (g *Generator) GenCSS(classID, acssAttr string) string {
	id := fmt.Sprintf(".%s, [data-acss-id=\"%[1]s\"] ", classID)
	idWithP := func(p string) string {
		return fmt.Sprintf(".%s:%s, [data-acss-id=\"%[1]s\"]:%[2]s ", classID, p)
	}

	type mediaVal struct {
		basic  strings.Builder
		pseudo map[string]strings.Builder
	}

	var (
		outStyles = strings.Builder{}
		basic     = strings.Builder{}
		media     = make(map[string]mediaVal)
		pseudo    = make(map[string]strings.Builder)
	)

	for _, class := range styleRgxp.FindAllString(acssAttr, -1) {
		var pOperator string
		if strings.Contains(class, ":") {
			pIndex := strings.Index(class, ":")
			pOperator = class[1+pIndex : 2+pIndex]
			class = class[:pIndex] + class[2+pIndex:]

			switch pOperator {
			case "a", "active":
				pOperator = "active"
				break
			case "c", "checked":
				pOperator = "checked"
				break
			case "f", "focus":
				pOperator = "focus"
				break
			case "h", "hover":
				pOperator = "hover"
				break
			case "d", "disabled":
				pOperator = "disabled"
				break
			}
		}

		var breakPoint string
		if strings.Contains(class, "@") {
			breakPoint = g.BreakPoints[class[strings.Index(class, "@")+len("@"):]]
			class = class[:strings.Index(class, "@")]
		}

		styleValue := GenerateStyleForClass(class, g.Custom)
		if len(styleValue) == 0 {
			continue
		}

		if pOperator == "" && breakPoint == "" {
			basic.WriteString("\n\t" + styleValue)
			continue
		}

		if pOperator != "" {
			if breakPoint != "" {
				mVal := media[breakPoint]
				if mVal.pseudo == nil {
					mVal.pseudo = make(map[string]strings.Builder)
				}

				p := mVal.pseudo[pOperator]
				p.WriteString("\n\t\t" + styleValue)
				mVal.pseudo[pOperator] = p

				media[breakPoint] = mVal

				continue
			}

			p := pseudo[pOperator]
			p.WriteString("\n\t" + styleValue)
			pseudo[pOperator] = p

			continue
		}

		if breakPoint != "" {
			mVal := media[breakPoint]
			mVal.basic.WriteString("\n\t\t" + styleValue)
			media[breakPoint] = mVal
		}
	}

	// basic
	outStyles.WriteString(id + "{" + basic.String() + "\n}\n")

	// pseudo classes
	for pKey, pVal := range pseudo {
		outStyles.WriteString(idWithP(pKey) + "{" + pVal.String() + "\n}\n")
	}

	// media
	for mKey, mVal := range media {
		mStyles := strings.Builder{}
		mStyles.WriteString(id + "{" + mVal.basic.String() + "\t\n}\n")
		for pKey, pVal := range mVal.pseudo {
			mStyles.WriteString(idWithP(pKey) + "{" + pVal.String() + "\t\t\n}\n")
		}
		outStyles.WriteString(mKey + "{\n\t" + mStyles.String() + "\t\n}\n")
	}

	return outStyles.String()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (g *Generator) GetStyles() string {
	// Some logic?

	out := g.Styles.String()
	g.Styles.Reset()

	return out
}
