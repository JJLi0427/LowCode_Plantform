package acme

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
)

type AcmeCA struct {
	AcmeshFile string
	WebRoot    string
	CAPath     string
	WebDomains []string
}

func (a *AcmeCA) Init() error {
	// issue a cert
	var params []string
	params = append(params, a.AcmeshFile)
	params = append(params, "--issue")
	has := false
	for _, dm := range a.WebDomains {
		if len(dm) > 0 {
			params = append(params, "-d")
			params = append(params, dm)
			has = true
		}
	}
	if !has {
		return fmt.Errorf("need input at least one domain name for acme.sh issue a CA")
	}

	params = append(params, "-w", a.WebRoot)
	params = append(params, "--key-file", path.Join(a.CAPath, "webtools.key"))
	params = append(params, "--cert-file", path.Join(a.CAPath, "webtools.crt"))
	params = append(params, "--force")

	exe := exec.Command("/bin/bash", params...)
	if stdout, err := exe.StdoutPipe(); err == nil {
		defer stdout.Close()
		go func() {
			io.Copy(os.Stdout, stdout)
		}()
	}

	return exe.Run()
}

func (a *AcmeCA) Run() {
	// renew a cert
	var params []string
	params = append(params, a.AcmeshFile)
	params = append(params, "--renew")
	for _, dm := range a.WebDomains {
		if len(dm) > 0 {
			params = append(params, "-d")
			params = append(params, dm)
		}
	}

	params = append(params, "-w", a.WebRoot)
	params = append(params, "--key-file", path.Join(a.CAPath, "webtools.key"))
	params = append(params, "--cert-file", path.Join(a.CAPath, "webtools.crt"))

	exe := exec.Command("/bin/bash", params...)
	out, _ := exe.CombinedOutput()
	fmt.Println("acms.sh renew ca:", string(out))
}
