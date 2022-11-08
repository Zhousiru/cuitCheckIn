package cuit

import (
	"errors"
	"regexp"

	"github.com/go-resty/resty/v2"
)

const (
	cuitLoginUrl = "http://login.cuit.edu.cn/Login/xLogin/Login.asp"
)

var ErrInvalidLoginCredential = errors.New("invalid login credential")

func getCodeKey(client *resty.Client) (string, error) {
	resp, err := client.R().Get(cuitLoginUrl)
	if err != nil {
		return "", err
	}

	r, _ := regexp.Compile(`(var codeKey = ')(.*?)(';)`)
	respStr := string(resp.Body())
	codeKey := r.FindStringSubmatch(respStr)[2]

	return codeKey, nil
}

func Login(id string, passwd string) (*resty.Client, error) {
	client := resty.New()

	// DEBUG
	// client.SetProxy("http://127.0.0.1:8888")
	// client.SetDebug(true)

	// req #0: get codeKey
	codeKey, err := getCodeKey(client)
	if err != nil {
		return nil, err
	}

	// req #1: post login form
	resp, err := client.R().SetHeader("Referer", cuitLoginUrl).
		SetFormData(map[string]string{
			"WinW":        genWinW(),
			"winH":        genWinH(),
			"txtId":       id,
			"txtMM":       passwd,
			"verifycode":  "", // not needed
			"codeKey":     codeKey,
			"Login":       "Check",
			"IbtnEnter.x": "0",
			"IbtnEnter.y": "0",
		}).
		Post(cuitLoginUrl)
	if err != nil {
		return nil, err
	}

	r, _ := regexp.Compile(`(content="0;URL=)(.*?)(">)`)

	// check login response
	sub := r.FindStringSubmatch(string(resp.Body()))
	if len(sub) == 0 {
		return nil, ErrInvalidLoginCredential
	}

	jkdkUrl := sub[2]

	// req #2: goto jkdk
	resp, err = client.R().Get(jkdkUrl)
	if err != nil {
		return nil, err
	}

	qqLoginUrl := r.FindStringSubmatch(string(resp.Body()))[2]

	// req #3: goto qqLogin
	_, err = client.R().Get(qqLoginUrl)
	if err != nil {
		return nil, err
	}

	return client, nil
}
