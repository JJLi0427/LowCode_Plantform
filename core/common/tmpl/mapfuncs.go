package tmpl

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var SiteTitleTag string

var linkTagRegexp *regexp.Regexp
var scriptTagRegexp *regexp.Regexp
var metaTagRegexp *regexp.Regexp
//var hrefRegexp *regexp.Regexp
//var srcRegexp *regexp.Regexp

func init() {
	linkTagRegexp = regexp.MustCompile(`^<[ \t]*link[ \t]+`)
	scriptTagRegexp = regexp.MustCompile(`^<[ \t]*script[ \t]+`)
	metaTagRegexp = regexp.MustCompile(`^<[ \t]*meta[ \t]+`)
	//hrefRegexp = regexp.MustCompile(`[ \t]+href[ \t]+=[ \t]*["'][^ ]+["'][ \t>]`)
	//srcRegexp = regexp.MustCompile(`[ \t]+src[ \t]+=[ \t]*["'][^ ]+["'][ \t>]`)
}

func AppTitle(title string) string {
	if !strings.Contains(title, SiteTitleTag) {
		return fmt.Sprintf("%s - %s", title, SiteTitleTag)
	}
	return title
}

func ValidateMeta(metaInControl string) string {
	if metaTagRegexp.Match([]byte(metaInControl)) {
		return metaInControl
	}

	panic(fmt.Sprintf("format of meta[%s] in control.yaml is invalid, only accept xml node", metaInControl))
}

func LinkAttributes(linkInControl string) string {
	if attrs, dcnt := Attributes(linkInControl, "href", false); dcnt <= 1 {
		if !linkTagRegexp.Match([]byte(attrs)) {
			attrs = fmt.Sprintf("<link %s >", attrs)
		}

		return attrs
	}
	panic(fmt.Sprintf("link[%s] in control.yaml is invalid", linkInControl))
}

func ScriptAttributes(scriptInControl string) string {
	if attrs, dcnt := Attributes(scriptInControl, "src", false); dcnt <= 1 {
		if !scriptTagRegexp.Match([]byte(attrs)) {
			attrs = fmt.Sprintf("<script %s ></script>", attrs)
		}

		return attrs
	}

	panic(fmt.Sprintf("script[%s] in control.yaml is invalid", scriptInControl))
}

func Attributes(src, defkey string, onlyUrl bool) (string, int) {

	cneq, def, url, ext := 0, defkey, "", ""
	var attrs []string
	if idxs, l := firstMatchedTagIndex([]byte(src)); len(idxs) == 2 {
		attrs = append(attrs, src)
		url, cneq = l, 1
	} else {
		fields := strings.Fields(src)
		for _, field := range fields {

			if strings.HasPrefix(field, defkey) && strings.Contains(field, "=") {
				attrs = append(attrs, field)
				url = strings.TrimLeft(strings.TrimPrefix(field, defkey), " =")
				cneq++
			} else if idx := strings.Index(field, "="); idx >= 0 && !strings.Contains(field[:idx], "?") {
				attrs = append(attrs, field)
			} else if ll := len(field); ll > 0 {
				if (field[0] == '"' && field[ll-1] == '"') || (field[0] == '\'' && field[ll-1] == '\'') {
					attrs = append(attrs, fmt.Sprintf("%s=%s", def, field))
				} else {
					attrs = append(attrs, fmt.Sprintf("%s=\"%s\"", def, field))
					ext = filepath.Ext(field)
				}
				cneq++
				url = field
			}
		}

	}

	if onlyUrl {
		return strings.TrimRight(strings.TrimLeft(url, "\"'"), "'\""), cneq
	}

	if dattrs := furlAttrs(ext); len(dattrs) > 0 {
		for _, attr := range dattrs {
			has := false
			for _, elem := range attrs {
				if strings.HasPrefix(elem, attr[0]) {
					has = true
				}
			}
			if !has {
				attrs = append(attrs, fmt.Sprintf("%s=%s", attr[0], attr[1]))
			}
		}
	}

	return strings.Join(attrs, " "), cneq
}

func firstMatchedTagIndex(field []byte) ([]int, string) {
	if idxs := linkTagRegexp.FindIndex([]byte(field)); len(idxs) == 2 {
		url := attr(string(field), "href")
		return idxs, url
	} else if idxs = scriptTagRegexp.FindIndex([]byte(field)); len(idxs) == 2 {
		url := attr(string(field), "src")
		return idxs, url
	}

	return nil, ""
}

func furlAttrs(ext string) [][2]string {
	var attrs [][2]string
	if ext == ".js" {
		attrs = append(attrs, [2]string{"type", `"text/javascript"`})
	} else if ext == ".ts" {
		attrs = append(attrs, [2]string{"type", `"ts/module"`})
	} else if ext == ".css" {
		attrs = append(attrs, [2]string{"rel", `"stylesheet"`})
		attrs = append(attrs, [2]string{"type", `"text/css"`})
	} else if ext == ".less" {
		attrs = append(attrs, [2]string{"rel", `"stylesheet/less"`})
		attrs = append(attrs, [2]string{"type", `"text/less"`})
	} else if ext == ".scss" {
	} else if ext == ".png" {
	} else if ext == ".jpg" {
	}

	return attrs
}

// exp. href = "ddd d" ==>return "ddd d"
// exp. href =  ddd  ==>return ddd
// exp. href = 'ddd' ==>return 'ddd'
func attr(src string, key string) string {
	if idx := strings.Index(src, key); idx >= 0 {
		var bts bytes.Buffer
		qq := 0
		for _, c := range []byte(strings.TrimLeft(src[idx+len(key):], " \t=")) {
			if qq%2 == 0 && (c == ' ' || c == '>' || c == '\t') {
				break
			} else {
				bts.WriteByte(c)
			}

			if c == '"' || c == '\'' {
				qq += 1
			}
		}
		return bts.String()
	}

	return ""
}
