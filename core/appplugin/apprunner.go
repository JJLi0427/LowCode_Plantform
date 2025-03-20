package appplugin

import (
	"bytes"
	"fmt"
	"net/http"
	"onlinetools/core/common/file"
	"onlinetools/core/control"
	"onlinetools/core/sformcompiler"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type exeEntry struct {
	cmdenvs []string
	cmdtype string
	cmdroot string
	runner  control.AppService
	cmd     string
	ishell  string
	add     string
	copy    string
	trace   bool

	period  string
	args    []string
	appctrl string

	envcache []string
}

func (e *exeEntry) Run() {
	if r, err := e.run([]string{}); err != nil {
		fmt.Println(string(r), err)
	}
}

func (e *exeEntry) run(params []string, envargs ...string) ([]byte, error) {
	if len(e.args) > 0 {
		params = append(e.args, params...)
	}
	var exe *exec.Cmd
	if len(e.cmd) > 0 {
		if strings.HasSuffix(e.cmd, ".sh") {
			exe = exec.Command("/bin/bash", append([]string{e.cmd}, params...)...)
		} else {
			exe = exec.Command(e.cmd, params...)
		}
	} else if len(e.ishell) > 0 {
		exe = exec.Command("/bin/bash", append([]string{"-c", e.ishell, "inline_shell"}, params...)...)
	} else if e.runner != nil {
		var bts bytes.Buffer
		if ch, err := e.runner.Run(params, nil); err == nil {
			for bt :=range ch {
				bts.Write(bt)
			}
			fmt.Printf("%s[runcmd %s]: %v\n", e.appctrl, "insvr", append([]string{"insvr"}, params...))
			return bts.Bytes(), nil
		} else {
			fmt.Printf("%s[runcmd %s]: %v [%s]\n", e.appctrl, "insvr", append([]string{"insvr"}, params...), err.Error())
			return nil, err
		}
	}else{
		return nil, fmt.Errorf("empty command runner")
	}
	//cmd.Args = params
	exe.Dir = e.cmdroot
	//cmd.Path =
	exe.Env = e.addRuntimeEnv(append(e.cmdenvs, exe.Environ()...))
	exe.Env = append(exe.Env, envargs...)
	//fmt.Println("env: ", exe.Env)

	//buf, err := cmd.CombinedOutput()

	var buf []byte
	var err error
	if e.trace {
		buf, err = exe.CombinedOutput()
	} else {
		buf, err = exe.Output()
	}

	if len(buf) > 0 && err != nil && strings.HasPrefix(err.Error(), "exit status 1") {
		err = nil
	}

	if len(e.cmd) != 0 {
		fmt.Printf("%s[runcmd %s]: %v\n", e.appctrl, e.cmdtype, append([]string{e.cmd}, params...))
		//fmt.Println(e.appctrl, "[runcmd]: ", append([]string{e.cmd}, params...))
	} else if len(e.ishell) != 0 {
		fmt.Printf("%s[runcmd %s]: %v\n", e.appctrl, e.cmdtype, append([]string{"${inline_shell}"}, params...))
		//fmt.Println(e.appctrl, "[runcmd]: ", append([]string{"${inline_shell}"}, params...))
	}

	if e.trace {
		fmt.Printf("%s[cmdoutput %s]: %s\n", e.appctrl, e.cmdtype, string(buf))
		//fmt.Println(e.appctrl, "[cmdoutput]: ", string(buf))
	}

	return buf, err
}

func (e *exeEntry) CopyDependence() {
	if libpath, err := filepath.Abs(e.cmdroot + "/lib"); err == nil {
		if _, err := os.Stat(libpath); err != nil {
			os.MkdirAll(libpath, 0550)
		}
		PackLibraries(e.cmd, libpath, e.cmdroot, e.cmdenvs)
	}
}

func (e *exeEntry) runADDCopy() error {
	adds := strings.Fields(e.add)
	copys := strings.Fields(e.copy)

	if len(adds) == 1 || len(adds) > 1 && !strings.HasPrefix(adds[len(adds)-1], "${") {
		adds = append(adds, "${indexPagePath}")
	}

	if len(copys) == 1 || len(copys) > 1 && !strings.HasPrefix(copys[len(copys)-1], "${") {
		copys = append(copys, "${indexPagePath}")
	}

	for i, add := range adds {
		adds[i] = e.expendVariable(add)
		if !strings.HasPrefix(adds[i], "/") {
			adds[i] = path.Join(e.cmdroot, adds[i])
		}
	}
	for i, copy := range copys {
		copys[i] = e.expendVariable(copy)
		if !strings.HasPrefix(copys[i], "/") {
			copys[i] = path.Join(e.cmdroot, copys[i])
		}
	}

	if l := len(adds); l >= 2 {
		for _, add := range adds[:l-1] {
			if file.IsSupportedPackFile(add) {
				if err := file.UnPackFile(add, adds[l-1]); err != nil {
					return err
				}
			} else {
				if err := file.CopyFiles(add, adds[l-1]); err != nil {
					return err
				}
			}
		}
	}

	if l := len(copys); l >= 2 {
		for _, copy := range copys[:l-1] {
			if err := file.CopyFiles(copy, copys[l-1]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *exeEntry) expendVariable(key string) string {
	var bts, btk bytes.Buffer
	p, tag := ' ', false
	for _, k := range key {
		if p == '$' && k == '{' {
			p = k
			tag = true
			continue
		}

		if k == '}' && tag {
			str := btk.String()
			for _, env := range e.cmdenvs {
				if strings.HasPrefix(env, str+"=") {
					bts.WriteString(strings.TrimPrefix(env, str+"="))
				}
			}
			btk.Reset()
			tag = false
			p = k
			continue
		}
		if tag {
			btk.WriteRune(k)
		} else if k != '$' {
			bts.WriteRune(k)
		}

		p = k
	}

	return bts.String()
}

func (e *exeEntry) addRuntimeEnv(envs []string) []string {
	if len(e.envcache) > 0 {
		return e.envcache
	}
	//add bin & current to PATH
	//add lib & current to LD_LIBRARY_PATH
	if apath, err := filepath.Abs(e.cmdroot); err == nil {
		bins := []string{apath}
		libs := []string{apath}
		if libpath, err := filepath.Abs(e.cmdroot + "/../lib"); err == nil {
			if df, err := os.Stat(libpath); err == nil && df.IsDir() {
				libs = append(libs, libpath)
			}
		}
		if libpath, err := filepath.Abs(path.Join(e.cmdroot, "lib")); err == nil {
			if df, err := os.Stat(libpath); err == nil && df.IsDir() {
				libs = append(libs, libpath)
			}
		}

		haspath, haslib := false, false
		for i, env := range envs {
			if strings.HasPrefix(env, "PATH=") {
				haspath = true
				envs[i] = fmt.Sprintf("PATH=%s:%s", strings.Join(bins, ":"), strings.TrimLeft(strings.TrimPrefix(env, "PATH="), " "))
			}
			if strings.HasPrefix(env, "LD_LIBRARY_PATH=") {
				haslib = true
				envs[i] = fmt.Sprintf("LD_LIBRARY_PATH=%s:%s", strings.Join(libs, ":"), strings.TrimLeft(strings.TrimPrefix(env, "LD_LIBRARY_PATH="), " "))
			}
		}

		if !haslib {
			envs = append(envs, fmt.Sprintf("LD_LIBRARY_PATH=%s", strings.Join(libs, ":")))
		}
		if !haspath {
			envs = append(envs, fmt.Sprintf("PATH=%s", strings.Join(bins, ":")))
		}
	}

	e.envcache = envs
	return envs
}

type appRunner struct {
	input  *appParamParser
	output *appParamParser

	cmd exeEntry
}

func (a *appRunner) build(ctl *control.Control) error {
	a.input = &appParamParser{params: ctl.Input.Stdin, sourceDataType: ctl.Input.Type}
	a.output = &appParamParser{params: ctl.Output.Stdout, sourceDataType: ctl.Output.Type, outputViewType: ctl.Output.View.Type, execWorkDir: ctl.Entrypoint.Workdir, indexPageDir: ctl.GetAppIndexPageHomePath()}
	if ctl.Output.View.Type == "sform" {
		a.output.sformbuilder = &sformcompiler.SformValueTransfomer{}
		if err := a.output.sformbuilder.LoadForm([]byte(ctl.Output.View.Inline_string)); err != nil {
			return err
		}
		//check if have params
		if len(a.output.params) == 0 {
			return fmt.Errorf("must specify output-stdout's params if output view is 'sform' type")
		}
	}
	a.cmd.cmdenvs = ctl.Entrypoint.Envs
	a.cmd.cmdenvs = append(a.cmd.cmdenvs, ctl.GetEnvs()...)

	a.cmd.cmdroot = ctl.Entrypoint.Workdir
	a.cmd.runner = ctl.Entrypoint.GetAppService()
	a.cmd.cmd = ctl.Entrypoint.Cmd
	a.cmd.ishell = ctl.Entrypoint.Inline_shell
	a.cmd.add = ctl.Entrypoint.Add
	a.cmd.copy = ctl.Entrypoint.Copy
	a.cmd.trace = ctl.Entrypoint.Trace
	a.cmd.period = ctl.Entrypoint.Period
	a.cmd.args = ctl.Entrypoint.Args
	a.cmd.appctrl = path.Join(ctl.ControlFilePath, "control.yaml")
	a.cmd.cmdtype = "entrypoint"

	if len(ctl.Entrypoint.Cmd) > 0 && ctl.Entrypoint.Packdepend {
		a.cmd.CopyDependence()
	}
	if len(ctl.Entrypoint.Add) > 0 || len(ctl.Entrypoint.Copy) > 0 {
		fmt.Println(a.cmd.runADDCopy())
	}

	return nil
}

func (a *appRunner) Run(w http.ResponseWriter, r *http.Request) {
	ps, err := a.input.ParseFromHttpRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		fmt.Println(err)
		return
	}
	bts, err := a.runApp(ps)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		fmt.Println(err)
		return
	}

	a.output.WriteHttpResponse(w, bts)
}

func (a *appRunner) runApp(ps []*appParam) ([]byte, error) {
	if params, closer, err := a.input.BuildExeParams(ps); err == nil {
		if closer != nil {
			defer closer.Close()
		}
		return a.cmd.run(params, a.envArgs(ps)...)
	} else {
		return nil, err
	}
}

func (a *appRunner) envArgs(ps []*appParam) []string {
	var envs []string
	for _, p := range ps {
		if p != nil && len(p.exeopt) > 0 && len(p.content) > 0 {
			envs = append(envs, fmt.Sprintf("arg_%s=%s", strings.TrimRight(strings.TrimLeft(p.exeopt, "-"), "-="), string(p.content)))
		}
	}
	return envs
}