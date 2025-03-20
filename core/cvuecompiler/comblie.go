package cvuecompiler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var vueexportdefaulttokens []string
var vueexportdefaulttokensmaxlen int
var vuecomponenttokens []string
var vuecomponenttokensmaxlen int
var vueexportdefaultregexp *regexp.Regexp
var vueHtmlTagexp *regexp.Regexp
var vueStyleTagexp *regexp.Regexp

func trimQ(s string) string {
 	return strings.TrimRight(strings.TrimLeft(s, "\"' "), "\"' ")
}

func init() {
	vueexportdefaulttokens = append(vueexportdefaulttokens, "methods", "name", "components", "data", "watch")
	vuecomponenttokens = append(vuecomponenttokens,
		"<template>",
		"<template ",
		"</template>",
		"<script>",
		"<script ",
		"</script>",
		"<style ",
		"<style>",
		"</style>")

	for _, e := range vueexportdefaulttokens {
		if len(e) > vueexportdefaulttokensmaxlen {
			vueexportdefaulttokensmaxlen = len(e)
		}
	}
	for _, e := range vuecomponenttokens {
		if len(e) > vuecomponenttokensmaxlen {
			vuecomponenttokensmaxlen = len(e)
		}
	}
	vueexportdefaultregexp = regexp.MustCompile(`export +default +{`)
	vueHtmlTagexp = regexp.MustCompile(`</?[a-z0-9_-]*[A-Z]+[a-zA-Z0-9_-]*[ >]`)
	vueStyleTagexp = regexp.MustCompile(`<[/]? *style[^>]*>`)
}

type componenttokenparser struct {
	tokens []string
	bt     []byte
	l      int
}

func (c *componenttokenparser) token(b byte) string {
	if c.l == 0 || c.bt == nil {
		c.l = vuecomponenttokensmaxlen * 2
		c.tokens = vuecomponenttokens
		c.bt = make([]byte, 0, c.l)
	}
	c.bt = append(c.bt, b)

	for _, t := range c.tokens {
		if strings.HasSuffix(string(c.bt), t) {
			return t
		}
	}

	if len(c.bt) > c.l {
		c.bt = c.bt[c.l/2:]
	}

	return ""
}

type VueComponent struct {
	Name             string
	Template         string
	VueExportDefault string
	ChildComponents  []string
	Scripts          string
	Styles           string

	AddAddtionalMethods []string
	AddAddtionalDivs    []string

	HasVuetify bool
}

func (v *VueComponent) AddChildComponent(name string) bool {
	for _, s := range v.ChildComponents {
		if s == name {
			return false
		}
	}
	v.ChildComponents = append(v.ChildComponents, name)
	return true
}

func (v *VueComponent) FormateTemplate() error {
	var btt bytes.Buffer
	cnt := 0
	p, pi := byte(' '), 0
	for i, c := range []byte(v.Template) {
		if c == '`' && p != '\\' {
			if cnt%2 == 1 {
				bak := &BackTickString{Raw: v.Template[pi : i+1]}
				if err := bak.Parse(); err != nil {
					return err
				}
				if d, err := bak.ConvertToJsString(); err != nil {
					return err
				} else {
					btt.WriteString(d)
				}
			}

			pi = i
			cnt++
		} else if cnt%2 == 0 {
			btt.WriteByte(c)
		}

		p = c
	}

	v.Template = btt.String()
	return nil
}

func (v *VueComponent) WriteTempate(s []byte) error {
	ss := string(s)
	ss = strings.TrimPrefix(ss, "<template>")
	ss = strings.TrimSuffix(ss, "</template>")
	ss = strings.TrimPrefix(ss, "\n")
	ss = strings.TrimSuffix(ss, "\n")

	if len(v.AddAddtionalDivs) > 0 {
		if idx := strings.LastIndex(ss, "</"); idx > 0 {
			var btt bytes.Buffer
			btt.WriteString(ss[:idx])
			for _, s := range v.AddAddtionalDivs {
				btt.WriteString(s)
			}
			btt.WriteString(ss[idx:])
			ss = btt.String()
		}
	}

	v.Template = ss
	v.FormateTemplate()
	v.formatTemplateTags()

	return nil
}

