package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xi2/xz"
)

// is supported pack file .tar.gz .zip .tar.xz .tar .gzip
func IsSupportedPackFile(packed string) bool {
	if strings.HasSuffix(packed, ".tar.gz") ||
		strings.HasSuffix(packed, ".tar.xz") ||
		strings.HasSuffix(packed, ".zip") ||
		strings.HasSuffix(packed, ".gzip") ||
		strings.HasSuffix(packed, ".tar") {

		return true
	}

	return false
}

// supported file type .tar.gz .zip .tar.xz .tar .gzip
func UnPackFile(packed, saveroot string) error {
	packed, saveroot = path.Clean(packed), path.Clean(saveroot)
	if strings.HasPrefix(saveroot, packed) {
		return fmt.Errorf("can not unpack [%s] to [%s]", packed, saveroot)
	}

	archive, err := os.Open(packed)
	if err != nil {
		return err
	}
	defer archive.Close()

	if strings.HasSuffix(packed, ".tar.gz") {
		untargz(archive, saveroot)
	} else if strings.HasSuffix(packed, ".tar.xz") {
		untarxz(archive, saveroot)
	} else if strings.HasSuffix(packed, ".tar") {
		untar(archive, saveroot)
	} else if strings.HasSuffix(packed, ".gzip") {
		ungzip(archive, saveroot)
	} else if strings.HasSuffix(packed, ".zip") {
		unzip(packed, saveroot)
	} else {
		return fmt.Errorf("not supported packed filetype %s", packed)
	}

	return nil
}

func unzip(ziped, dst string) error {
	archive, err := zip.OpenReader(ziped)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path")
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			os.Chtimes(filePath, f.FileInfo().ModTime(), f.FileInfo().ModTime())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		if fstat, err := os.Stat(filePath); err == nil && fstat.ModTime().Equal(f.FileInfo().ModTime()) {
			continue
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
		os.Chtimes(filePath, f.FileInfo().ModTime(), f.FileInfo().ModTime())
	}

	return nil
}

func untargz(reader io.Reader, target string) error {
	//unpack .tar.gz
	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	return untar(archive, target)
}

func untarxz(reader io.Reader, target string) error {
	//unpack .tar.xz
	archive, err := xz.NewReader(reader, 0)
	if err != nil {
		return err
	}

	return untar(archive, target)
}

func ungzip(reader io.Reader, target string) error {

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func untar(reader io.Reader, target string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()

		if fstat, err := os.Stat(path); err == nil && fstat.ModTime().Equal(info.ModTime()) {
			continue
		}

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			os.Chtimes(path, info.ModTime(), info.ModTime())
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
		file.Close()
		os.Chtimes(path, info.ModTime(), info.ModTime())
	}
	return nil
}

func Tar(source, target string) error {
	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.tar", filename))
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

func Gzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}

	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.gz", filename))
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filename
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	return err
}
