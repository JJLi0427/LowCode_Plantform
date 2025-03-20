package appplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"onlinetools/core/control"
	"onlinetools/core/sformcompiler"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func trimQ(con string) string {
	return strings.TrimRight(strings.TrimLeft(con, "\"'"), "\"'")
}

type appParamParser struct {
	params         []control.ParamMaps
	sourceDataType string
	outputViewType string
	execWorkDir    string
	indexPageDir   string
	runtimeDir     string
	sformbuilder   *sformcompiler.SformValueTransfomer

	envs []string
}

func (a *appParamParser) GetAppRootPathEnv() []string {
	if len(a.runtimeDir) == 0 && len(a.indexPageDir) != 0 {
		if path, err := filepath.Abs(a.indexPageDir); err == nil {
			a.runtimeDir = path
		}
	}

	if len(a.runtimeDir) > 0 {
		return append(a.envs,
			[]string{fmt.Sprintf("%s=%s", "appPageRoot", a.runtimeDir),
				fmt.Sprintf("%s=%s", "appPageHome", a.runtimeDir),
				fmt.Sprintf("%s=%s", "apppageroot", a.runtimeDir),
				fmt.Sprintf("%s=%s", "apppagehome", a.runtimeDir)}...)
	}

	return a.envs
}

func (a *appParamParser) ParseFromHttpRequest(r *http.Request) ([]*appParam, error) {
	//fmt.Println(r.Method)
	//fmt.Println(r.URL)
	if len(a.params) == 0 {
		return nil, nil
	}

	appParams := make([]*appParam, len(a.params))
	flags := make([]bool, len(a.params))

	contenttype := strings.Join(r.Header["Content-Type"], ",")
	if a.sourceDataType == "form" || strings.Contains(contenttype, "form") {
		r.ParseMultipartForm(4096)
		r.ParseForm()

		if r.Form != nil {
			for i, k := range a.params {
				if v, ok := r.Form[k.Viewparam]; ok && !flags[i] {

					appParams[i] = &appParam{}
					appParams[i].SetHintFileExt(k.Hintftype)
					appParams[i].Set(paramType(k.Paramtype), contentType(contenttype), k.Exeopt)
					appParams[i].SetContentBytes([]byte(strings.Join(v, ";;")))
					flags[i] = true
				}
			}
		}

		if r.MultipartForm != nil {
			//TODO: should call defer r.MultipartForm.RemoveAll() ?  to remove tmp file ?
			defer r.MultipartForm.RemoveAll()

			for i, k := range a.params {
				if flags[i] {
					continue
				}

				if v, ok := r.MultipartForm.Value[k.Viewparam]; ok {

					appParams[i] = &appParam{}
					appParams[i].SetHintFileExt(k.Hintftype)
					appParams[i].Set(paramType(k.Paramtype), contentType(contenttype), k.Exeopt)
					appParams[i].SetContentBytes([]byte(strings.Join(v, ";;")))
					flags[i] = true
				} else if f, ok := r.MultipartForm.File[k.Viewparam]; ok && len(f) > 0 {

					appParams[i] = &appParam{}
					appParams[i].SetHintFileExt(k.Hintftype)
					appParams[i].Set(paramType(k.Paramtype), contentTypeByMIME(f[0].Header), k.Exeopt)
					if err := appParams[i].SetContentMultipart(f); err != nil {
						return nil, err
					}
					flags[i] = true
				}
			}
		}

		if isAllTrue(flags) {
			return appParams, nil
		}
	}

	if r.Body != nil && (a.sourceDataType == "json" || strings.Contains(contenttype, "json")) {
		dec := json.NewDecoder(r.Body)
		for {
			m := make(map[string]interface{})
			if err := dec.Decode(&m); err == io.EOF {
				break
			} else if err != nil {
				//log.Fatal(err)
				break
			}

			for i, flg := range flags {
				if flg {
					continue
				}

				k := a.params[i]
				param := &appParam{}
				if param.SetContentMap(m, k.Viewparam, k.Paramtype, k.Exeopt) {
					appParams[i] = param
					//appParams[i].Set(paramType(k.Paramtype), TXT, k.Exeopt)
					flags[i] = true
				}
			}
		}
		if isAllTrue(flags) {
			return appParams, nil
		}
	}

	//TODO: what about other inputed data, except form & json

	//fill default for non-inputed element
	for i, flg := range flags {
		if !flg {
			k := a.params[i]

			if k.Required && len(k.Default) == 0 {
				return nil, fmt.Errorf("input param [%s] is required", k.Viewparam)
			}

			if len(k.Default) > 0 || len(k.Exeopt) > 0 {
				appParams[i] = &appParam{}
				appParams[i].SetHintFileExt(k.Hintftype)
				appParams[i].Set(paramType(k.Paramtype), "txt", k.Exeopt)
				appParams[i].SetContentString(k.Default)
				flags[i] = true
			}
		}
	}

	return appParams, nil
}

