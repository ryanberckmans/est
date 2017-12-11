package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

var flagAgo string

func applyFlagAgo(t time.Time) time.Time {
	if flagAgo == "" {
		return t
	}
	ago, err := parseDurationHours(flagAgo, "duration ago")
	if err != nil {
		fmt.Printf("fatal: %v\n", err)
		os.Exit(1)
		return time.Time{}
	}
	return t.Add(-ago)
}

var durationHoursRegexp = regexp.MustCompile(`^([1-9][0-9]*(\.[0-9]*)?|0\.[0-9]+)(m|h|d|w)?$`)

// TODO unit test
func parseDurationHours(e string, name string) (time.Duration, error) {
	if e == "" {
		return 0, nil
	}
	if !durationHoursRegexp.MatchString(e) {
		return 0, errors.New("invalid " + name)
	}
	unitMultiplier := 1.0 // default to hours
	var eWithoutUnit string
	switch e[len(e)-1:] {
	case "m":
		eWithoutUnit = e[:len(e)-1]
		unitMultiplier = 1 / 60.0 // 1/60 hours in a minute
	case "h":
		eWithoutUnit = e[:len(e)-1]
	case "d":
		eWithoutUnit = e[:len(e)-1]
		unitMultiplier = 8 // 8 hours in a day
	case "w":
		eWithoutUnit = e[:len(e)-1]
		unitMultiplier = 8 * 5 // 40 hours in a week
	default:
		eWithoutUnit = e
	}

	f, err := strconv.ParseFloat(eWithoutUnit, 64)
	if err != nil {
		return 0, errors.New("estimate wasn't a float")
	}
	d, err := time.ParseDuration(fmt.Sprintf("%fh", f*unitMultiplier))
	if err != nil {
		panic(err)
	}
	return d, nil
}
