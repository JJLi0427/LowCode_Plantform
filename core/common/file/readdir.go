package file

import (
	"fmt"
	"os"
)

func ReadFilesInDir(dir string, withpath bool, accept func(os.FileInfo, string) bool) ([]string, error) {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("ReadFiles failed for dir[%s], %s", dir, err.Error())
	}

	var files []string
	if d, err := os.Open(dir); err == nil {
		defer d.Close()
		if infos, err := d.Readdir(-1); err == nil {
			for _, info := range infos {
				if info.IsDir() {
					if fs, err := ReadFilesInDir(dir+"/"+info.Name(), withpath, accept); err == nil {
						files = append(files, fs...)
					} else {
						return nil, err
					}
				} else if accept(info, dir) {
					if withpath {
						files = append(files, dir+"/"+info.Name())
					} else {
						files = append(files, info.Name())
					}
				}
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	return files, nil
}
