package cuit

import (
	"path"
	"regexp"
	"strings"

	"github.com/Zhousiru/cuitCheckIn/internal/util"
	"github.com/go-resty/resty/v2"
)

const (
	cuitListUrl = "http://jszx-jxpt.cuit.edu.cn/Jxgl/Xs/netks/sj.asp"
)

type CheckIn struct {
	Title     string
	Date      string
	IsChecked bool
	Url       string
}

func reqList(client *resty.Client) (string, error) {
	resp, err := client.R().Get(cuitListUrl)
	if err != nil {
		return "", err
	}

	return util.DecodeGbk(resp.Body())
}

func dumpList(listBody string) ([]*CheckIn, error) {
	r, _ := regexp.Compile(`middle;">(.*?)</td>.*?<a href="(sjDb.*?)".*?>(.*?)<\/a>.*?middle;">(.*?) .*?<br>`)

	regexResult := r.FindAllStringSubmatch(strings.ReplaceAll(listBody, "\n", ""), -1)

	var checkInSlice []*CheckIn

	for _, el := range regexResult {
		// el[0]: full string
		// el[1]: check flag (true: "√", false: "&nbsp;")
		// el[2]: url(relative)
		// el[3]: title
		// el[4]: date(2022-11-05)

		checkIn := new(CheckIn)
		checkIn.Title = el[3]
		checkIn.Date = el[4]
		checkIn.Url = path.Join(path.Dir(cuitListUrl), el[2])
		if el[1] == "√" {
			checkIn.IsChecked = true
		} else {
			checkIn.IsChecked = false
		}

		checkInSlice = append(checkInSlice, checkIn)
	}

	return checkInSlice, nil
}
