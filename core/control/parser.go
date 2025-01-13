package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"onlinetools/core/common/tmpl"
	"onlinetools/core/htmlrender"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"gopkg.in/yaml.v2"
)

const (
	RootViewPath string = "views"
	AppRootPath  string = "gapplications"
	PreToolsPath string = "pretools"
	RootAppName  string = "_root"
)

type AppService interface {
	Run(args []string, envs []string) (chan []byte, error)
}

func oneofvalues(src string, supports ...string) bool {
	for _, s := range supports {
		if src == s {
			return true
		}
	}
	return false
}

type OgMeta struct {
	Property string `json:"property"`
	Content  string `json:"content"`
}
type Equiv struct {
	Equiv   string `json:"equiv"`
	Content string `json:"content"`
}

type TailControl struct {
	Links   []string `json:"links,omitempty"`
	Scripts []string `json:"scripts,omitempty"`
}

//func (t *TailControl) Include(tag string) bool {
//}

type Resources struct {
	Markdown htmlrender.UiResource `json:"markdown"`
	Html     htmlrender.UiResource `json:"html"`
}

type HeadControl struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Links       []string `json:"links,omitempty"`
	Scripts     []string `json:"scripts,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Jsonldfiles []string `json:"jsonldfiles,omitempty"`
	Ogs         []OgMeta `json:"ogs,omitempty"`
	Metas       []string `json:"metas,omitempty"`
	Equivs      []Equiv  `json:"equiv,omitempty"`

	TailScripts []string
	TailLinks   []string
}

func (c *HeadControl) validate() error {

	return nil
}

type ViewControl struct {
	Target        string `json:"target"`
	Type          string `json:"type"`
	Filename      string `json:"filename,omitempty"`
	Inline_string string `json:"inline_string,omitempty"`
}

func (c *ViewControl) validate() error {

	if len(c.Type) == 0 && len(c.Filename) == 0 && len(c.Inline_string) == 0 {
		return nil
	}

	if !oneofvalues(c.Type, "sform", "vue", "html") {
		return fmt.Errorf("value of type[%s] in view not supported", c.Type)
	}

	if len(c.Filename) == 0 && len(c.Inline_string) == 0 {
		return fmt.Errorf("should specify one feild of filename or inline_string in view")
	}

	return nil
}

type ParamMaps struct {
	Viewparam string `json:"viewparam"`
	Exeparam  string `json:"exeparam"`
	Exeopt    string `json:"exeopt"`
	Paramtype string `json:"paramtype"`
	Default   string `json:"default"`
	Required  bool   `json:"required"`
	Hintftype string
}

func (p *ParamMaps) validate() error {
	if len(p.Paramtype) > 0 {
		if !oneofvalues(p.Paramtype, "txt", "number", "file") {
			if strings.HasPrefix(p.Paramtype, "file.") {
				p.Hintftype = strings.TrimPrefix(p.Paramtype, "file")
				p.Paramtype = "file"
			} else {
				return fmt.Errorf("value of paramtype[%s] in (stdin or stdout) is not supported", p.Paramtype)
			}
		}
	}

	if len(p.Viewparam) == 0 {
		p.Viewparam = p.Exeparam
	}
	if len(p.Exeparam) == 0 {
		p.Exeparam = p.Viewparam
	}

	return nil
}

type IOControl struct {
	Type   string      `json:"type"`
	Stdin  []ParamMaps `json:"stdin,omitempty"`
	Stdout []ParamMaps `json:"stdout,omitempty"`
	View   ViewControl `json:"view"`
}

func (c *IOControl) validate(hasexe bool) error {

	for i, e := 0, len(c.Stdout); i < e; i++ {
		if err := c.Stdout[i].validate(); err != nil {
			return err
		}
	}

	for i, e := 0, len(c.Stdin); i < e; i++ {
		if err := c.Stdin[i].validate(); err != nil {
			return err
		}
	}

	if hasexe && !oneofvalues(c.Type, "html", "form", "txt", "json", "xml", "pdf", "mp4", "m3u8", "png", "jpg", "gif", "link", "any", "none", "kvstr", "kvstr,json", "json,kvstr") {
		return fmt.Errorf("value of type[%s] in [input or output] not supported, or must be setted corrently if using [entrypoint] in control.yaml", c.Type)
	}

	return c.View.validate()
}

type ExeEntrypoint struct {
	exeRunner     AppService
	Add           string   `json:"add"`
	Copy          string   `json:"copy"`
	Cmd           string   `json:"cmd"`
	Inline_shell  string   `json:"inline_shell"`
	Workdir       string   `json:"workdir"`
	Args          []string `json:"args"`
	Envs          []string `json:"envs"`
	Period        string   `json:"period"`
	Trace         bool     `json:"trace"`
	Packdepend    bool     `json:"packdepend"`
	Paramsinview  []string `json:"paramsinview"`
	Resultsinview []string `json:"resultsinview"`
}

func (e *ExeEntrypoint) GetAppService() AppService {
	return e.exeRunner
}

func (e *ExeEntrypoint) SetAppService(exe AppService) {
	e.exeRunner = exe
}

func (e *ExeEntrypoint) HasExeEntrypoint() bool {
	return len(e.Cmd) != 0 || len(e.Inline_shell) != 0 || e.exeRunner != nil
}

func (e *ExeEntrypoint) HasViewParams() bool {
	return len(e.Paramsinview) != 0
}

func (e *ExeEntrypoint) HasViewResults() bool {
	return len(e.Resultsinview) != 0
}

func (e *ExeEntrypoint) BuildInputControl() *IOControl {
	//TODO: parse Paramsinview and build IOControl

	return nil
}

type Control struct {
	ControlFilePath string
	Name            string          `json:"name"`
	Head            HeadControl     `json:"head"`
	Tail            TailControl     `json:"tail"`
	Entrypoint      ExeEntrypoint   `json:"entrypoint"`
	Backtasks       []ExeEntrypoint `json:"backtasks"`
	Resource        Resources       `json:"resource"`
	Input           IOControl       `json:"input"`
	Output          IOControl       `json:"output"`
	Layout          string          `json:"layout"`

	IndexPageUrl string
	PageUrls     []string

	//copyfiles []string
	tjs      []byte
	tcss     []byte
	bjs      []byte
	bcss     []byte
	lfiles   []string
	envcache []string
}

func (c *Control) GetEnvs() []string {
	if len(c.envcache) > 0 {
		return c.envcache
	}

	envs := []string{}
	if path, err := filepath.Abs(c.GetAppIndexPageHomePath()); err == nil {
		envs = append(envs,
			[]string{fmt.Sprintf("%s=%s", "indexpagepath", path),
				fmt.Sprintf("%s=%s", "IndexPagePath", path),
				fmt.Sprintf("%s=%s", "indexPagePath", path),
				fmt.Sprintf("%s=%s", "AppPagePath", path),
				fmt.Sprintf("%s=%s", "appPagePath", path),
				fmt.Sprintf("%s=%s", "apppagepath", path)}...)
	}

	envs = append(envs, fmt.Sprintf("%s=%s", "IndexPageUrl", c.IndexPageUrl))
	envs = append(envs, fmt.Sprintf("%s=%s", "indexPageUrl", c.IndexPageUrl))
	envs = append(envs, fmt.Sprintf("%s=%s", "indexpageurl", c.IndexPageUrl))
	envs = append(envs, fmt.Sprintf("%s=%s", "AppPageUrl", c.IndexPageUrl))
	envs = append(envs, fmt.Sprintf("%s=%s", "appPageUrl", c.IndexPageUrl))
	envs = append(envs, fmt.Sprintf("%s=%s", "apppageurl", c.IndexPageUrl))
	envs = append(envs, fmt.Sprintf("%s=%s", "AppName", c.Name))
	envs = append(envs, fmt.Sprintf("%s=%s", "appName", c.Name))
	envs = append(envs, fmt.Sprintf("%s=%s", "appname", c.Name))

	if apath, err := filepath.Abs(c.ControlFilePath); err == nil {
		envs = append(envs, fmt.Sprintf("%s=%s", "AppControlPath", apath))
		envs = append(envs, fmt.Sprintf("%s=%s", "appControlPath", apath))
		envs = append(envs, fmt.Sprintf("%s=%s", "appcontrolpath", apath))
		//envs = append(envs, fmt.Sprintf("%s=%s", "LD_LIBRARY_PATH", fmt.Sprintf("%s:%s", path.Join(apath, "lib"), os.Getenv("LD_LIBRARY_PATH"))))
		//envs = append(envs, fmt.Sprintf("%s=%s", "PATH", fmt.Sprintf("%s:%s", path.Join(apath, "bin"), os.Getenv("PATH"))))
	} else {
		fmt.Println("Error: add [ControlFilePath] env variable failed: ", err.Error())
	}

	c.envcache = envs

	return envs
}

func (c *Control) GetAssociateFiles() []string {
	return c.lfiles
}

func (c *Control) AppendAssociateFiles(file string) bool {
	for _, f := range c.lfiles {
		if f == file {
			return false
		}
	}
	c.lfiles = append(c.lfiles, file)
	return true
}

func (c *Control) Validate() error {
	if e := c.Head.validate(); e != nil {
		return e
	}
	//if e := c.Tail.validate(); e != nil {
	//	return e
	//}
	if e := c.Input.validate(c.Entrypoint.HasExeEntrypoint()); e != nil {
		return e
	}
	if e := c.Output.validate(c.Entrypoint.HasExeEntrypoint()); e != nil {
		return e
	}

	if len(c.Input.View.Target) == 0 {
		c.Input.View.Target = fmt.Sprintf("/%s/api", c.GetAppIndexPageHomePath())
	}

	if len(c.Input.View.Target) > 0 {
		if strings.Contains(strings.Join(c.Head.Scripts, ","), "core_iobind.js") {
			var bts bytes.Buffer
			bts.WriteString(fmt.Sprintf(";AsyncPostDataRegisterUrl('%s');", c.Input.View.Target))
			c.tjs = append(c.tjs, bts.Bytes()...)
		}
		if strings.Contains(strings.Join(c.Tail.Scripts, ","), "core_iobind.js") {
			var bts bytes.Buffer
			bts.WriteString(fmt.Sprintf(";AsyncPostDataRegisterUrl('%s');", c.Input.View.Target))
			c.tjs = append(c.bjs, bts.Bytes()...)
		}
	}

	if len(c.Name) == 0 {
		return fmt.Errorf("can not find [name] in %s/%s", c.ControlFilePath, "control.yaml")
	}

	return nil
}

func (c *Control) Parse(controlFile string) error {
	if yamlFile, err := ioutil.ReadFile(controlFile); err == nil {
		err = yaml.Unmarshal(yamlFile, c)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	c.AppendAssociateFiles(controlFile)

	if idx := strings.LastIndex(controlFile, "/"); idx != -1 {
		c.ControlFilePath = controlFile[:idx+1]
	}

	if len(c.Head.Jsonldfiles) > 0 {
		for i, e := range c.Head.Jsonldfiles {
			if len(e) == 0 {
				return fmt.Errorf("empty jsonldfiles element")
			}
			if e[0] != '~' && e[0] != '/' {
				e = c.ControlFilePath + e
				c.AppendAssociateFiles(e)
			}
			if bt, err := ioutil.ReadFile(e); err == nil {
				c.Head.Jsonldfiles[i] = string(bt)
			} else {
				return fmt.Errorf("read jsonld file %s, %s", e, err.Error())
			}
		}
	}


	//build in app
	for _, builtin := range GetBuiltinApplications() {
		if err := builtin(c); err != nil {
			return err
		}
	}

	if len(c.Input.View.Filename) > 0 {
		e := c.Input.View.Filename
		if e[0] != '~' && e[0] != '/' {
			e = c.ControlFilePath + e
			c.AppendAssociateFiles(e)
		}
		if bt, err := ioutil.ReadFile(e); err == nil {
			c.Input.View.Inline_string = string(bt) + c.Input.View.Inline_string
		} else {
			e := filepath.Join(RootViewPath, c.Input.View.Filename)
			if bt, err := ioutil.ReadFile(e); err == nil {
				c.AppendAssociateFiles(e)
				c.Input.View.Inline_string = string(bt) + c.Input.View.Inline_string
			} else {
				return err
			}
		}
	}

	if len(c.Output.View.Filename) > 0 {
		e := c.Output.View.Filename
		if e[0] != '~' && e[0] != '/' {
			e = c.ControlFilePath + e
			c.AppendAssociateFiles(e)
		}
		if bt, err := ioutil.ReadFile(e); err == nil {
			c.Output.View.Inline_string = string(bt) + c.Output.View.Inline_string
		} else {
			e := filepath.Join(RootViewPath, c.Output.View.Filename)
			if bt, err := ioutil.ReadFile(e); err == nil {
				c.AppendAssociateFiles(e)
				c.Output.View.Inline_string = string(bt) + c.Output.View.Inline_string
			} else {
				return err
			}
		}
	}

	if c.Entrypoint.HasViewParams() {
		entryviewparser := &ViewParamBuilder{Ctrl: c}
		if input, err := entryviewparser.BuildIOControl(c.Entrypoint.Paramsinview, "/"+c.GetAppIndexPageHomePath()+"/api", "POST"); err == nil {
			c.Head.Scripts = append(c.Head.Scripts, "/assets/js/base/core_iobind.js")
			c.SaveJS(fmt.Sprintf("sformOnSubmitConfig('%s', '%s');", "#entrypointform", "/"+c.GetAppIndexPageHomePath()+"/api"), false)
			c.Input = *input
		} else {
			return err
		}
		if c.Entrypoint.HasViewResults() {
			entryviewparser := &ViewParamBuilder{Ctrl: c}
			if output, script, err := entryviewparser.BuildOuputIOControl(c.Entrypoint.Resultsinview); err == nil {
				c.SaveJS(script, false)
				c.Output = *output
			} else {
				return err
			}
		}
	}

	return nil
}

func (c *Control) ToJson(indent bool) (string, error) {
	var bt []byte
	var err error
	if !indent {
		if bt, err = json.Marshal(c); err != nil {
			return "", err
		}
	} else {
		if bt, err = json.MarshalIndent(c, " ", " "); err != nil {
			return "", err
		}
	}
	return string(bt), nil
}

// save content to file, and add http target link to Head.TailLinks or Head.Links
func (c *Control) SaveCSS(content string, totop bool) {
	if totop {
		c.tcss = append(c.tcss, []byte(content)...)
		c.tcss = append(c.tcss, '\n')
	} else {
		c.bcss = append(c.bcss, []byte(content)...)
		c.bcss = append(c.bcss, '\n')
	}
}

// save content to file, return http target link
func (c *Control) SaveJS(content string, totop bool) {
	if totop {
		c.tjs = append(c.tjs, []byte(content)...)
		c.tjs = append(c.tjs, '\n')
	} else {
		c.bjs = append(c.bjs, []byte(content)...)
		c.bjs = append(c.bjs, '\n')
	}
}

func (c *Control) GetAppIndexPageHomePath() string {
	return path.Join(AppRootPath, c.Name)
}

//func (c *Control) GetAppIndexPageUrl() string {
//	return "/" + path.Join(AppRootPath, c.Name)
//}

func (c *Control) SaveHtml(filename string) (*os.File, error) {
	p := c.GetAppIndexPageHomePath()
	if err := os.MkdirAll(p, 0750); err != nil {
		return nil, err
	}
	ff := fmt.Sprintf("%s/%s", p, filename)
	c.IndexPageUrl = "/" + p
	c.PageUrls = append(c.PageUrls, "/"+ff)
	return os.OpenFile(ff, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
}

func splitfiles(files []string) ([]string, []string) {
	var cfiles []string
	var gfiles []string
	for _, f := range files {
		if len(f) == 0 {
			continue
		}
		if f[0] == '/' || strings.HasPrefix(f, "http") {
			gfiles = append(gfiles, f)
		} else {
			cfiles = append(cfiles, f)
		}
	}

	return gfiles, cfiles
}

func (c *Control) splitfilesbyext(cc []string, ext string) ([]string, []string) {
	var c1 []string
	var c2 []string
	for _, f := range cc {
		if filepath.Ext(f) == ext {
			c1 = append(c1, f)
		} else {
			c2 = append(c2, f)
		}
	}

	return c1, c2
}

func (c *Control) loadLocalLinks(links []string) {
	for _, link := range links {
		if f, cnt := tmpl.Attributes(link, "href", true); cnt == 1 && len(f) > 0 {
			if f[0] != '/' && !strings.HasPrefix(f, "http") {
				c.localrelativeurl(f)
			}
		}
	}
}

func (c *Control) loadLocalScripts(scripts []string) {
	for _, link := range scripts {
		if f, cnt := tmpl.Attributes(link, "src", true); cnt == 1 && len(f) > 0 {
			if f[0] != '/' && !strings.HasPrefix(f, "http") {
				c.localrelativeurl(f)
			}
		}
	}
}

func (c *Control) loadLocalLinkScripts() {
	c.loadLocalLinks(c.Head.Links)
	c.loadLocalLinks(c.Tail.Links)
	c.loadLocalScripts(c.Head.Scripts)
	c.loadLocalLinks(c.Tail.Scripts)
}

func (c *Control) formatJS(content string) []byte {

	result := api.Transform(content, api.TransformOptions{
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	})

	if len(result.Errors) > 0 {
		return []byte(content)
	}
	return result.Code
}

func (c *Control) localrelativeurl(file string) (string, error) {
	p := path.Join(AppRootPath, c.Name, filepath.Dir(file))
	if err := os.MkdirAll(p, 0750); err != nil {
		return "", err
	}

	f := path.Join(AppRootPath, c.Name) + "/" + file
	if bts, err := ioutil.ReadFile(path.Join(c.ControlFilePath, file)); err == nil {
		c.AppendAssociateFiles(path.Join(c.ControlFilePath, file))

		if ext := filepath.Ext(file); ext == ".js" || ext == ".ts" {
			bts = c.formatJS(string(bts))
		}

		ioutil.WriteFile(f, bts, 0640)
	} else {
		return "", err
	}

	return "/" + f, nil
}

/*
func (c *Control) loadlocalCssJsold() {
	if gg, cc := splitfiles(c.Head.Links); len(cc) > 0 {
		c.Head.Links = gg
		cc, ss := c.splitfilesbyext(cc, ".css")
		if len(ss) > 0 {
			for _, s := range ss {
				if url, err := c.localrelativeurl(s); err == nil {
					c.Head.Links = append(c.Head.Links, url)
				}
			}
		}
		c.SaveCSS(c.loadlocalfiles(cc), true)

	}

	if gg, cc := splitfiles(c.Head.Scripts); len(cc) > 0 {
		c.Head.Scripts = gg
		cc, ss := c.splitfilesbyext(cc, ".js")
		if len(ss) > 0 {
			for _, s := range ss {
				if url, err := c.localrelativeurl(s); err == nil {
					c.Head.Links = append(c.Head.Scripts, url)
				}
			}
		}
		c.SaveJS(c.loadlocalfiles(cc), true)
	}

	if gg, cc := splitfiles(c.Tail.Links); len(cc) > 0 {
		c.Tail.Links = gg
		cc, ss := c.splitfilesbyext(cc, ".css")
		if len(ss) > 0 {
			for _, s := range ss {
				if url, err := c.localrelativeurl(s); err == nil {
					c.Tail.Links = append(c.Tail.Links, url)
				}
			}
		}
		c.SaveCSS(c.loadlocalfiles(cc), false)
	}

	if gg, cc := splitfiles(c.Tail.Scripts); len(cc) > 0 {
		c.Tail.Scripts = gg
		cc, ss := c.splitfilesbyext(cc, ".js")
		if len(ss) > 0 {
			for _, s := range ss {
				if url, err := c.localrelativeurl(s); err == nil {
					c.Tail.Scripts = append(c.Tail.Scripts, url)
				}
			}
		}
		c.SaveJS(c.loadlocalfiles(cc), false)
	}
}

func (c *Control) loadlocalfiles(files []string) string {
	var btts bytes.Buffer
	for _, f := range files {
		lfile := path.Join(c.ControlFilePath, f)
		c.AppendAssociateFiles(lfile)
		if fd, err := os.Open(lfile); err == nil {
			io.Copy(&btts, fd)
			fd.Close()
		}
	}

	return btts.String()
}*/

func (c *Control) Save() error {
	c.loadLocalLinkScripts()
	if len(c.bjs) > 0 {
		result := api.Transform(string(c.bjs), api.TransformOptions{
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
		})
		if len(result.Errors) == 0 {
			c.bjs = result.Code
		}
	}
	if len(c.tjs) > 0 {
		result := api.Transform(string(c.tjs), api.TransformOptions{
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
		})
		if len(result.Errors) == 0 {
			c.tjs = result.Code
		}
	}

	if len(c.bcss) > 0 {
		if pf, pu, err := c.apppath("css", "below"); err == nil {
			if err := ioutil.WriteFile(pf, c.bcss, 0640); err == nil {
				c.Tail.Links = append(c.Tail.Links, pu)
			} else {
				return err
			}
		} else {
			return err
		}
	}

	if len(c.tcss) > 0 {
		if pf, pu, err := c.apppath("css", "top"); err == nil {
			if err := ioutil.WriteFile(pf, c.tcss, 0640); err == nil {
				c.Head.Links = append(c.Head.Links, pu)
			} else {
				return err
			}
		} else {
			return err
		}
	}

	if len(c.tjs) > 0 {
		if pf, pu, err := c.apppath("js", "top"); err == nil {
			if err := ioutil.WriteFile(pf, c.tjs, 0640); err == nil {
				c.Head.Scripts = append(c.Head.Scripts, pu)
			} else {
				return err
			}
		} else {
			return err
		}
	}

	if len(c.bjs) > 0 {
		if pf, pu, err := c.apppath("js", "below"); err == nil {
			if err := ioutil.WriteFile(pf, c.bjs, 0640); err == nil {
				c.Tail.Scripts = append(c.Tail.Scripts, pu)
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// file save path , http url path
func (c *Control) apppath(ftype string, tag string) (string, string, error) {
	p := path.Join(AppRootPath, c.Name, ftype)
	//pu := path.Join(c.Name, ftype)
	if err := os.MkdirAll(p, 0750); err != nil {
		return "", "", err
	}
	ff := fmt.Sprintf("/%s_%s.%s", tag, strings.ToLower(c.Name), ftype)

	return p + ff, "/" + p + ff, nil
}