func (a *appParamParser) WriteHttpResponse(w http.ResponseWriter, payload []byte) {
	// payload is exe's output
	if payload == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8080")

	if len(a.params) == 0 {
		w.Write(payload)
		return
	}

	if bts, closer, err := a.BuildViewData(payload); err == nil {
		if closer != nil {
			defer closer.Close()
		}
		w.WriteHeader(http.StatusOK)
		w.Write(bts)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

}

func (a *appParamParser) BuildExeParams(params []*appParam) ([]string, io.Closer, error) {
	if len(params) == 0 {
		return nil, nil, nil
	}
	//build exe command line params  by []*appParam
	//io.Closer is for  tmp file, that is created by commandline logics, which need to be remove or close,
	//when executabale app has running done
	//params elements maybe nil
	fileremover := &fileRemover{}
	hasfile := false
	//var bts bytes.Buffer
	var sps []string
	for _, param := range params {
		if param == nil {
			continue
		}
		if ps := param.Get(); ps != nil {
			sps = append(sps, ps...)
		}
		//bts.WriteString(param.Get())
		//bts.WriteByte(' ')

		if param.paramType == FILE {
			hasfile = true
			fileremover.add(string(param.content))
		}
	}

	if hasfile {
		return sps, fileremover, nil
	}
	return sps, nil, nil
}

func (a *appParamParser) mapData(src map[string]interface{}, dest map[string]interface{}) error {
	for _, kv := range a.params {
		if len(kv.Viewparam) > 0 {
			vkeys := strings.Split(kv.Exeparam, ".")
			dest[kv.Viewparam] = strings.Join(mapdata(src, vkeys), ",")

			if kv.Exeparam == "__wholecontent__" {
				break
			}
		}
	}

	return nil
}

func (a *appParamParser) BuildViewData(payload []byte) ([]byte, io.Closer, error) {
	//TODO: build final view data by exeoutput as payload, which will be setted to http response
	//io.Closer,  tmp created by view logics, which need to be remove or close, when calling Close(),
	//sourceDataType : html | txt | json | xml | pdf | mp4 | m3u8 | png | jpg | gif | form | link | kvstr
	//outputViewType : sform | vue | html
	//w.Write(payload)

	if len(a.runtimeDir) == 0 && len(a.indexPageDir) != 0 {
		if path, err := filepath.Abs(a.indexPageDir); err == nil {
			a.runtimeDir = path
		}
	}

	//must use string value for sform
	viewparams := make(map[string]interface{})

	if strings.Contains(a.sourceDataType, "json") && len(a.params) > 0 {
		payloadmap := make(map[string]interface{})
		if err := json.Unmarshal(payload, &payloadmap); err != nil {
			if strings.Contains(a.sourceDataType, "kvstr") {
				a.sourceDataType = "kvstr"
			} else {
				return nil, nil, fmt.Errorf("json parse exe's output-payload error: %s", err.Error())
			}
		} else {
			a.mapData(payloadmap, viewparams)
			if a.outputViewType == "sform" {
				for key, v := range viewparams {
					if _, ok := v.(map[string]interface{}); ok {
						return nil, nil, fmt.Errorf("json param map for key[%s] in exe's output-payload", key)
					}

					if val, ok := v.([]interface{}); ok {
						viewparams[key] = fmt.Sprintln(val)
					}

					if val, ok := v.(float64); ok {
						viewparams[key] = fmt.Sprintln(val)
					}
				}
			}
		}
	}

	if a.sourceDataType == "kvstr" && len(a.params) > 0 {
		if len(payload) > 0 {
			payload = append(payload, ' ')
		}
		for _, param := range a.params {
			if idx := strings.Index(string(payload), param.Exeparam); idx >= 0 {
				var btts bytes.Buffer
				for _, c := range payload[idx+len(param.Exeparam):] {
					if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ':' {
						if btts.Len() > 0 {
							if param.Paramtype == "file" {
								rpath := path.Join(a.runtimeDir, "runtime")
								if fn, err := os.Stat(rpath); err != nil || !fn.IsDir() {
									os.MkdirAll(rpath, 0750)
								}
								_, file := filepath.Split(trimQ(btts.String()))
								exe := exec.Command("mv", trimQ(btts.String()), rpath+"/")
								exe.Dir = a.execWorkDir
								exe.Output()
								file = url.PathEscape(file)
								viewparams[param.Viewparam] = "/" + path.Join(a.indexPageDir, "runtime", file)
							} else {
								viewparams[param.Viewparam] = trimQ(btts.String())
							}
							break
						}
					} else {
						btts.WriteByte(c)
					}
				}
			} else if len(param.Default) > 0 {
				viewparams[param.Viewparam] = param.Default
			} else if _, ok := viewparams[param.Viewparam]; !ok && param.Exeparam == "__wholecontent__" {
				viewparams[param.Viewparam] = string(payload)
			}
		}
	}

	if len(viewparams) > 0 {
		if a.outputViewType == "sform" {
			var params [][2]string
			for k, v := range viewparams {
				if val, ok := v.(string); ok {
					params = append(params, [2]string{k, val})
				}
			}
			if sformdata, err := a.sformbuilder.WriteKVs(params); err == nil {
				return []byte(sformdata), nil, nil
			} else {
				return nil, nil, err
			}
		} else {
			if bts, err := json.Marshal(viewparams); err == nil {
				return bts, nil, nil
			} else {
				return nil, nil, err
			}
		}
	}

	return payload, nil, nil
}

type appParam struct {
	paramType   ParamType
	contentType ContentType
	hintFileExt string
	content     []byte
	exeopt      string
}

func (a *appParam) Set(ptyp ParamType, ctyp ContentType, exeopt string) {
	a.paramType = ptyp
	a.contentType = ctyp
	a.exeopt = exeopt
}

func (a *appParam) SetHintFileExt(ext string) {
	a.hintFileExt = ext
}

func (a *appParam) SetContentString(str string) {
	a.SetContentBytes([]byte(str))
}

func (a *appParam) SetContentBytes(con []byte) {
	if a.paramType != FILE {
		a.content = con
	} else {
		fmode := "onlinetool.*" + contentFileExt(a.contentType)
		if len(a.hintFileExt) > 0 {
			fmode = "onlinetool.*" + a.hintFileExt
		}
		if file, err := os.CreateTemp(os.TempDir(), fmode); err == nil {
			defer file.Close()
			file.Write(con)
			a.content = []byte(file.Name())
		} else {
			fmt.Println(err)
		}
	}
}

func (a *appParam) SetContentMultipart(f []*multipart.FileHeader) error {
	if a.paramType == FILE && len(f) > 0 {
		if srcfile, err := f[0].Open(); err == nil {
			defer srcfile.Close()
			//os.Stat(filepath.Join(os.TempDir(), srcfile.Name()))

			fmode := "onlinetool.*" + contentFileExt(a.contentType)
			if len(a.hintFileExt) > 0 {
				fmode = "onlinetool.*" + a.hintFileExt
			}
			if file, err := os.CreateTemp(os.TempDir(), fmode); err == nil {
				defer file.Close()
				io.Copy(file, srcfile)
				a.content = []byte(file.Name())
			} else {
				return err
			}
		} else {
			return err
		}
	} else if len(f) > 0 {
		var bts bytes.Buffer
		if srcfile, err := f[0].Open(); err == nil {
			defer srcfile.Close()
			if _, err := io.Copy(&bts, srcfile); err == nil {
				a.content = bts.Bytes()
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func mapdata(src map[string]interface{}, vkeys []string) []string {
	amap := src
	for i, k := range vkeys {
		if t, ok := amap[k].(map[string]interface{}); ok {
			amap = t
		} else if arr, ok := amap[k].([]map[string]interface{}); ok {
			var datas []string
			for _, smap := range arr {
				if vs := mapdata(smap, vkeys[i+1:]); vs != nil {
					datas = append(datas, vs...)
				}
			}
			return datas
		} else if i+1 == len(vkeys) {
			if s, ok := amap[k].(string); ok {
				return []string{s}
			} else if f, ok := amap[k].(float64); ok {
				return []string{strconv.FormatFloat(f, 'f', -1, 64)}
			} else if ar, ok := amap[k].([]string); ok {
				return ar
			} else if ar, ok := amap[k].([]float64); ok {
				var ss []string
				for _, ff := range ar {
					ss = append(ss, strconv.FormatFloat(ff, 'f', -1, 64))
				}
				return ss
			}
		} else {
			return nil
		}
	}

	return nil
}

func (a *appParam) SetContentMap(src map[string]interface{}, keys string, ptype, opt string) bool {
	//keys: bdd.jdd.jdd
	//keys: bdd.arrays.jdd.ddd
	vkeys := strings.Split(keys, ".")
	if arr := mapdata(src, vkeys); arr != nil {
		a.Set(paramType(ptype), "txt", opt)
		a.SetContentString(strings.Join(arr, "`"))

		return true
	}

	return false
}

func (a *appParam) Get() []string {
	if len(a.content) == 0 {
		return nil
	}
	var btts []string
	if len(a.content) > 0 {
		if len(a.exeopt) > 0 {
			btts = append(btts, a.exeopt)
		}
		btts = append(btts, string(a.content))
	}
	return btts
}

func isAllTrue(flgs []bool) bool {
	ret := true
	for _, fg := range flgs {
		if !fg {
			ret = false
			break
		}
	}
	return ret
}

type fileRemover struct {
	files []string
}

func (f *fileRemover) add(file string) {
	f.files = append(f.files, file)
}
func (f *fileRemover) Close() error {
	for _, fhs := range f.files {
		os.Remove(fhs)
	}
	return nil
}
