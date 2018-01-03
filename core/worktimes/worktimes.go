package worktimes

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/rickar/cal"
)

/*
	TODO
		unit tests
		remove commented out debug printfs
		public docs
		this pkg could be its own repo or PR'd into rickar/cal
*/

type WorkTimes interface {
	GetWorkTimesOnDay(day time.Time) []time.Time
	DurationBetween(start, end time.Time) time.Duration
	TimeAfter(start time.Time, d time.Duration) time.Time
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
// TODO real doc and unit test
func (wt *workTimes) GetWorkTimesOnDay(t time.Time) []time.Time {
	if !wt.calendar.IsWorkday(t) {
		return nil
	}
	return getTimesOnDay(wt.workHours, t)
}

// TODO doc, finish unit tests.
func (wt *workTimes) DurationBetween(start, end time.Time) time.Duration {
	if end.Before(start) {
		return wt.DurationBetween(end, start)
	}
	// Business hours are relative to a specific timezone; we assume local time.
	if start.Location() != time.Local {
		return wt.DurationBetween(start.Local(), end)
	}
	if end.Location() != time.Local {
		return wt.DurationBetween(start, end.Local())
	}
	// fmt.Printf("DurationBetween start=%v end=%v\n", start, end)
	d := time.Duration(0) // accumulated business hours between start and end
	s2 := start
	for {
		if s2.Year() == end.Year() && s2.YearDay() == end.YearDay() {
			// fmt.Printf("on same day s2=%v end=%v\n", s2, end)
			return d + durationBetweenOnSameDay(wt, s2, end)
		}
		// fmt.Printf("not on same day same day s2=%v end=%v\n", s2, end)
		d += durationBetweenOnSameDay(wt, s2, endOfDay(s2))
		s2 = startOfDay(s2.AddDate(0, 0, 1))
	}
}

// TODO doc, unit test
func (wt *workTimes) TimeAfter(start time.Time, d time.Duration) time.Time {
	// TODO this function is the dual of DurationBetween(); do an explicit impl of that dual, supporting negative durations. For now we'll just do a binary search approximation using DurationBetween().
	if d < 0 {
		panic(fmt.Sprintf("negative duration unsupported: %v", d))
	}
	// Business hours are relative to a specific timezone; we assume local time.
	if start.Location() != time.Local {
		return wt.TimeAfter(start.Local(), d)
	}
	low := start
	high := start.Add(time.Hour * 24 * 365 * 100) // 100 years in future; algorithm will never terminate if true result is more than 100 years in future.
	for {
		test := time.Unix((low.Unix()+high.Unix())/2, 0) // halfway between low and high
		d2 := wt.DurationBetween(start, test)
		// fmt.Printf("i=%d test=%v d=%v start=%v end=%v d2=%v\n", i, test, d, start, end, d2)
		if time.Duration(int(math.Abs(float64(d-d2)))) < time.Minute {
			// approximation is within one minute of actual value, good enough
			// fmt.Printf("i=%d test=%v d=%d start=%v finished\n", i, test, d, start)
			return test
		}
		if d2 > d {
			high = test
		} else {
			low = test
		}
	}
}

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

// TODO doc, unit test, maybe rename
func getTimesOnDay(ts []time.Time, t time.Time) []time.Time {
	ts2 := make([]time.Time, len(ts))
	for i, t2 := range ts {
		ts2[i] = time.Date(t.Year(), t.Month(), t.Day(), t2.Hour(), t2.Minute(), 0, 0, t.Location())
	}
	return ts2
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

func durationBetweenOnSameDay(wt WorkTimes, start, end time.Time) time.Duration {
	if end.Before(start) {
		return durationBetweenOnSameDay(wt, end, start)
	}
	if !(start.Year() == end.Year() && start.YearDay() == end.YearDay()) {
		panic("expected start and end on same day")
	}
	ts := wt.GetWorkTimesOnDay(start)
	// fmt.Printf("OnSameDay start=%v end=%v wts=%+v\n", start, end, ts)
	if len(ts) < 1 {
		// passed day is not a workday
		return 0
	}
	d := time.Duration(0)
	s2 := start
	for i := 0; i*2+1 < len(ts); i++ {
		if end.Before(ts[i*2]) {
			// end is prior to the start of this working time block,
			// no more working hours can be accumulated today.
			// fmt.Printf("OnSameDay s2=%v workstart=%v workend=%v duration=%v\n", s2, ts[i*2], ts[i*2+i], d)
			break
		}
		if s2.After(ts[i*2+1]) {
			// s2 is after to the end of this working time block, no more working
			// hours can be accumulated from this block, but maybe from later blocks.
			// fmt.Printf("OnSameDay s2=%v workstart=%v workend=%v duration=%v\n", s2, ts[i*2], ts[i*2+i], d)
			continue
		}
		if s2.Before(ts[i*2]) {
			// s2 was in non working hours, advance it to next working hour start
			s2 = ts[i*2]
		}
		if end.Before(ts[i*2+1]) {
			// end is prior to the end of this working time block
			// fmt.Printf("OnSameDay s2=%v workstart=%v workend=%v duration=%v\n", s2, ts[i*2], ts[i*2+i], d+end.Sub(s2))
			return d + end.Sub(s2)
		}
		d += ts[i*2+1].Sub(s2)
		// fmt.Printf("OnSameDay s2=%v workstart=%v workend=%v duration=%v\n", s2, ts[i*2], ts[i*2+i], d)
	}
	return d
}

// Return start of day for passed time in the passed time's location.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Return end of day for passed time in the passed time's location.
func endOfDay(t time.Time) time.Time {
	return startOfDay(t.AddDate(0, 0, 1)).Add(-time.Nanosecond)
}

// GetAnonymousWorkTimes is a convenience function for downstream testing.
func GetAnonymousWorkTimes() WorkTimes {
	wt, err := New(map[time.Weekday]bool{
		time.Monday:    true,
		time.Tuesday:   true,
		time.Wednesday: true,
		time.Thursday:  true,
		time.Friday:    true,
	}, []string{
		// Work 9:30am-noon
		"9:30am",
		"12:00pm",
		// 30 minutes for lunch, then work 12:30pm-5:30pm
		"12:30pm",
		"5:30pm",
	})
	if err != nil {
		panic(err)
	}
	return wt
}
