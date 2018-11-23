/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package utils

import (
	"time"
)

const (
	Nanosecond  = int64(time.Nanosecond) // Nanosecond
	Microsecond = 1000 * Nanosecond      // Nanosecond
	Millisecond = 1000 * Microsecond     // Nanosecond
	Second      = 1000 * Millisecond     // Nanosecond
	Minute      = 60 * Second            // Nanosecond
	Hour        = 60 * Minute            // Nanosecond
	Day         = 24 * Hour              // Nanosecond
	Year        = 365 * Day              // Nanosecond

	SecondsPerYear   = Year / Second   // Second
	SecondsPerDay    = Day / Second    // Second
	SecondsPerHour   = Hour / Second   // Second
	SecondsPerMinute = Minute / Second // Second

	MillisecondsPerYear   = Year / Millisecond   // Millisecond
	MillisecondsPerDay    = Day / Millisecond    // Millisecond
	MillisecondsPerHour   = Hour / Millisecond   // Millisecond
	MillisecondsPerMinute = Minute / Millisecond // Millisecond
	MillisecondsPerSecond = Second / Millisecond // Millisecond

	MicrosecondsPerYear        = Year / Microsecond        // Microsecond
	MicrosecondsPerDay         = Day / Microsecond         // Microsecond
	MicrosecondsPerHour        = Hour / Microsecond        // Microsecond
	MicrosecondsPerMinute      = Minute / Microsecond      // Microsecond
	MicrosecondsPerSecond      = Second / Microsecond      // Microsecond
	MicrosecondsPerMillisecond = Millisecond / Microsecond // Microsecond
)

const timeLayout = "2006-01-02 15:04:05"

func GetCurrentSecond() int64 {
	return time.Now().Unix()
}

func GetCurrentMillisecond() int64 {
	return time.Now().UnixNano() / Millisecond
}

func GetCurrentMicrosecond() int64 {
	return time.Now().UnixNano() / Microsecond
}

func GetCurrentNanosecond() int64 {
	return time.Now().UnixNano()
}

func FromSecondToLocalDate(v int64) string {
	return time.Unix(v, 0).Format(timeLayout)
}

func GetCurrentTime() string {
	return time.Now().Format(timeLayout)
}

func FromLocalDateToSecond(v string) int64 {
	loc, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation(timeLayout, v, loc)
	return t.Unix()
}

func IsSameDay(t1, t2 time.Time) bool {
	year1, month1, day1 := t1.Date()
	year2, month2, day2 := t2.Date()
	if year1 == year2 && month1 == month2 && day1 == day2 {
		return true
	}
	return false
}

func IsSameDayWithTimestamp(d1, d2 int64) bool {
	t1 := time.Unix(d1, 0)
	t2 := time.Unix(d2, 0)
	return IsSameDay(t1, t2)
}

const (
	offset_day   = 1
	offset_monty = 100 * offset_day
	offset_year  = 100 * offset_monty
)

func FromTime2TimeInt(t time.Time) int {
	y, m, d := t.Date()
	return y*offset_year + int(m)*offset_monty + d*offset_day
}

func SplitTimeInt(t int) (year, month, day int) {
	return t / offset_year, t % offset_year / offset_monty, t % offset_monty
}

func DGetCurrentTime(layout string) string {
	return time.Now().Format(layout)
}

func Sleep(millisecond int64) {
	time.Sleep(time.Duration(millisecond) * time.Millisecond)
}

func Duration(d int64) time.Duration {
	return time.Duration(d)
}