package textsearch

import (
	"bytes"
	"strings"
)

var _splitdict_ []string = []string{" is ", " was ", " does ", " do ", " has ", " have ", " can ", " should ", " could ", "，", "。", "、", "——", "‘", "”", "】", "【", "《", "》", "？", "（", "）", "·", "=", ",", ".", " ", "+", "-", "_", "{", "}", "[", "]", "|", "\"", "(", ")"}

type WordsSpliter struct {
}

func (w *WordsSpliter) Split(text string) []string {
	var words []string
	var bts bytes.Buffer
	for len(text) > 0 {
		if idx := w.tag(text); idx > 0 {
			text = text[idx:]
			if bts.Len() > 0 {
				words = append(words, bts.String())
				bts.Reset()
			}
		} else if len(text) > 1 && !w.sametype(text[0], text[1]) {
			bts.WriteByte(text[0])
			words = append(words, bts.String())
			bts.Reset()
			text = text[1:]
		} else {
			bts.WriteByte(text[0])
			text = text[1:]
		}
	}
	if bts.Len() > 0 {
		words = append(words, bts.String())
	}

	return words
}

func (w *WordsSpliter) tag(text string) int {
	for i, e := 0, len(_splitdict_); i < e; i++ {
		if strings.HasPrefix(text, _splitdict_[i]) {
			return len(_splitdict_[i])
		}
	}
	return 0
}

func (w *WordsSpliter) sametype(a, b byte) bool {
	ia := (a >= 'a' && a <= 'z') || (a >= 'A' && a <= 'Z') || (a >= '0' && a <= '9')
	ib := (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')

	return ia == ib
}
