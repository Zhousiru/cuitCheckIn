package cuit

import (
	"regexp"

	"github.com/go-resty/resty/v2"
)

var (
	cuitLoginUrl = "http://login.cuit.edu.cn/Login/xLogin/Login.asp"
	cuitJkdkUrl  = "http://jszx-jxpt.cuit.edu.cn/jxgl/xs/netks/sj.asp?jkdk=Y"
)

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
	client.SetHeader("Referer", cuitLoginUrl)

	// req #0: get codeKey
	codeKey, err := getCodeKey(client)
	if err != nil {
		return nil, err
	}

	// req #1: post login form
	_, err = client.R().
		SetFormData(map[string]string{
			"WinW":        genWinW(),
			"winH":        genWinH(),
			"txtId":       id,
			"txtMM":       passwd,
			"verifycode":  "不区分大小写", // wtf
			"codeKey":     codeKey,
			"Login":       "Check",
			"IbtnEnter.x": "0",
			"IbtnEnter.y": "0",
		}).
		Post(cuitLoginUrl)
	if err != nil {
		return nil, err
	}

	// req #2: goto jkdk
	resp, err := client.R().Get(cuitJkdkUrl)
	if err != nil {
		return nil, err
	}

	r, _ := regexp.Compile(`(content="0;URL=)(.*?)(">)`)
	respStr := string(resp.Body())
	qqLoginUrl := r.FindStringSubmatch(respStr)[2]

	// req #3: goto qqLogin
	_, err = client.R().Get(qqLoginUrl)
	if err != nil {
		return nil, err
	}

	return client, nil
}
