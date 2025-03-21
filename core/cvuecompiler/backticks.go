package cvuecompiler

import (
    "fmt"
    "strings"
)

type BackTickString struct {
    Raw  string   
    raws []string 
    vars []string 
}

func (b *BackTickString) Init(backticks string) {
    b.Raw = backticks
    b.raws = nil
    b.vars = nil
}

func (b *BackTickString) Parse() error {
    if s, l := 0, len(b.Raw)-1; s < l && b.Raw[s] == '`' && b.Raw[l] == '`' {
        b.Raw = b.Raw[1:l]
        var builder strings.Builder
        p1, p2, pi, open := byte(' '), byte(' '), 0, false
        for i, c := range []byte(b.Raw) {
            if c == '}' && open {
                pi = i + 1
                open = false
                if i == len(b.Raw)-1 {
                    b.raws = append(b.raws, "")
                }
                if builder.Len() > 0 {
                    b.vars = append(b.vars, builder.String())
                    builder.Reset()
                } else {
                    return fmt.Errorf("parse backticks-string failed [%s]", b.Raw)
                }
            } else if c == '{' && p1 == '$' && p2 != '\\' {
                b.raws = append(b.raws, b.Raw[pi:i-1])
                open = true
            } else if open {
                if c == '{' || c == ',' || c == ':' || c == '-' {
                    return fmt.Errorf("parse backticks-string failed [%s]", b.Raw)
                }
                builder.WriteByte(c)
            }

            p2 = p1
            p1 = c
        }

        if pi < len(b.Raw) {
            b.raws = append(b.raws, b.Raw[pi:])
        }
    } else {
        return fmt.Errorf("the string is not a backticks-string [%s]", b.Raw)
    }

    return nil
}

func (b *BackTickString) ConvertToJsString() (string, error) {
    if len(b.raws) == len(b.vars)+1 {
        var builder strings.Builder
        for i, s := range b.raws {
            if i == 0 {
                if len(s) > 0 {
                    builder.WriteByte('\'')
                    builder.WriteString(s)
                    builder.WriteByte('\'')
                    
                    if i < len(b.vars) {
                        builder.WriteByte('+')
                        builder.WriteString(b.vars[i])
                    }
                } else if i < len(b.vars) {
                    builder.WriteString(b.vars[i])
                }
            } else {
                if len(s) > 0 {
                    builder.WriteByte('+')
					builder.WriteByte('\'')
                    builder.WriteString(s)
                    builder.WriteByte('\'')
                    
                    if i < len(b.vars) {
                        builder.WriteByte('+')
                        builder.WriteString(b.vars[i])
                    }
                } else if i < len(b.vars) {
                    builder.WriteByte('+')
                    builder.WriteString(b.vars[i])
                }
            }
        }

        return builder.String(), nil
    }
    return "", fmt.Errorf("backticks string convert to js failed")
}