package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("Malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported HTTP version")

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, errors.Join(fmt.Errorf("Unable to io.ReadAll"), err)
	}

	str := string(data)
	requestLineStruct, err := parseRequestLine(str)

	if err != nil {
		return nil, err
	}

	requestStruct := &Request{
		RequestLine: *requestLineStruct,
	}

	return requestStruct, err
}

func parseRequestLine(str string) (*RequestLine, error) {
	lines := strings.Split(str, "\r\n")

	if len(lines) == 0 {
		return nil, ERROR_MALFORMED_REQUEST_LINE
	}

	requestLine := lines[0]
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return nil, ERROR_MALFORMED_REQUEST_LINE
	}

	method := parts[0]
	requestTarget := parts[1]
	httpParts := strings.Split(parts[2], "/")

	if len(httpParts) != 2 {
		return nil, ERROR_MALFORMED_REQUEST_LINE
	}

	httpVersion := httpParts[1]

	if httpVersion != "1.1" {
		return nil, ERROR_UNSUPPORTED_HTTP_VERSION
	}

	requestLineStruct := &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpVersion,
	}

	return requestLineStruct, nil
}
