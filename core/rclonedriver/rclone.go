package rclonedriver

import (
	"context"
	"fmt"
	"io"
	iofs "io/fs"
	"net/http"
	_ "onlinetools/core/rclonedriver/rclone/rclone/backend/all"
	"onlinetools/core/rclonedriver/rclone/rclone/fs"
	"onlinetools/core/rclonedriver/rclone/rclone/fs/config/obscure"
	"onlinetools/core/rclonedriver/rclone/rclone/fs/object"
	"os"
	"path"
	"time"
)

func SetRcloneAuth(section, user, pass string) error {
	if err := SetRcloneConf(section, "user", user); err != nil {
		return err
	}

	out, err := obscure.Obscure(pass)
	if err == nil {
		if err := SetRcloneConf(section, "pass", out); err != nil {
			return err
		}
	}
	return err
}

func GetRcloneConf(section, key string) (string, bool) {
	return fs.ConfigFileGet(section, key)
}
func SetRcloneConf(section, key, value string) error {
	return fs.ConfigFileSet(section, key, value)
}

// GetCloudDiskFileSystemByPath makes a new Fs object from the path
// The path is of the form remote:path, such as alist:liyun/music
func GetCloudDiskFileSystemByPath(path string) (fs.Fs, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	//config redirect request http header, for aliyunpan
	cctx, conf := fs.AddConfig(ctx)
	conf.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		req.Header.Del("Referer")
		return nil
	}

	return fs.NewFs(cctx, path)
}

// The remotepath is of the form remote:path, such as alist:liyun/music
func Download(remotepath, localpath string) error {

	//mkdir of local path
	if fin, err := os.Stat(localpath); err != nil {
		if err := os.MkdirAll(localpath, 0755); err != nil {
			return err
		}
	} else if !fin.IsDir() {
		return fmt.Errorf("[%s] should be a dir", localpath)
	}

	//access remotepath
	rfs, err := GetCloudDiskFileSystemByPath(remotepath)
	if err != nil {
		return err
	}

	//recursing download remote files
	if entries, err := rfs.List(context.Background(), ""); err == nil {
		entries.ForObject(func(o fs.Object) {
			if o.Storable() {
				fpath := path.Join(localpath, o.Remote())
				if finfo, err := os.Stat(fpath); err != nil || !finfo.IsDir() {

					if tm := o.ModTime(context.Background()); err != nil || !tm.Equal(finfo.ModTime()) {
						if reader, err := o.Open(context.Background(), &fs.HTTPOption{Key: "User-Agent", Value: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Mobile Safari/537.36"}, &fs.HTTPOption{Key: "Accept", Value: "*/*"}); err == nil {
							//if reader, err := o.Open(context.Background()); err == nil {
							if file, err := os.Create(fpath); err == nil {
								io.Copy(file, reader)
								file.Close()
								os.Chtimes(fpath, tm, tm)
								fmt.Printf("%s downloaded\n", fpath)
							}
							reader.Close()
						} else {
							fmt.Println(err)
						}
					}
				}
			}
		})

		entries.ForDir(func(d fs.Directory) {
			Download(path.Join(remotepath, d.Remote()), path.Join(localpath, d.Remote()))
		})
	} else {
		return err
	}

	return nil
}

// The remotepath is of the form remote:path, such as alist:liyun/music
func Upload(remotepath, localpath string, delUploaded bool) error {
	//mkdir of local path
	fin, err := os.Stat(localpath)
	if err != nil || !fin.IsDir() {
		fmt.Printf("Error: local path [%s] not a dir or can not access", localpath)
		return err
	}

	//access remotepath
	rfs, err := GetCloudDiskFileSystemByPath(remotepath)
	if err != nil {
		return err
	}

	if entries, err := os.ReadDir(localpath); err == nil {
		for _, entry := range entries {
			lpath := path.Join(localpath, entry.Name())
			if entry.IsDir() {
				if !isEmptyDir(lpath) {
					rfs.Mkdir(context.Background(), entry.Name())
					Upload(path.Join(remotepath, entry.Name()), lpath, delUploaded)
				}
			} else if entry.Type() != iofs.ModeTemporary {
				if finfo, err := os.Stat(lpath); err == nil {
					if f, err := os.Open(lpath); err == nil {
						obinfo := object.NewStaticObjectInfo(entry.Name(), finfo.ModTime(), finfo.Size(), true, nil, nil)
						_, err := rfs.Put(context.Background(), f, obinfo)
						f.Close()
						if err != nil {
							fmt.Printf("%s Error: %s\n", lpath, err.Error())
						} else if delUploaded {
							os.Remove(lpath)
						} else {
							fmt.Printf("%s uploaded\n", lpath)
						}
					}
				}
			}
		}
	}

	return nil
}

func isEmptyDir(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)

	return err == io.EOF
}
