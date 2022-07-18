package utils

import (
	"fmt"
	"html/template"
	"io"
	"os"
)

const (
	tmplStorageKey = "TMPL_STORAGE"
)

var (
	builder = &tmplBuilder{}
)

func Tmpl() *tmplBuilder {
	return builder
}

type tmplBuilder struct {
	storage *template.Template
}

// Загрузить шаблоны из директории
func (tb *tmplBuilder) Load() (err error) {
	folder, ok := os.LookupEnv(tmplStorageKey)
	if !ok {
		folder = "tmpl"
	}

	tb.storage, err = template.ParseGlob(fmt.Sprintf("%s/*.tmpl", folder))
	return err
}

// Составить заполненный макет по выбранному шаблону
func (tb *tmplBuilder) Write(wr io.Writer, name string, obj interface{}) error {
	return tb.storage.ExecuteTemplate(wr, name, obj)
}
