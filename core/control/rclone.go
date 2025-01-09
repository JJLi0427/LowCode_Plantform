package control

import (
	"fmt"
	"onlinetools/core/rclonedriver"
	"path"
	"time"
)

type RClone struct {
	Remote        string `json:"remote"`
	RemotePath    string `json:"remotePath"`
	LocalCtlPath  string `json:"localCtlPath"`
	LocalViewPath string `json:"localViewPath"`

	Upload        bool `json:"upload"`
	CleanUploaded bool `json:"cleanUploaded"`

	CheckFrequencyInSecond int `json:"checkFrequencyInSecond"`

	viewroot string
	ctlroot  string
}

func (r *RClone) IsValid() bool {
	return len(r.Remote) > 0 && len(r.RemotePath) > 0 && (len(r.LocalCtlPath) > 0 || len(r.LocalViewPath) > 0)
}

func (r *RClone) LocalPath() string {
	localPath := path.Join(r.ctlroot, r.LocalCtlPath)
	if len(r.LocalCtlPath) == 0 {
		localPath = path.Join(r.viewroot, r.LocalViewPath)
	}

	return localPath
}

func (r *RClone) Run() {
	if !r.IsValid() {
		return
	}
	runner := func() {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println(e)
			}
		}()

		duration, _ := time.ParseDuration(fmt.Sprintf("%ds", r.CheckFrequencyInSecond))

		localPath := r.LocalPath()
		remotePath := fmt.Sprintf("%s:%s", r.Remote, r.RemotePath)

		for {
			if r.Upload {
				rclonedriver.Upload(remotePath, localPath, r.CleanUploaded)
			} else {
				rclonedriver.Download(remotePath, localPath)
			}
			if r.CheckFrequencyInSecond <= 1 {
				break
			}
			time.Sleep(duration)
		}
	}

	go runner()
}
