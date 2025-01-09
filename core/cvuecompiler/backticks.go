package cvuecompiler

import (
	"fmt"
	"bytes"
)

type BackTickString struct {
	Raw string
	raws []string
	vars []string
}

func (b *BackTickString) Init(backticks string) {
	b.Raw = backticks
	b.raws = nil
	b.vars = nil
}

func (b *BackTickString) Parse() error {
	if s, l :=0, len(b.Raw)-1; s < l && b.Raw[s] == '`' && b.Raw[l] == '`' {
		b.Raw = b.Raw[1:l]
		var btt bytes.Buffer
		p1, p2, pi, open := byte(' '),byte(' '), 0, false
		for i, c := range []byte(b.Raw) {
			if c == '}' && open {
				pi = i+1
				open = false
				if i == len(b.Raw)-1 {
					b.raws = append(b.raws, "")
				}
				if btt.Len() > 0 {
					b.vars = append(b.vars, btt.String())
					btt.Reset()
				}else{
					return fmt.Errorf("parse backticks-string var failed [%s]", b.Raw)
				}
			//${add} //\${add}
			}else if c == '{' && p1 == '$' && p2 != '\\' {
				b.raws = append(b.raws, b.Raw[pi:i-1])
				open = true
			}else if open {
				if c == '{' || c == ',' || c == ':' || c == '-' {
					return fmt.Errorf("parse backticks-string failed [%s]", b.Raw)
				}
				btt.WriteByte(c)
			}

			p2 = p1
			p1 = c
		}

		if pi < len(b.Raw) {
			b.raws = append(b.raws, b.Raw[pi:])
		}
	}else{
		return fmt.Errorf("the string is not a backticks-string [%s]", b.Raw)
	}

	return nil
}

func (b *BackTickString) ConvertToJsString() (string, error) {

	if len(b.raws) == len(b.vars)+1 {
		var btt bytes.Buffer
		for i, s :=range b.raws {
			if i == 0 {
				if len(s) > 0 {
					if i < len(b.vars) {
						btt.WriteString(fmt.Sprintf("'%s'+%s", s, b.vars[i]))
					}else{
						btt.WriteString(fmt.Sprintf("'%s'", s))
					}
				}else{
					if i < len(b.vars) {
						btt.WriteString(fmt.Sprintf("%s", b.vars[i]))
					}
				}
			}else{
				if len(s) > 0 {
					if i < len(b.vars) {
						btt.WriteString(fmt.Sprintf("+'%s'+%s", s, b.vars[i]))
					}else{
						btt.WriteString(fmt.Sprintf("+'%s'", s))
					}
				}else{
					if i < len(b.vars) {
						btt.WriteString(fmt.Sprintf("+%s", b.vars[i]))
					}
				}
			}

		}

		return btt.String(), nil
	}

	return "", fmt.Errorf("backticks string convert to js failed")
}