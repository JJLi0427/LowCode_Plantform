package file

import (
	"os"
	"sync"
	"time"
	"crypto/tls"
)

func NewFileChecker(files []string) *FileChecker {
	return &FileChecker{Files: files}
}

type FileChecker struct {
	Files []string

	tamps []string
}

func (f *FileChecker) FileChanged () bool {
	ret := false

	if l := len(f.Files); l != len(f.tamps) {
		ret = true
		f.tamps = make([]string,l, l)
	}

	for i, s :=range f.Files {
		if info, err := os.Stat(s); err == nil {
			if tm := info.ModTime().String(); f.tamps[i] != tm {
				ret = true
				f.tamps[i] = tm
			}
		}
	}


	return ret
}


type DynamicCertificate struct {
	certFile string
	keyFile  string

	cert *tls.Certificate
	sync.RWMutex
}

func (d *DynamicCertificate) Init (crtfile, keyfile string) error {
        cert, err := tls.LoadX509KeyPair(crtfile, keyfile)
        if err != nil {
                return err
        }
	d.cert = &cert
	d.certFile = crtfile
	d.keyFile  = keyfile

	go func () {
		checker := NewFileChecker([]string{d.certFile, d.keyFile})
		checker.FileChanged()
		for {
			time.Sleep(time.Second*1200)
			if checker.FileChanged() {
        			if cert, err := tls.LoadX509KeyPair(d.certFile, d.keyFile); err == nil {
					d.Lock()
					d.cert = &cert
					d.Unlock()
				}
			}
		}
	}()

        return nil 
}


func (d *DynamicCertificate) GetCertificate() *tls.Certificate {
	d.RLock()
	defer d.RUnlock()
	return d.cert
}


