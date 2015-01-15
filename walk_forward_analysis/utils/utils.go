package utils

import (
	"fmt"
	"strconv"
	"time"
)

const (
	Day = time.Hour * 24
	Week = Day * 7
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func MergeDateTime(part1, part2 string) time.Time {
	str := part1 + " " + part2

	t, err := time.Parse("2006-01-02 15:04:05", str)
	// t, err := time.Parse("01/02/2006 15:04:05", str)
	if err != nil {
		panic(fmt.Sprintf("time.Parse failed on \"%s\": %s\n", str, err.Error()))
	}

	return t
}

func ParseTime(str string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", str)
	// t, err := time.Parse("01/02/2006 15:04:05", str)
	if err != nil {
		panic(fmt.Sprintf("time.Parse failed on \"%#v\": %s\n", str, err.Error()))
	}

	return t
}

func StringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("strconv.ParseFloat: " + err.Error())
	}

	return f
}

func StringToInt(s string) int {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic("strconv.ParseInt: " + err.Error())
	}

	return int(i)
}
