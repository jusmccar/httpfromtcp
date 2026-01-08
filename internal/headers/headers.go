package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

var crlf = []byte("\r\n")

var ErrorMalformedHeader = fmt.Errorf("Malformed header")
var ErrorInvalidFieldName = fmt.Errorf("Invalid field name")
var ErrorDuplicateFieldName = fmt.Errorf("Duplicate field name")

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	n = 0
	done = false
	err = nil

	if !bytes.Contains(data, crlf) {
		return n, done, nil
	}

	parts := bytes.Split(data, crlf)

	if len(parts) == 0 {
		return n, done, ErrorMalformedHeader
	}

	if bytes.Equal(parts[0], crlf) {
		done = true
		return n, done, nil
	}

	headerParts := bytes.Split(parts[0], []byte(" "))

	if len(headerParts) != 2 {
		return n, done, ErrorMalformedHeader
	}

	fieldNameColon := string(headerParts[0])
	fieldNameColonLastInd := len(fieldNameColon) - 1

	if fieldNameColon[fieldNameColonLastInd] != ':' {
		return n, done, ErrorMalformedHeader
	}

	fieldName := fieldNameColon[:fieldNameColonLastInd]

	if !isValidFieldName(fieldName) {
		return n, done, ErrorInvalidFieldName
	}

	if isDuplicateFieldName(fieldName, h) {
		return n, done, ErrorDuplicateFieldName
	}

	fieldNameLower := strings.ToLower(fieldName)
	fieldValue := string(headerParts[1])

	h[fieldNameLower] = fieldValue
	n = len(parts[0]) + len(crlf)

	return n, done, err
}

func isValidFieldName(s string) bool {
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '!' || c == '#' || c == '$' || c == '%' || c == '&' ||
			c == '\'' || c == '*' || c == '+' || c == '-' || c == '.' ||
			c == '^' || c == '_' || c == '`' || c == '|' || c == '~') {
			return false
		}
	}

	return true
}

func isDuplicateFieldName(s string, h Headers) bool {
	for k := range h {
		if strings.EqualFold(k, s) {
			return true
		}
	}

	return false
}
