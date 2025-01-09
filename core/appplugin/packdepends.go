package appplugin

import (
	"bufio"
	"os/exec"
	"strings"
)

type dep_ignores struct {
	dep_ignore []string
}

func (d *dep_ignores) Match(str string) bool {
	for _, c := range d.dep_ignore {
		if strings.Contains(str, c) {
			return true
		}
	}
	return false
}

var _dep_ignores *dep_ignores

func init() {
	_dep_ignores = &dep_ignores{
		dep_ignore: []string{"libstdc++.so", "libc.so", "libpthread.so", "linux-vdso.so", "ld-linux-x86-64.so"},
	}
}

func PackLibraries(excutable, targetpath string, workdir string, envs []string) bool {
	//exec.Command(fmt.Sprintf("ldd %s", excutable))
	copyed := make(map[string]bool)
	packLibraries(excutable, targetpath, workdir, envs, copyed)

	return true
}

func packLibraries(excutable, targetpath string, workdir string, envs []string, copyed map[string]bool) bool {
	if deps := dependencies(excutable, workdir, envs); len(deps) > 0 {
		for _, dep := range deps {
			if _, ok := copyed[dep]; !ok {
				copyed[dep] = true
				if !_dep_ignores.Match(dep) {
					filecopy(dep, targetpath, workdir, envs)
					packLibraries(dep, targetpath, workdir, envs, copyed)
				}
			}
		}
	}

	return true
}

func dependencies(excutable string, workdir string, envs []string) []string {
	run := exec.Command("ldd", excutable)
	run.Dir = workdir
	run.Env = append(envs, run.Environ()...)

	var depends []string
	if bts, err := run.Output(); err == nil && len(bts) > 0 {
		reader := bufio.NewReader(strings.NewReader(string(bts)))
		for {
			l, e := reader.ReadBytes('\n')
			if e != nil && len(l) == 0 {
				break
			}

			sl := string(l)
			if idx := strings.Index(sl, "=>"); idx > 0 {
				if end := strings.LastIndex(sl, "("); end > idx {
					depends = append(depends, strings.TrimSpace(sl[idx+2:end]))
				}
			}
		}
	}

	return depends
}

func filecopy(src, dest string, workdir string, envs []string) error {
	run := exec.Command("cp", src, dest)
	run.Dir = workdir
	run.Env = append(envs, run.Environ()...)
	return run.Run()
}
