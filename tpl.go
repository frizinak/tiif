package main

import (
	"io"
	"path"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/frizinak/tiif/assets"
	fstrings "github.com/frizinak/tiif/strings"
)

var tpl *template.Template

func init() {
	tpl = template.New("root")
	tpl.Funcs(
		template.FuncMap{
			"colorLines": func(color, str string) string {
				// unix less resets the color after each newline
				s := strings.Split(str, "\n")
				return color + strings.Join(s, "\n"+color) + "\033[0m"
			},
			"sum":    func(a, b int) int { return a + b },
			"concat": func(a, b string) string { return a + b },
			"indent": func(amount int, s string) string {
				return fstrings.Indent(s, amount)
			},
			"trim": func(s string) string {
				runes := []rune(s)
				var l int
				var o int
				for i := range runes {
					ln := utf8.RuneLen(runes[i])
					if ln+l > terminalWidth {
						break
					}

					o++
					l += ln
				}

				return string(runes[0:o])
			},
			"fold": func(s string) string {
				return fstrings.Fold(s, terminalWidth)
			},
		},
	)

	tpls, err := assets.AssetDir("tpls/base")
	if err != nil {
		panic(err)
	}

	for i := range tpls {
		tpls[i] = path.Join("base", tpls[i])
	}

	provDir := "tpls/providers"
	provs, _ := assets.AssetDir(provDir)
	for _, provider := range provs {
		provTpls, _ := assets.AssetDir(path.Join(provDir, provider))
		for _, provTpl := range provTpls {
			tpls = append(tpls, path.Join("providers", provider, provTpl))
		}
	}

	for _, fn := range tpls {
		template.Must(
			tpl.New(fn).Parse(
				string(assets.MustAsset(path.Join("tpls", fn))),
			),
		)
	}
}

func execTpl(w io.Writer, name string, data interface{}, overrides []string) error {
	name = name + ".txt"
	tplName := path.Join("base", name)
	for i := range overrides {
		_tplName := path.Join("providers", overrides[i], name)
		if tpl.Lookup(_tplName) != nil {
			tplName = _tplName
			break
		}

	}
	return tpl.ExecuteTemplate(w, tplName, data)
}
