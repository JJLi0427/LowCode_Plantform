package appplugin

import (
	"fmt"
	"net/textproto"
	"path/filepath"
	"strings"
)

type ParamType int
type ContentType = string

const (
	FILE ParamType = iota
	STRING
	NUMBER
)

func contentTypeByMIME(mime textproto.MIMEHeader) ContentType {
	//fmt.Println("formh", mime)
	if content := mime.Get("Content-Disposition"); len(content) > 0 {
		if nameidx := strings.Index(content, "filename="); nameidx > 0 {
			content = trimQ(strings.Fields(content[nameidx+9:])[0])
		}
		if ext := filepath.Ext(content); len(ext) > 0 {
			if ext == ".jpg" {
				ext = ".jpeg"
			}
			return ext
		}
	}

	if content := strings.ToLower(mime.Get("Content-Type")); len(content) > 0 {
		if content == "image/png" {
			return ".png"
		} else if content == "image/jpeg" || content == "image/jpg" {
			return ".jpeg"
		} else if content == "image/ico" {
			return ".ico"
		} else if content == "image/gif" {
			return ".gif"
		} else if content == "image/webp" {
			return ".webp"
		} else if content == "image/bmp" {
			return ".bmp"
		}
	}

	return ".txt"
}

func contentType(typ string) ContentType {
	// typ is http content type or string type
	return typ
}

func paramType(typ string) ParamType {
	if typ == "txt" {
		return STRING
	} else if typ == "number" {
		return NUMBER
	} else if typ == "file" {
		return FILE
	}

	return STRING
}

//func contentFileExt(typ string) string { }
func contentFileExt(typ ContentType) string {
	typ =strings.Trim(strings.ToLower(typ), "/\t\n*\r")
	if ll := filepath.Ext(typ); len(ll) > 0 && len(ll) <= 6 {
		return typ
	}

	if l := len(typ); l > 0 && l <= 6 {
		return fmt.Sprintf(".%s", typ)
	}

	return ".txt"
}