func (v *VueComponent) WriteScript(s []byte) error {
	if se := vueexportdefaultregexp.FindIndex(s); len(se) == 2 {
		parser := &componenttokenparser{
			tokens: vueexportdefaulttokens,
			l:      vueexportdefaulttokensmaxlen * 2,
			bt:     make([]byte, 0, vueexportdefaulttokensmaxlen*2),
		}
		scope, s1, s2, eidx, methodidx, componentsidx, componentseidx := 1, false, false, 0, 0, 0, 0
		for i, c := range s[se[1]:] {
			if token := parser.token(c); token != "" && scope == 1 {
				if strings.HasPrefix(token, "name") {
					var bts bytes.Buffer
					open := false
					for _, m := range s[i+se[1]:] {
						if (m == '"' || m == '\'') && open {
							break
						} else if m == '"' || m == '\'' {
							open = true
						} else if open {
							bts.WriteByte(m)
						}
					}
					v.Name = bts.String()
					if err := v.formatComponentName(); err != nil {
						return err
					}
				}
				if strings.HasPrefix(token, "methods") {
					methodidx = i
					for k, m := range s[i+se[1]:] {
						if m == '{' {
							methodidx = methodidx + k + 1
							break
						}
					}
				}
				if strings.HasPrefix(token, "components") {
					componentsidx = i - 9
					var bts bytes.Buffer
					open, ignore := false, false
					for k, m := range s[i+se[1]:] {
						if m == '{' || m == '[' {
							ignore = false
							open = true
						} else if m == '}' || m == ']' {
							componentseidx = i + k + 1
							if btt := bts.String(); len(btt) > 0 {
								if btts, err := v.FormatComponentName(strings.TrimSpace(btt)); err == nil {
									v.AddChildComponent(btts)
									//v.ChildComponents = append(v.ChildComponents, btts)
								} else {
									return err
								}
								bts.Reset()
							}
							break
						} else if m == ':' {
							ignore = true
						} else if m == ',' {
							ignore = false
							if btt := bts.String(); len(btt) > 0 {
								if btts, err := v.FormatComponentName(strings.TrimSpace(btt)); err == nil {
									v.AddChildComponent(btts)
									//v.ChildComponents = append(v.ChildComponents, btts)
								} else {
									return err
								}
								bts.Reset()
							}
						} else if open && !ignore {
							bts.WriteByte(m)
						}
					}
					//v.Name = bts.String()
				}
			}
			if c == '{' && !s1 && !s2 {
				scope++
			}
			if c == '}' && !s1 && !s2 {
				scope--
			}
			if c == '\'' {
				s1 = !s1
			}
			if c == '"' {
				s2 = !s2
			}
			if scope == 0 && !s1 && !s2 && c == '}' {
				eidx = i
				break
			}
		}

		if eidx > se[1] {
			v.VueExportDefault = string(s[se[1] : se[1]+eidx])

			if methodidx > 0 && componentsidx > methodidx && componentsidx < componentseidx {
				var btt bytes.Buffer
				btt.WriteString(v.VueExportDefault[:componentsidx])
				for i, c := range v.VueExportDefault[componentseidx:] {
					if c == ' ' || c == '\t' {
						continue
					} else if c == ',' {
						componentseidx += i + 1
					}
					break
				}
				btt.WriteString(v.VueExportDefault[componentseidx:])
				v.VueExportDefault = btt.String()
			}

			if len(v.AddAddtionalMethods) > 0 {
				if methodidx > 0 {
					var btt bytes.Buffer
					btt.WriteString(v.VueExportDefault[:methodidx])
					for _, s := range v.AddAddtionalMethods {
						btt.WriteString(s)
						if s[len(s)-1] != ',' {
							btt.WriteByte(',')
						}
					}
					btt.WriteString(v.VueExportDefault[methodidx:])
					v.VueExportDefault = btt.String()
				} else {
					var btt bytes.Buffer
					btt.WriteString("methods:{")
					for _, s := range v.AddAddtionalMethods {
						btt.WriteString(s)
						if s[len(s)-1] != ',' {
							btt.WriteByte(',')
						}
					}
					btt.WriteString("},")
					v.VueExportDefault += btt.String()
				}
			}

			if (methodidx == 0 && componentsidx > 0) || componentsidx < methodidx {
				var btt bytes.Buffer
				btt.WriteString(v.VueExportDefault[:componentsidx])
				for i, c := range v.VueExportDefault[componentseidx:] {
					if c == ' ' || c == '\t' {
						continue
					} else if c == ',' {
						componentseidx += i + 1
					}
					break
				}
				btt.WriteString(v.VueExportDefault[componentseidx:])
				v.VueExportDefault = btt.String()
			}

			return nil
		}
		return fmt.Errorf("parse Vue ExportDefault failed")
	}

	v.Scripts += string(s)
	return nil
}

func (v *VueComponent) WriteStyle(s []byte) error {
	pairslice := vueStyleTagexp.FindAllIndex(s, -1)
	if len(pairslice) > 0 {
		if len(pairslice) != 2 {
			return fmt.Errorf("error write vuecomponet <style> %s", string(s))
		}

		s = s[pairslice[0][1]:pairslice[1][0]]
	}

	v.Styles += string(s)

	return nil
}

func (v *VueComponent) GenerateComponentNameByFileName(f string) bool {
	if strings.HasSuffix(f, ".vue") {
		_, name := filepath.Split(f)
		v.Name = strings.TrimSuffix(name, ".vue")
		if err := v.formatComponentName(); err != nil {
			return false
		}

		return true
	}
	return false
}

