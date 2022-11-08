package util

import (
	"bytes"
	"io"
	"net/url"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func DecodeGbk(gbkData []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(gbkData), simplifiedchinese.GBK.NewDecoder())

	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err

	}
	return string(b), nil
}

func EncodeGbkUrl(s string) (string, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewEncoder())

	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err

	}

	encoded := url.QueryEscape(string(b))

	return encoded, nil
}

func TrimNewline(str string) string {
	ret := strings.ReplaceAll(str, "\n", "")
	ret = strings.ReplaceAll(ret, "\r", "")

	return ret
}
