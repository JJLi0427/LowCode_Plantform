package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// A Tmpl implements keeper, loader and reloader for HTML templates
type Tmpl struct {
	*template.Template // root template

	dir, ext string
}

// NewTmpl creates new Tmpl.
func NewTmpl() (tmpl *Tmpl) {
	tmpl = new(Tmpl)
	tmpl.Template = template.New("") // unnamed root template
	return
}

// SetFuncs sets template functions to underlying templates
func (t *Tmpl) SetFuncs(funcMap template.FuncMap) {
	t.Template = t.Template.Funcs(funcMap)
}

func (t *Tmpl) TmplName(layoutfile string) (string, error) {
	rel := strings.TrimPrefix(layoutfile, "views/")

	rel = strings.TrimSuffix(rel, t.ext)
	rel = strings.Join(strings.Split(rel, string(os.PathSeparator)), "/")
	return rel, nil
}

// Load templates. The dir argument is a directory to load templates from.
// The ext argument is extension of tempaltes.
func (t *Tmpl) Load(dir, ext string) (err error) {

	// get absolute path
	if dir, err = filepath.Abs(dir); err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}
	t.dir, t.ext = dir, ext

	var root = t.Template

	var walkFunc = func(path string, info os.FileInfo, err error) (_ error) {

		// handle walking error if any
		if err != nil {
			return err
		}

		// skip all except regular files
		// TODO (kostyarin): follow symlinks (?)
		if !info.Mode().IsRegular() {
			return
		}

		// filter by extension
		if filepath.Ext(path) != ext {
			return
		}

		// get relative path
		var rel string
		if rel, err = filepath.Rel(dir, path); err != nil {
			return err
		}

		// name of a template is its relative path
		// without extension
		rel = strings.TrimSuffix(rel, ext)
		rel = strings.Join(strings.Split(rel, string(os.PathSeparator)), "/")

		fmt.Println(rel)
		// load or reload
		var (
			nt = root.New(rel)
			b  []byte
		)

		if b, err = os.ReadFile(path); err != nil {
			return err
		}

		_, err = nt.Parse(string(b))
		return err
	}

	if err = filepath.Walk(dir, walkFunc); err != nil {
		return
	}

	t.Template = root // set or replace (does it needed?)
	return
}

// Render is equal to ExecuteTemplate.
//
// DEPRECATED: use Go native ExeuteTempalte instead.
func (t *Tmpl) Render(w io.Writer, name string, data interface{}) error {
	return t.ExecuteTemplate(w, name, data)
}
