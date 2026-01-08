package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var crlf = []byte("\r\n")

var (
	ErrorMalformedHeader  = fmt.Errorf("Malformed header")
	ErrorInvalidFieldName = fmt.Errorf("Invalid field name")
)

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
	header := parts[0]

	if len(header) == 0 {
		done = true
		return n, done, nil
	}

	fieldNameBytes, fieldValueBytes, found := bytes.Cut(header, []byte(":"))

	if !found {
		return n, done, ErrorMalformedHeader
	}

	if bytes.HasSuffix(fieldNameBytes, []byte(" ")) {
		return n, done, ErrorInvalidFieldName
	}

	fieldNameBytes = bytes.TrimSpace(fieldNameBytes)
	fieldValueBytes = bytes.TrimSpace(fieldValueBytes)

	if !isValidFieldName(fieldNameBytes) {
		return n, done, ErrorInvalidFieldName
	}

	fieldName := string(bytes.ToLower(fieldNameBytes))
	fieldValue := string(fieldValueBytes)

	_, exists := h[fieldName]

	if exists {
		h[fieldName] += ", " + fieldValue
	} else {
		h[fieldName] = fieldValue
	}

	n = len(header) + len(crlf)

	return n, done, err
}

func isValidFieldName(b []byte) bool {
	for _, c := range b {
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
