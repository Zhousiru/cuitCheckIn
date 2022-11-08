package cuit

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Zhousiru/cuitCheckIn/internal/util"
	"github.com/go-resty/resty/v2"
)

type CheckInReq struct {
	// LocationFlag int
	// Address      struct {
	// 	Province string
	// 	City     string
	// 	District string
	// }
	// PersonalWorkFlag   int
	// PersonalHealthFlag int
	// PersonalLiveFlag   int
	// FamilyFlag         int
	// Other              string
	PassportReq *PassportReq
}

type PassportReq struct {
	Dest   string
	Reason string
	Range  struct {
		Start time.Time
		End   time.Time
	}
}

type passportRangeFlag struct {
	StartDayFlag  string
	StartTimeFlag string
	EndDayFlag    string
	EndTimeFlag   string
}

var (
	ErrHourNotAvailable = errors.New("hour not available")
	ErrInvalidDate      = errors.New("invalid date")
)

const (
	cuitCheckInPostUrl = "http://jszx-jxpt.cuit.edu.cn/Jxgl/Xs/netks/editSjRs.asp"
	postFormTemplate   = `RsNum=3&Tx=33_1&canTj=1&isNeedAns=0&UTp=Xs&th_1=21650&wtOR_1=1%5C%7C%2F%CB%C4%B4%A8%5C%7C%2F%B3%C9%B6%BC%5C%7C%2F%CB%AB%C1%F7%5C%7C%2F1%5C%7C%2F%5C%7C%2F%5C%7C%2F%5C%7C%2F%5C%7C%2F&sF21650_1=1&sF21650_2=%CB%C4%B4%A8&sF21650_3=%B3%C9%B6%BC&sF21650_4=%CB%AB%C1%F7&sF21650_5=1&sF21650_6=&sF21650_7=&sF21650_8=&sF21650_9=&sF21650_10=&sF21650_N=10&th_2=21912&wtOR_2=&sF21912_1={{dest}}&sF21912_2={{reason}}&sF21912_3={{startDayFlag}}&sF21912_4={{startTimeFlag}}&sF21912_5={{endDayFlag}}&sF21912_6={{endTimeFlag}}&sF21912_N=6&th_3=21648&wtOR_3=N%5C%7C%2F%5C%7C%2FN%5C%7C%2F%5C%7C%2FN%5C%7C%2F&sF21648_1=N&sF21648_2=&sF21648_3=N&sF21648_4=&sF21648_5=N&sF21648_6=&sF21648_N=6&zw1=&cxStYt=A&zw2=&B2=%CC%E1%BD%BB%B4%F2%BF%A8`
)

func (f *passportRangeFlag) From(passportReq *PassportReq, relatedCheckIn *CheckIn) error {
	checkInDate := relatedCheckIn.Date

	if !checkInDate.Before(passportReq.Range.Start) || !passportReq.Range.Start.Before(passportReq.Range.End) {
		return ErrInvalidDate
	}

	startDur := passportReq.Range.Start.Sub(checkInDate)
	startDay, startHour := getDayHour(startDur)
	endDur := passportReq.Range.End.Sub(trimTime(passportReq.Range.Start)) // `trimTime() removes hour, min...`
	endDay, endHour := getDayHour(endDur)

	// check available hours
	// start: 6-22 end: 7-23
	if startHour < 6 || startHour > 22 || endHour < 7 || endHour > 23 {
		return ErrHourNotAvailable
	}

	startDayFlag := startDay + 1
	endDayFlag := endDay + 1

	if startDayFlag > 3 || endDayFlag > 3 {
		return ErrInvalidDate
	}

	startTimeFlag := fmt.Sprintf("%02d", startHour)
	endTimeFlag := fmt.Sprintf("%02d", endHour)

	f.StartDayFlag = strconv.Itoa(startDayFlag)
	f.StartTimeFlag = startTimeFlag
	f.EndDayFlag = strconv.Itoa(endDayFlag)
	f.EndTimeFlag = endTimeFlag

	return nil
}

func getDayHour(d time.Duration) (int, int) {
	day := int(math.Floor(d.Hours() / 24))
	hour := int(d.Hours()) % 24

	return day, hour
}

func trimTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func PostCheckInReq(client *resty.Client, checkIn *CheckIn, req *CheckInReq) error {
	flag := new(passportRangeFlag)
	flag.From(req.PassportReq, checkIn)

	dest, _ := util.EncodeGbkUrl(req.PassportReq.Dest)
	reason, _ := util.EncodeGbkUrl(req.PassportReq.Reason)

	r := strings.NewReplacer("{{dest}}", dest, "{{reason}}", reason, "{{startDayFlag}}", flag.StartDayFlag, "{{startTimeFlag}}", flag.StartTimeFlag, "{{endDayFlag}}", flag.EndDayFlag, "{{endTimeFlag}}", flag.EndTimeFlag)

	checkInUrl, _ := url.Parse(checkIn.Url)
	checkInQuery := checkInUrl.Query()

	formUrl, _ := url.ParseQuery(r.Replace(postFormTemplate))
	formUrl.Add("Id", checkInQuery.Get("Id"))
	formUrl.Add("ObjId", checkInQuery.Get("ObjId"))

	resp, err := client.R().SetHeader("Referer", checkIn.Url).
		SetFormDataFromValues(formUrl).
		Post(cuitCheckInPostUrl)
	if err != nil {
		return err
	}

	fmt.Println(util.DecodeGbk(resp.Body()))

	return nil
}
