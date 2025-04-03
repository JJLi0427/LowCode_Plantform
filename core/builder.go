package core

import (
	"bytes"
	"fmt"
	"io"
	"lowcode/core/common/tmpl"
	"lowcode/core/control"
	"lowcode/core/cvuecompiler"
	"lowcode/core/sformcompiler"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type HTMLBuilder struct {
	DebugApp    string
	RootPath    string
	tmplBuilder *Tmpl
	control     *control.Control
	sform       sformcompiler.SFormBuilder
	vue         cvuecompiler.VueBuilder
	sync.Mutex
}

func (h *HTMLBuilder) HasDebugApp() bool {
	return len(h.DebugApp) > 0
}

func (h *HTMLBuilder) IsDebugApp(ctl *control.Control) bool {
	ret := false
	if len(h.DebugApp) > 0 {
		ret = strings.Contains(ctl.Name, h.DebugApp)
		if h.DebugApp != ctl.Name && ret {
			h.DebugApp = ctl.Name
		}
	}
	return ret
}

func (h *HTMLBuilder) Init() error {
	if len(h.RootPath) == 0 {
		h.RootPath = "."
	}

	if h.tmplBuilder == nil {
		h.tmplBuilder = NewTmpl()
		funcMap := make(map[string]interface{})
		funcMap["join"] = func(s []string, sep string) string { return strings.Join(s, sep) }
		funcMap["linkAttributes"] = tmpl.LinkAttributes
		funcMap["scriptAttributes"] = tmpl.ScriptAttributes
		funcMap["validateMeta"] = tmpl.ValidateMeta
		funcMap["title"] = tmpl.AppTitle

		h.tmplBuilder.SetFuncs(funcMap)
		if err := h.tmplBuilder.Load(path.Join(h.RootPath, "views"), ".tmpl"); err != nil {
			return err
		}
	}
	if err := h.vue.LoadAllComponents(path.Join(h.RootPath, "views", "components", "vue")); err != nil {
		return err
	}

	return nil
}

func (h *HTMLBuilder) ParseControlFile(controlfile string) (*control.Control, error) {
	ctl := &control.Control{}
	if err := ctl.Parse(controlfile); err != nil {
		return nil, err
	}
	ctl.ControlFilePath = filepath.Dir(controlfile)
	if (len(ctl.Entrypoint.Workdir) > 0 && ctl.Entrypoint.Workdir[0] != '/') || len(ctl.Entrypoint.Workdir) == 0 {
		if ctl.Entrypoint.Workdir == "." {
			ctl.Entrypoint.Workdir = ""
		}
		if apath, err := filepath.Abs(path.Join(ctl.ControlFilePath, ctl.Entrypoint.Workdir)); err == nil {
			ctl.Entrypoint.Workdir = apath
		} else {
			return nil, fmt.Errorf("%s, %s", controlfile, err.Error())
		}
	}

	for i, task := range ctl.Backtasks {
		//if len(task.Cmd) == 0 && len(task.Inline_shell) == 0 {
		//	continue
		//}

		if (len(task.Workdir) > 0 && task.Workdir[0] != '/') || len(task.Workdir) == 0 {
			if task.Workdir == "." {
				ctl.Backtasks[i].Workdir = ""
			}
			if apath, err := filepath.Abs(path.Join(ctl.ControlFilePath, task.Workdir)); err == nil {
				ctl.Backtasks[i].Workdir = apath
			} else {
				return nil, fmt.Errorf("%s, %s", controlfile, err.Error())
			}
		}
	}

	if h.IsDebugApp(ctl) {
		ctl.Head.Equivs = append(ctl.Head.Equivs, control.Equiv{Equiv: "pragma", Content: "no-cache"})
		ctl.Head.Equivs = append(ctl.Head.Equivs, control.Equiv{Equiv: "Cache-Control", Content: "no-cache"})
	}

	return ctl, nil
}

func (h *HTMLBuilder) Generate(control *control.Control) (*control.Control, error) {
	defer h.Unlock()
	h.Lock()
	h.control = control
	//if err := h.control.Parse(controlfile); err != nil {
	//	return nil, err
	//}
	//h.control.ControlFilePath = filepath.Dir(controlfile)

	//TODO: auto refresh html page, only if file changed
	//if h.DebugApp == h.control.Name {
	//	h.control.SaveJS("function selfrefresh(){window.location.reload();}setTimeout('selfrefresh()',1500);", false)
	//}

	if err := h.control.Validate(); err != nil {
		return nil, err
	}

	if err := h.sform.Generate(h.control); err != nil {
		return nil, err
	}

	if err := h.vue.Generate(h.control); err != nil {
		return nil, err
	}

	layout := h.control.Layout
	if len(layout) == 0 {
		layout = "views/layout/default.tmpl"
	} else {
		if fin, err := os.Stat(layout); err != nil || fin.IsDir() {
			newlayout := path.Join(h.control.ControlFilePath, layout)
			if fin, err = os.Stat(newlayout); err == nil && !fin.IsDir() {
				layout = newlayout
				h.control.Layout = layout
			} else {
				layout = path.Join("views/layout", layout)
				h.control.Layout = layout
			}
		}
	}
	if name, err := h.tmplBuilder.TmplName(layout); err == nil {
		var bts bytes.Buffer
		if err := h.control.Save(); err != nil {
			return nil, err
		}
		if err := h.tmplBuilder.Render(&bts, name, &h.control); err != nil {
			return nil, err
		}

		if w, err := h.control.SaveHtml("index.html"); err == nil {
			defer w.Close()
			if _, err = io.Copy(w, &bts); err != nil {
				return nil, err
			}
		}
	} else {
		return nil, err
	}

	return h.control, nil
}
