package cuit

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Zhousiru/cuitCheckIn/internal/util"
	"github.com/go-resty/resty/v2"
)

type Passport struct {
	Dest     string
	Reason   string
	Approved bool
	Range    struct {
		Start time.Time
		End   time.Time
	}
}

func reqDetail(client *resty.Client, checkInUrl string) (string, error) {
	resp, err := client.R().Get(checkInUrl)
	if err != nil {
		return "", err
	}

	return util.DecodeGbk(resp.Body())
}

func dumpPassport(detailBody string, relatedCheckIn *CheckIn) (*Passport, error) {
	// preprocess
	s := util.TrimNewline(detailBody)

	r := regexp.MustCompile(`(?i)name=sF21912_1 value="(.*?)".*?name=sF21912_2 value="(.*?)".*?selected value="([0-9])".*?selected value="([0-9]{2})".*?selected value="([0-9])".*?selected value="([0-9]{2})".*?<span style="color:#0000FF">(.*?)</span>`)

	sub := r.FindStringSubmatch(s)

	passport := new(Passport)

	// sub[0]: full string
	// sub[1]: destination
	// sub[2]: reason
	// sub[3]: start day
	//         1: check-in date, 2: +1d, 3: +2d
	// sub[4]: start time(XX: XX:00)
	// sub[5]: end day
	//         1: start date, 2: +1d, 3: +2d, 9: night of start date?
	// sub[6]: end time(XX: XX:00)
	// sub[7]: approval result("已通过..." or "未通过...")

	passport.Dest = sub[1]
	passport.Reason = sub[2]

	if strings.HasPrefix(sub[7], "已通过") {
		passport.Approved = true
	} else {
		passport.Approved = false
	}

	checkInDate := relatedCheckIn.Date

	flagStartDay, _ := strconv.Atoi(sub[3])
	flagStartHour, _ := strconv.Atoi(sub[4])
	startDate := checkInDate.AddDate(0, 0, flagStartDay-1)               // like 2022-11-06
	startTime := startDate.Add(time.Hour * time.Duration(flagStartHour)) // like 2022-11-06 12:00

	passport.Range.Start = startTime

	flagEndDay, _ := strconv.Atoi(sub[5])

	if flagEndDay == 9 {
		// night of start date
		flagEndDay = 1 // maybe it euqal to `start date`?
	}

	flagEndHour, _ := strconv.Atoi(sub[6])
	endDate := startDate.AddDate(0, 0, flagEndDay-1)               // like 2022-11-08
	endTime := endDate.Add(time.Hour * time.Duration(flagEndHour)) // like 2022-11-08 12:00

	passport.Range.End = endTime

	return passport, nil
}
