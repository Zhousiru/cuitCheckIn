package cuit

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type CheckInReq struct {
	LocationFlag int
	Address      struct {
		Province string
		City     string
		District string
	}
	PersonalWorkFlag   int
	PersonalHealthFlag int
	PersonalLiveFlag   int
	FamilyFlag         int
	Other              string
	PassportReq        *PassportReq
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
	StartDayFlag  int
	StartTimeFlag string
	EndDayFlag    int
	EndTimeFlag   string
}

var (
	ErrHourNotAvailable = errors.New("hour not available")
	ErrInvalidDate      = errors.New("invalid date")
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

	f.StartDayFlag = startDayFlag
	f.StartTimeFlag = startTimeFlag
	f.EndDayFlag = endDayFlag
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
