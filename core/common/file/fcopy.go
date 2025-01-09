package file

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

func CopyFiles(source, targetPath string) error {
	source, targetPath = path.Clean(source), path.Clean(targetPath)
	if strings.HasPrefix(targetPath, source) {
		return fmt.Errorf("can not copy [%s] to [%s]", source, targetPath)
	}

	fstat, err := os.Stat(source)
	if err != nil {
		return err
	}

	targetSource := path.Join(targetPath, fstat.Name())

	if stat, err := os.Stat(targetSource); err == nil && stat.ModTime().Equal(fstat.ModTime()) {
		return nil
	}

	if fstat.IsDir() {
		os.MkdirAll(targetSource, fstat.Mode())
		os.Chtimes(targetSource, fstat.ModTime(), fstat.ModTime())
		if src, err := os.Open(source); err == nil {
			if subs, err := src.Readdir(-1); err == nil {
				for _, sub := range subs {
					if sub.Name() != "." && sub.Name() != ".." {
						if sub.IsDir() {
							CopyFiles(path.Join(source, sub.Name()), targetSource)
						} else {
							CopyFile(path.Join(source, sub.Name()), path.Join(targetSource, sub.Name()), sub.Mode(), sub.ModTime())
						}
					}
				}
			}
			src.Close()
		}
	} else {
		CopyFile(source, targetSource, fstat.Mode(), fstat.ModTime())
	}

	return nil
}

func CopyFile(source, target string, mode os.FileMode, modtime time.Time) error {
	if stat, err := os.Stat(target); err == nil && stat.ModTime().Equal(modtime) {
		return nil
	}

	srcFile, err := os.OpenFile(source, os.O_RDONLY, mode)
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	srcFile.Close()
	dstFile.Close()
	os.Chtimes(target, modtime, modtime)

	return nil
}
