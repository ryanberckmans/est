package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/ryanberckmans/est/core"
)

var flagLog string      // duration of logged time e.g. "30m"
var flagEstimate string // duration estimate e.g. "2.5h"
var flagAgo string      // duration ago e.g. "0.5d"

func doFlagLog(t *core.Task, now time.Time) {
	if flagLog != "" && flagAgo != "" {
		// --log and --ago may not co-occur because this creates weird auto time tracking issues which, while logically consistent, are probably really confusing to users.
		fmt.Print("fatal: --log may not be used with --ago\n")
		os.Exit(1)
		return
	}
	if flagLog == "" {
		return
	}
	d, err := parseDurationHours(flagLog, "log duration")
	if err != nil {
		fmt.Printf("fatal: %v\n", err)
		os.Exit(1)
		return
	}
	if !t.IsStarted() && !t.IsPaused() {
		fmt.Print("fatal: cannot log time on a task which isn't started or paused\n")
		os.Exit(1)
		return
	}
	if err := t.AddActual(d, now); err != nil {
		panic(err)
	}
}

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

var durationRegexp = regexp.MustCompile(`^([1-9][0-9]*(\.[0-9]*)?|0\.[0-9]+)(m|h)$`)

// TODO unit test
func parseDurationHours(e string, name string) (time.Duration, error) {
	if e == "" {
		return 0, nil
	}
	if !durationRegexp.MatchString(e) {
		return 0, errors.New("invalid " + name + ". For example, \"1.5h\", \"0.5h\", or \"90m\".")
	}
	unitMultiplier := 1.0 // default to hours
	var eWithoutUnit string
	switch e[len(e)-1:] {
	case "m":
		eWithoutUnit = e[:len(e)-1]
		unitMultiplier = 1 / 60.0 // 1/60 hours in a minute
	case "h":
		eWithoutUnit = e[:len(e)-1]
	default:
		panic("expected estimate to end in 'm' or 'h' due to durationHoursRegexp")
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
