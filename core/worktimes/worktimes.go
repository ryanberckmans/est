package worktimes

import (
	"errors"
	"fmt"
	"time"

	"github.com/rickar/cal"
)

type WorkTimes interface {
	GetWorkTimesOnDay(t time.Time) []time.Time
}

type workTimes struct {
	// calendar owns whether or not a given day is a workday
	calendar *cal.Calendar

	// workHours owns whether or not a given block of time during a work day is working hours
	workHours []time.Time // workHours must be even length with monotonically increasing time of day. This is enforced during construction. Only hour and minute of these times are defined: the hour and minute are used to construct specific workdays in GetWorkTimesOnDay().
	// TODO if we wanted to support different working hours for different days of the week, this is achievable by making workhours []string --> map[time.Weekday]string
	// TODO Calendar has a pretty nifty holidays interface: support holidays, vacation
}

// returns working start/end times on the day of the passed time, nil if passed time isn't a workday. Guaranteed that len([]time.Time) % 2 == 0, and that these times have monotonically increasing hour:minute in local time on day of passed time.
// TODO unit test
func (w *workTimes) GetWorkTimesOnDay(t time.Time) []time.Time {
	if !w.calendar.IsWorkday(t) {
		return nil
	}
	return getTimesOnDay(w.workHours, t)
}

// TODO doc, unit test, maybe rename
func getTimesOnDay(ts []time.Time, t time.Time) []time.Time {
	ts2 := make([]time.Time, len(ts))
	for i, t2 := range ts {
		ts2[i] = time.Date(t.Year(), t.Month(), t.Day(), t2.Hour(), t2.Minute(), 0, 0, t.Location())
	}
	return ts2
}

// TODO WorkTimes package
func New(workdays map[time.Weekday]bool, workhours []string) (WorkTimes, error) {
	ts, err := parseWorkHours(workhours)
	if err != nil {
		return nil, fmt.Errorf("new WorkTimes failed: %s", err.Error())
	}
	c := cal.NewCalendar()
	c.SetWorkday(time.Sunday, false)
	c.SetWorkday(time.Monday, false)
	c.SetWorkday(time.Tuesday, false)
	c.SetWorkday(time.Wednesday, false)
	c.SetWorkday(time.Thursday, false)
	c.SetWorkday(time.Friday, false)
	c.SetWorkday(time.Saturday, false)
	for workday, isWorkday := range workdays {
		c.SetWorkday(workday, isWorkday)
	}
	return &workTimes{calendar: c, workHours: ts}, nil
}

// TODO doc workhours hh:mm(am|pm), non-zero even number, monotonically increasing
// TODO unit test
func parseWorkHours(workhours []string) ([]time.Time, error) {
	if len(workhours) == 0 {
		return nil, errors.New("work hours was empty and must be non-zero even length and monotonically increasing")
	}
	if len(workhours)%2 == 1 {
		return nil, errors.New("work hours was odd length and must be non-zero even length and monotonically increasing")
	}

	ts := make([]time.Time, len(workhours))
	for i := range workhours {
		t, err := time.Parse("3:04pm", workhours[i])
		if err != nil {
			return nil, err
		}
		ts[i] = t
	}

	for i := 1; i < len(ts); i++ {
		if !ts[i-1].Before(ts[i]) {
			return nil, fmt.Errorf("work hours must be monotonically increasing, workhours[%d]==%s was not before workhours[%d]==%s", i-1, workhours[i-1], i, workhours[i])
		}
	}

	return ts, nil
}
