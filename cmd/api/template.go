package main

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/bueti/shrinkster/ui"
	"github.com/labstack/echo/v4"
)

type templateData struct {
	CurrentYear     int
	Url             *model.Url
	URLs            []*model.Url
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	User            *model.User
}

type Template struct {
	templates map[string]*template.Template
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base", data)
}

func (app *application) initTemplate() map[string]*template.Template {
	templates := make(map[string]*template.Template)
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		panic(err)
	}
	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.html",
			page,
		}

		// Parse the template
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			panic(err)
		}
		templates[name] = ts
	}
	return templates
}
