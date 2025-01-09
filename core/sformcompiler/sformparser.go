package sformcompiler

import (
	"bytes"
	"fmt"
	"strings"
)

type SformValueTransfomer struct {
	fieldNames        []string
	fieldTypes        []string
	fieldHasValueLine []bool
	fieldValueLineIdx []int
	source            []byte
}

func (s SformValueTransfomer) GetFields() []string {
	return s.fieldNames
}

func (s *SformValueTransfomer) WriteKVs(kvs [][2]string) (string, error) {
	var bts bytes.Buffer
	pidx := 0
	for i, k := range s.fieldNames {
		v := ""
		for m, e := 0, len(kvs); m < e; m++ {
			if kvs[m][0] == k {
				v = kvs[m][1]
				break
			}
		}
		cidx := s.fieldValueLineIdx[i]

		if _, err := bts.Write(s.source[pidx:cidx]); err != nil {
			return "", err
		}
		if _, err := bts.WriteString(fmt.Sprintf("    value %s\n", strings.Trim(v, "\n"))); err != nil {
			return "", err
		}

		if !s.fieldHasValueLine[i] {
			pidx = cidx
		} else {
			for i, e := cidx, len(s.source); i < e; i++ {
				if s.source[i] == '\n' {
					pidx = i + 1
					break
				}
			}
		}
	}

	if _, err := bts.Write(s.source[pidx:]); err != nil {
		return "", err
	}

	return bts.String(), nil
}

func (s *SformValueTransfomer) LoadForm(sform []byte) error {
	var token bytes.Buffer
	open, linestart := false, 0
	var fields []string
	name, typ, nameidx, valueidx, squot, dquot := "", "", 0, 0, true, true

	for i, c := range sform {
		if c == '\n' && squot && dquot {
			if token.Len() > 0 {
				fields = append(fields, token.String())
			}
			if len(fields) == 1 && fields[0] == "field" {
				open = true
				name, typ, nameidx, valueidx = "", "", 0, 0
			}
			if len(fields) == 2 && fields[1] == "field" && fields[0] == "end" {
				open = false
				if len(name) > 0 && len(typ) > 0 {
					s.fieldNames = append(s.fieldNames, name)
					s.fieldTypes = append(s.fieldTypes, typ)
					if valueidx == 0 {
						s.fieldHasValueLine = append(s.fieldHasValueLine, false)
						s.fieldValueLineIdx = append(s.fieldValueLineIdx, nameidx)
					} else {
						s.fieldHasValueLine = append(s.fieldHasValueLine, true)
						s.fieldValueLineIdx = append(s.fieldValueLineIdx, valueidx)
					}
				}
			}
			if open && len(fields) >= 2 {
				if fields[0] == "name" {
					name = fields[1]
					nameidx = linestart
				}
				if fields[0] == "value" {
					valueidx = linestart
				}
				if fields[0] == "type" {
					typ = fields[1]
				}
			}

			linestart = i + 1
			token.Reset()
			fields = fields[:0]
			continue
		}

		if c == '\'' {
			squot = !squot
		}
		if c == '"' {
			dquot = !dquot
		}

		if c == ' ' && token.Len() > 0 {
			fields = append(fields, token.String())
			token.Reset()
		}

		if c != ' ' {
			token.WriteByte(c)
		}
	}

	if len(s.fieldNames) == 0 {
		return fmt.Errorf("failed to parse sform")
	}
	s.source = sform

	return nil
}
