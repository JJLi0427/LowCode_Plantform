package core

import (
	"fmt"
	"lowcode/core/appplugin"
	"lowcode/core/builtin"
	"lowcode/core/control"
	"lowcode/core/cvuecompiler"
	"lowcode/core/httpproxy"
	"lowcode/core/sitemap"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type Httpd struct {
	TlsPort             string
	Port                string
	AppsControlRootPath string
	SiteDomainName      string
	CACert              bool
	HttpRedirect        bool

	DebugAppname string

	appbuilder     *HTMLBuilder
	appcenter      *appplugin.AppRegisterCenter
	siteMapBuilder *sitemap.SiteMapBuilder
	tlsisready     bool
}

func (h *Httpd) Init() error {
	h.tlsisready = false
	h.appcenter = appplugin.NewAppRegisterCenter()
	h.appbuilder = &HTMLBuilder{DebugApp: h.DebugAppname}
	if err := h.appbuilder.Init(); err != nil {
		return err
	}

	if len(h.Port) == 0 && len(h.TlsPort) == 0 {
		h.Port = ":8081"
	}
	if len(h.Port) > 0 && h.Port[0] != ':' {
		h.Port = fmt.Sprintf(":%s", h.Port)
	}
	if len(h.TlsPort) > 0 && h.TlsPort[0] != ':' {
		h.TlsPort = fmt.Sprintf(":%s", h.TlsPort)
	}

	if len(h.SiteDomainName) == 0 {
		h.SiteDomainName = "localhost"
	}
	SiteBaseUrl := ""
	if len(h.TlsPort) > 0 {
		SiteBaseUrl = fmt.Sprintf("https://%s", h.SiteDomainName)
		if !strings.Contains(h.TlsPort, ":443") {
			SiteBaseUrl += h.TlsPort
		}
	} else {
		SiteBaseUrl = fmt.Sprintf("http://%s", h.SiteDomainName)
		if !strings.Contains(h.Port, ":80") {
			SiteBaseUrl += h.Port
		}
	}

	h.siteMapBuilder = sitemap.GetDefaultSitemapBuilder()
	h.siteMapBuilder.Init(SiteBaseUrl, path.Join(control.RootViewPath, "sitemap"), 0, 0)

	root := http.FileServer(http.Dir(control.RootViewPath))
	apps := http.FileServer(http.Dir(control.AppRootPath))
	uroot := fmt.Sprintf("/%s/", control.RootViewPath)
	uapps := fmt.Sprintf("/%s/", control.AppRootPath)

	http.Handle("/", stripPrefix(uroot, root))
	http.Handle(uapps, stripPrefix(uapps, apps))

	//interal config

	go h.buildNewApps()

	go h.rebuildApps()

	return nil
}

func (h *Httpd) ListenAndServe() error {
	if h.appcenter != nil {
		defer h.appcenter.Stop()
	}

	fmt.Printf("http listenning on [%s]\n", h.Port)
	return http.ListenAndServe(h.Port, nil)
}

func (h *Httpd) rebuildApps() {
	<-time.After(10 * time.Second)
	if len(h.appbuilder.DebugApp) > 0 {
		for {
			if app, ok := h.appcenter.Find(h.appbuilder.DebugApp); ok && app != nil {
				if app.Changed() {
					if ct, err := h.appbuilder.ParseControlFile(app.GetControlFile()); err == nil {
						if ct, err := h.appbuilder.Generate(ct); err == nil {
							if _, err := h.appcenter.Register(ct); err != nil {
								fmt.Println(err)
							}
						} else {
							fmt.Println(err)
						}
					} else {
						fmt.Println(err)
					}
				}
			} else {
				fmt.Printf("waitting debugapp[%s] ready\n", h.DebugAppname)
			}

			<-time.After(time.Second)
		}
	}

	//TODO: normal app change

}

func (h *Httpd) buildNewApps() {
	builtin.RegisterApps()

	var once sync.Once
	flags := make(map[string]string)
	for {
		h.build(flags)
		once.Do(func() { h.appcenter.Start() })
		<-time.After(5 * time.Second)
	}
}

func (h *Httpd) build(cfiletags map[string]string) {
	nctls, _ := cvuecompiler.ReadFilesInDir(h.AppsControlRootPath, true, func(f os.FileInfo, d string) bool {
		if ff := strings.ToLower(f.Name()); ff == "control.yaml" {
			if _, ok := cfiletags[d+ff]; ok {
				return false
			}
			cfiletags[d+ff] = "1"
			return true
		}

		return false
	})

	for _, ctl := range nctls {
		if ct, err := h.appbuilder.ParseControlFile(ctl); err == nil {

			if h.appbuilder.HasDebugApp() && !h.appbuilder.IsDebugApp(ct) {
				continue
			}

			if h.appbuilder.HasDebugApp() {
				fmt.Printf("debuging app[%s] at [%s]\n", ct.Name, ctl)
			} else {
				fmt.Printf("building app[%s] at [%s]\n", ct.Name, ctl)
			}

			if ct, err := h.appbuilder.Generate(ct); err == nil {
				if app, err := h.appcenter.Register(ct); err == nil {
					for _, u := range app.GetAccessableUrls() {
						h.siteMapBuilder.AddSiteMapofUrl(u)
					}
					if url, handler, err := app.GetHttpHandler(); err == nil {
						if len(url) > 0 {
							http.Handle(url, handler)
						}
					} else {
						fmt.Println(ctl, err)
					}
				} else {
					fmt.Println(ctl, err)
				}
			} else {
				fmt.Println(ctl, err)
			}
		} else {
			fmt.Println(ctl, err)
		}
	}
}

func stripPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	var homepage []byte

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		//w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")
		if r.URL.Path == "/" {
			if len(homepage) == 0 {
				homepage, _ = os.ReadFile(path.Join(control.AppRootPath, control.RootAppName, "index.html"))
			}
			if len(homepage) > 0 {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(homepage)
			} else {
				http.NotFound(w, r)
			}
		} else if httpproxy.IsThirdParyHttpRedirectProxyURLRequest(r.URL) {
			if prxy, ret := httpproxy.ParserThirdParyHttpRedirectProxyURL(r.URL); ret {
				httpproxy.HttpProxy(w, r, prxy)
			} else {
				http.NotFound(w, r)
			}
		} else if p := strings.TrimLeft(strings.TrimPrefix(r.URL.Path, prefix), "/"); len(p) <= len(r.URL.Path) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	})
}
