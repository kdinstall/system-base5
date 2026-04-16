package tmpl

import (
	"html/template"
	"path/filepath"
)

// LoadTemplates は指定ディレクトリの全 .html ファイルをロードする
func LoadTemplates(dir string) (*template.Template, error) {
	pattern := filepath.Join(dir, "*.html")
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
