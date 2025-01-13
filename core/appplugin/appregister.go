package appplugin

import (
	"bytes"
	"net/http"
	"onlinetools/core/control"
	"os"
	"strings"
	"sync"
	"time"
)

// regiter app links to http server, generate sitemap
type AppRegisterCenter struct {
	sync.Mutex
	apps map[string]AppPlugin
	hub  *AppHub
	*TaskScheduler
}

func NewAppRegisterCenter() *AppRegisterCenter {
	return &AppRegisterCenter{TaskScheduler: NewScheduler(), hub: DefaultAppHub()}
}

func (a *AppRegisterCenter) Check(ctl *control.Control) error {
	defer a.Unlock()
	a.Lock()
	return nil
}

func (a *AppRegisterCenter) Find(appname string) (AppPlugin, bool) {
	defer a.Unlock()
	a.Lock()
	if _, ok := a.apps[appname]; ok {
		return a.apps[appname], true
	}

	return nil, false
}

func (a *AppRegisterCenter) Register(ctl *control.Control) (AppPlugin, error) {
	a.AddTask(ctl)
	defer a.Unlock()
	a.Lock()
	if a.apps == nil {
		a.apps = make(map[string]AppPlugin)
	}

	if a.hub != nil {
		a.hub.Update(a.hub.Analysis(ctl))
	}

	//if _, ok := a.apps[ctl.Name]; ok {
	//    return nil, fmt.Errorf("app [%s] has been registered before", ctl.Name)
	//}
	app := &appPlugin{}
	if err := app.build(ctl); err != nil {
		return nil, err
	}
	a.apps[ctl.Name] = app

	return app, nil
}

type appPlugin struct {
	name    string
	urlroot string
	urls    []string
	target  string
	lfiles  []string
	runner  appRunner

	envs      []string
	changetag string
}

func (a *appPlugin) build(ctl *control.Control) error {
	a.name = ctl.Name
	a.target = ctl.Input.View.Target
	a.urlroot = ctl.IndexPageUrl
	a.urls = ctl.PageUrls
	a.lfiles = ctl.GetAssociateFiles()
	a.envs = ctl.GetEnvs()
	a.Changed()

	return a.runner.build(ctl)
}

func (a *appPlugin) GetControlFile() string {
	for _, f := range a.lfiles {
		if strings.HasSuffix(strings.TrimRight(f, " "), "control.yaml") {
			return f
		}
	}

	return ""
}

func (a *appPlugin) GetEnvs() []string {
	return a.envs
}

func (a *appPlugin) Changed() bool {
	var bts bytes.Buffer
	for _, f := range a.lfiles {
		if in, err := os.Stat(f); err == nil {
			bts.WriteString(in.ModTime().Format(time.UnixDate))
		}
	}

	if a.changetag == bts.String() {
		return false
	}

	a.changetag = bts.String()

	return true
}

func (a *appPlugin) GetAppName() string {
	return a.name
}
func (a *appPlugin) GetUrlRoot() string {
	return a.urlroot
}
func (a *appPlugin) GetAccessableUrls() []string {
	return a.urls
}

func (a *appPlugin) GetHttpHandler() (string, http.HandlerFunc, error) {

	HandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		a.runner.Run(w, r)
	}

	return a.target, HandlerFunc, nil
}