func (v *VueComponent) formatTemplateTags() {
	templatebytes := []byte(v.Template)
	if pairslice := vueHtmlTagexp.FindAllIndex(templatebytes, -1); pairslice != nil {
		var bts bytes.Buffer
		pi := 0
		for _, pair := range pairslice {
			bts.Write(templatebytes[pi:pair[0]])
			tag := templatebytes[pair[0]:pair[1]]
			tagidx0, tagidx1 := 0, 0
			for i, c := range tag {
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || (c == '-') || (c == '_') {
					if tagidx0 == 0 {
						tagidx0 = i
					}
					tagidx1 = i
				}
			}
			bts.Write(tag[0:tagidx0])
			name, err := v.FormatComponentName(string(tag[tagidx0 : tagidx1+1]))
			if err != nil {
				panic(err)
			}
			v.AddChildComponent(name)
			bts.WriteString(name)
			bts.Write(tag[tagidx1+1:])

			pi = pair[1]
		}

		bts.Write(templatebytes[pi:])

		v.Template = bts.String()
	}
}

func (v *VueComponent) formatComponentName() error {
	if n, e := v.FormatComponentName(v.Name); e != nil {
		return e
	} else {
		v.Name = n
	}

	return nil
}

func (v *VueComponent) FormatComponentName(Name string) (string, error) {
	Name = trimQ(Name)

	var btt bytes.Buffer
	for i, c := range []byte(Name) {
		if i == 0 && c >= 'A' && c <= 'Z' {
			btt.WriteByte('a' + c - 'A')
		} else if i > 0 && c >= 'A' && c <= 'Z' {
			btt.WriteByte('-')
			btt.WriteByte('a' + c - 'A')
		} else if (c >= 'a' && c <= 'z') || c == '-' {
			btt.WriteByte(c)
		} else {
			return "", fmt.Errorf("format vue component name[%s] failed", v.Name)
		}
	}
	return btt.String(), nil
}

func (v *VueComponent) BuildFromComponentFile(f string) error {
	if cf, err := os.ReadFile(f); err == nil {
		if err = v.BuildFromComponentFileContent(cf); err != nil {
			return err
		}
	} else {
		return err
	}

	if len(v.Name) == 0 && !v.GenerateComponentNameByFileName(f) {
		return fmt.Errorf("GenerateComponentNameByFileName[%s] failed, should specify [name] attr in vue [export default]", f)
	}

	//if len(v.Name) > 0 {
	//	if err := v.formatComponentName(); err != nil {
	//		return err
	//	}
	//}

	return nil
}
func (v *VueComponent) BuildFromComponentFileContent(cf []byte) error {
	parser := &componenttokenparser{}
	template, script, style := 0, 0, 0
	var bts bytes.Buffer
	for _, c := range cf {
		if token := parser.token(c); len(token) > 0 {
			if strings.HasPrefix(token, "<template") {
				if template == 0 {
					bts.Reset()
					bts.WriteString(token)
				} else {
					bts.WriteByte(c)
				}
				template++
			} else if strings.HasPrefix(token, "<script") {
				if template == 0 {
					bts.Reset()
					bts.WriteString(token)
				} else {
					bts.WriteByte(c)
				}
				script++
			} else if strings.HasPrefix(token, "<style") {
				if template == 0 {
					bts.Reset()
					bts.WriteString(token)
				} else {
					bts.WriteByte(c)
				}
				style++
			} else if strings.HasPrefix(token, "</template") {
				template--
				if template == 0 {
					bts.WriteByte(c)
					if err := v.WriteTempate(bts.Bytes()); err != nil {
						return err
					}
					bts.Reset()
				} else {
					bts.WriteByte(c)
				}
			} else if strings.HasPrefix(token, "</script") {
				script--
				if template == 0 && script == 0 {
					bts.WriteByte(c)
					if err := v.WriteScript(bts.Bytes()); err != nil {
						return err
					}
					bts.Reset()
				} else {
					bts.WriteByte(c)
				}
			} else if strings.HasPrefix(token, "</style") {
				style--
				if template == 0 && style == 0 {
					bts.WriteByte(c)
					if err := v.WriteStyle(bts.Bytes()); err != nil {
						return err
					}
					bts.Reset()
				} else {
					bts.WriteByte(c)
				}
			} else {
				return fmt.Errorf("invalid token %s", token)
			}
		} else {
			bts.WriteByte(c)
		}
	}
	if script != 0 || style != 0 {
		return fmt.Errorf("must use </script> or </style> end tag")
	}

	if len(v.VueExportDefault) == 0 {
		v.VueExportDefault = "{}"
	}

	if len(v.Template) > 0 {
		var btt bytes.Buffer
		btt.WriteString("{\n")
		btt.WriteString("template:")
		btt.WriteByte('`')
		btt.WriteString(v.Template)
		btt.WriteByte('`')
		btt.WriteByte(',')
		btt.WriteString(v.VueExportDefault)
		btt.WriteString("\n}")
		v.VueExportDefault = btt.String()

		//fmt.Println("VueExportDefault", v.VueExportDefault )
	} else {
		return fmt.Errorf("failed to parse vue file")
	}

	return nil
}
