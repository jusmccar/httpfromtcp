package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var crlf = []byte("\r\n")

var ErrorMalformedHeader = fmt.Errorf("Malformed header")

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
	fieldValue := string(headerParts[1])

	h[fieldName] = fieldValue
	n = len(parts[0]) + len(crlf)

	return n, done, err
}
