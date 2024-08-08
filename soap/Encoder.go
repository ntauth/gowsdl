package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"strings"

	"golang.org/x/net/html/charset"
)

func cleanXMLReader(input io.Reader) (io.Reader, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}

	unescapedData := unescapeInvalidXMLChars(string(data))
	cleanData := replaceIllegalXMLChars(string(unescapedData))

	return bytes.NewBufferString(cleanData), nil
}

func replaceIllegalXMLChars(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if r == 0x9 || r == 0xA || r == 0xD ||
			(r >= 0x20 && r <= 0xD7FF) ||
			(r >= 0xE000 && r <= 0xFFFD) ||
			(r >= 0x10000 && r <= 0x10FFFF) {
			builder.WriteRune(r)
		} else {
			builder.WriteRune(' ')
		}
	}
	return builder.String()
}

func NewDecoder(data io.Reader) *xml.Decoder {
	decoder := xml.NewDecoder(data)
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false
	decoder.CharsetReader = func(charSet string, input io.Reader) (io.Reader, error) {
		utf8Reader, err := charset.NewReaderLabel(charSet, input)
		if err != nil {
			return nil, err
		}
		rawData, err := io.ReadAll(utf8Reader)
		if err != nil {
			return nil, fmt.Errorf("Unable to read data: %q", err)
		}
		filteredBytes := bytes.Map(filterValidXMLChar, rawData)
		return bytes.NewReader(filteredBytes), nil
	}

	return decoder
}

func filterValidXMLChar(r rune) rune {
	if r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF {
		return r
	}
	return -1
}

func unescapeInvalidXMLChars(s string) string {
	var builder strings.Builder
	i := 0
	for i < len(s) {
		if s[i] != '&' {
			builder.WriteByte(s[i])
			i++
			continue
		}

		j := i
		for j < len(s) && s[j] != ';' {
			j++
		}

		if j == len(s) || !isInvalidXMLCharRef(s[i:j+1]) {
			builder.WriteString(s[i : j+1])
			i = j + 1
			continue
		}

		unescapedChar := html.UnescapeString(s[i : j+1])
		builder.WriteString(unescapedChar)
		i = j + 1
	}
	return builder.String()
}

func isInvalidXMLCharRef(s string) bool {
	invalidXMLCharRefs := []string{"&#2;", "&#3;", "&#4;", "&#5;", "&#6;", "&#7;", "&#8;", "&#11;", "&#12;", "&#14;", "&#15;", "&#16;", "&#17;", "&#18;", "&#19;", "&#20;", "&#21;", "&#22;", "&#23;", "&#24;", "&#25;", "&#26;", "&#27;", "&#28;", "&#29;", "&#30;", "&#31;", "&#127;", "&#128;", "&#129;", "&#130;", "&#131;", "&#132;", "&#133;", "&#134;", "&#135;", "&#136;", "&#137;", "&#138;", "&#139;", "&#140;", "&#141;", "&#142;", "&#143;", "&#144;", "&#145;", "&#146;", "&#147;", "&#148;", "&#149;", "&#150;", "&#151;", "&#152;", "&#153;", "&#154;", "&#155;", "&#156;", "&#157;", "&#158;", "&#159;"}
	for _, ref := range invalidXMLCharRefs {
		if s == ref {
			return true
		}
	}
	return false
}
