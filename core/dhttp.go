package core

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"onlinetools/core/appplugin"
	"onlinetools/core/builtin"
	"onlinetools/core/common/acme"
	"onlinetools/core/common/file"
	"onlinetools/core/control"
	"onlinetools/core/cvuecompiler"
	"onlinetools/core/httpproxy"
	"onlinetools/core/rclonedriver"
	"onlinetools/core/sitemap"
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
	rclonedriver.ConnectDefaultRclone()

	go h.buildNewApps()

	go h.rebuildApps()

	return nil
}

func (h *Httpd) needRedirect(caroot string) http.Handler {
	hasca := false
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.HttpRedirect && len(h.TlsPort) > 0 && h.tlsisready {
			if !hasca {
				if _, err := os.Stat(path.Join(caroot, "webtools.crt")); err == nil {
					hasca = true
				}
			}
			if hasca {
				http.Redirect(w, r, fmt.Sprintf("https://%s%s%s", h.SiteDomainName, h.TlsPort, r.RequestURI), http.StatusMovedPermanently)
				return
			}
		}
		http.DefaultServeMux.ServeHTTP(w, r)
	})
}

func (h *Httpd) ListenAndServe() error {
	if h.appcenter != nil {
		defer h.appcenter.Stop()
	}

	caroot := path.Join(control.PreToolsPath, "temporaryCA")
	var acmer *acme.AcmeCA
	if h.Port == ":80" && len(h.TlsPort) > 0 && h.CACert {
		caroot = "cert"
		os.MkdirAll(caroot, 0755)
		acmer = &acme.AcmeCA{AcmeshFile: path.Join(control.PreToolsPath, "acme.sh", "acme.sh"),
			WebRoot:    control.RootViewPath,
			CAPath:     caroot,
			WebDomains: []string{h.SiteDomainName}}
	}

	if len(h.TlsPort) > 0 {
		if len(h.Port) > 0 {
			fmt.Printf("http listenning on [%s]\n", h.Port)
			go http.ListenAndServe(h.Port, h.needRedirect(caroot))
		}
		if acmer != nil {
			time.Sleep(time.Second * 3)
			acmer.Init()
			h.appcenter.AddCronTask("acme.sh-renew-ca", "0 0 1 * *", acmer)
		}

		h.tlsisready = true
		crt, key := path.Join(caroot, "webtools.crt"), path.Join(caroot, "webtools.key")

		dycert := &file.DynamicCertificate{}
		if err := dycert.Init(crt, key); err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("https listenning on [%s]\n", h.TlsPort)
		server := &http.Server{Addr: h.TlsPort,
			TLSConfig: &tls.Config{
				GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
					return dycert.GetCertificate(), nil
				},
			},
		}
		return server.ListenAndServeTLS("", "")
		//return server.ListenAndServeTLS(crt, key)
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
