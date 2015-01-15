package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const MIN_STOP_DISTANCE_PIPS = 10.0 // minimum distance is 10 pips from open
const MIN_STOP_DISTANCE      = MIN_STOP_DISTANCE_PIPS / 100000 // i.e., 0.0001
const MIN_STOP_DISTANCE_JPY  = MIN_STOP_DISTANCE      / 1000  // i.e., 0.01

const BASE_PIP_DIVISOR = 10000 // i.e., 2.0 -> 0.0002
const JPY_PIP_DIVISOR  = 100   // i.e., 2.0 -> 0.02

const Day = time.Hour * 24

// ===== MISC FUNCTIONS ============================================================================

func Average(f []float64) float64 {
	l := float64(len(f))
	sum := 0.0

	for _, val := range f {
		sum += val
	}

	return sum / l
}

func EnsureZeroOrGreater(f float64) {
	if 0.0 > f {
		panic("number is not >= 0")
	}
}

func ReverseString(s string) string {
	n := len(s)
	runes := make([]rune, n)
	for _, rune := range s {
		n--
		runes[n] = rune
	}

	return string(runes[n:])
}

func AddCommas(val int64) string {
	// call the cops, i don't care
	return addCommasToIntString(fmt.Sprintf("%d", val))
}

func addCommasToIntString(s string) string {
	s = ReverseString(s)
	sl := len(s) - 1
	lastIndex := 0

	commaArray := []string{}

	for i := 0; i <= sl; i++ {
		if i == sl {
			commaArray = append(commaArray, s[lastIndex:(i + 1)])
			break
		}

		if lastIndex + 2 == i  {
			commaArray = append(commaArray, s[lastIndex:(lastIndex + 3)])
			lastIndex = i + 1
		}
	}

	return ReverseString(strings.Join(commaArray, ","))
}

func FormatMoney(shekels float64) string {
	parts := strings.Split(fmt.Sprintf("%.2f", math.Abs(shekels)), ".")
	dollars := parts[0]
	decimal := parts[1]

	str := "$" + addCommasToIntString(dollars) + "." + decimal

	if shekels >= 0.0 {
		return str
	} else {
		return "-" + str
	}
}

func MarshalOrDie(i interface{}) []byte {
	bytes, err := json.Marshal(i)
	if err != nil {
		panic(fmt.Sprintf(
			"MarshalOrDie: failed to marshal. Error: %s - Object: %#v\n",
			err,
			i,
		))
	}

	return bytes
}

func ParsePGTime(s string) time.Time {
	// dates from PG look like: 2014-10-30 07:47:42.190842-07
	// convert them to RFC3339 format for easier parsing
	fixed_date := strings.Replace(s, " ", "T", -1) + ":00"

	t, err := time.Parse(time.RFC3339Nano, fixed_date)
	if err != nil {
		panic("time.Parse: " + err.Error())
	}

	return t
}

func StringArrayContainsString(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}

	return false
}

func StringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("strconv.ParseFloat: " + err.Error())
	}

	return f
}

func FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func StringToInt(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic("strconv.ParseInt: " + err.Error())
	}

	return i
}

func PercentChange(f1, f2 float64) float64 {
	return (f2 - f1) / f1
}

func NiceTimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func NiceYearMonthFormat(t time.Time) string {
	return t.Format("2006-01")
}
