package main

import (
	"io"
	"path"
	"text/template"

	"github.com/frizinak/tiif/assets"
)

var tpl *template.Template

func init() {
	tpl = template.New("root")
	tpl.Funcs(
		template.FuncMap{
			"sum":    func(a, b int) int { return a + b },
			"concat": func(a, b string) string { return a + b },
			"trim": func(s string) string {
				runes := []rune(s)
				l := terminalWidth
				if l > len(runes) {
					l = len(runes)
				}

				return string(runes[0:l])
			},
		},
	)

	tpls, err := assets.AssetDir("tpls/base")
	if err != nil {
		panic(err)
	}

	for i, _ := range tpls {
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
	for i, _ := range overrides {
		_tplName := path.Join("providers", overrides[i], name)
		if tpl.Lookup(_tplName) != nil {
			tplName = _tplName
			break
		}

	}
	return tpl.ExecuteTemplate(w, tplName, data)
}
