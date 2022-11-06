package cuit

import (
	"net/url"
	"regexp"
	"time"

	"github.com/Zhousiru/cuitCheckIn/internal/util"
	"github.com/go-resty/resty/v2"
)

const (
	cuitListUrl = "http://jszx-jxpt.cuit.edu.cn/Jxgl/Xs/netks/sj.asp"
)

type CheckIn struct {
	Title     string
	Date      time.Time
	IsChecked bool
	Url       string
}

func (c *CheckIn) FormatDate() string {
	return c.Date.Format("2006-01-02")
}

func reqList(client *resty.Client) (string, error) {
	resp, err := client.R().Get(cuitListUrl)
	if err != nil {
		return "", err
	}

	return util.DecodeGbk(resp.Body())
}

func dumpList(listBody string) ([]*CheckIn, error) {
	// preprocess
	s := util.TrimNewline(listBody)

	r := regexp.MustCompile(`(?i)middle;">(√|&nbsp;)</td><td align="left" style="vertical-align: middle;"><a href=(sjDb.*?) .*?>(.*?)<\/a>.*?middle;">(.*?) .*?<br>`)

	regexResult := r.FindAllStringSubmatch(s, -1)

	var checkInSlice []*CheckIn

	for _, el := range regexResult {
		// el[0]: full string
		// el[1]: check flag (true: "√", false: "&nbsp;")
		// el[2]: url(relative)
		// el[3]: title
		// el[4]: date(2022-11-05)

		checkIn := new(CheckIn)
		checkIn.Title = el[3]
		checkIn.Date, _ = time.Parse("2006-01-02", el[4])

		listUrl, _ := url.Parse(cuitListUrl)
		listUrl = listUrl.JoinPath("../")

		checkIn.Url = listUrl.String() + el[2]

		if el[1] == "√" {
			checkIn.IsChecked = true
		} else {
			checkIn.IsChecked = false
		}

		checkInSlice = append(checkInSlice, checkIn)
	}

	return checkInSlice, nil
}

func GetCheckInList(client *resty.Client) ([]*CheckIn, error) {
	s, err := reqList(client)
	if err != nil {
		return nil, err
	}

	return dumpList(s)
}
