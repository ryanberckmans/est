package core

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/rickar/cal"
)

// TODO probability mass function = pmf :: []Date -> map[Date]float64 s.t. 0 <= pmf(ds)[i] <= 1 && sum_i pmf(ds)[i] == 1   --> or, type DateChance struct, []DateChance
// TODO probability density function

/*
	WorkTimes
		init(times []string, d map[time.Weekday] bool) // list of (start,end) times where start_i < end_i < start_{i+1}; times are "3:04pm", all times local

		init()
			c := NewCal
			set all workdays to false in c
			set workdays in c using passed map

		GetWorkTimesOnDay(time.Time) []time.Time // returns working start/end times on the day of the passed time, nil if passed time isn't a workday. Guaranteed that len([]time.Time) % 2 == 0, and that these times have monotonically increasing hour:minute in local time on day of passed time.
*/

// TODO WorkTimes package or even own repo
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

// Return start of day for passed time, in local time.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Return end of day for passed time, in local time.
func endOfDay(t time.Time) time.Time {
	return startOfDay(t.AddDate(0, 0, 1)).Add(-time.Nanosecond)
}

func businessHoursBetweenTimesOnSameDay(wt WorkTimes, start, end time.Time) time.Duration {
	if end.Before(start) {
		return businessHoursBetweenTimesOnSameDay(wt, end, start)
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

// TODO doc, finish unit tests.
func businessHoursBetweenTimes(wt WorkTimes, start, end time.Time) time.Duration {
	if end.Before(start) {
		return businessHoursBetweenTimes(wt, end, start)
	}
	// Business hours are relative to a specific timezone; we assume local time.
	if start.Location() != time.Local {
		return businessHoursBetweenTimes(wt, start.Local(), end)
	}
	if end.Location() != time.Local {
		return businessHoursBetweenTimes(wt, start, end.Local())
	}
	// fmt.Printf("businessHoursBetweenTimes start=%v end=%v\n", start, end)
	d := time.Duration(0) // accumulated business hours between start and end
	s2 := start
	for {
		if s2.Year() == end.Year() && s2.YearDay() == end.YearDay() {
			// fmt.Printf("on same day s2=%v end=%v\n", s2, end)
			return d + businessHoursBetweenTimesOnSameDay(wt, s2, end)
		}
		// fmt.Printf("not on same day same day s2=%v end=%v\n", s2, end)
		d += businessHoursBetweenTimesOnSameDay(wt, s2, endOfDay(s2))
		s2 = startOfDay(s2.AddDate(0, 0, 1))
	}
}

// TODO unit test
func futureBusinessHoursToTime(bhs []float64) []time.Time {
	c := cal.NewCalendar()
	// Assume we work M-F. TODO configurable business hours, workdays, business holidays, vacation
	ts := make([]time.Time, len(bhs))
	now := time.Now() // TODO inject now
	for i := range bhs {
		ts[i] = c.WorkdaysFrom(now, businessHoursToDays(bhs[i]))
	}
	return ts
}

// TODO unit test
func businessHoursToDays(h float64) int {
	// Assume we work 8 hours per day. TODO configurable business hours, workdays, business holidays, vacation
	businessHoursInAday := 8.0
	d := 0
	for h > businessHoursInAday {
		d++
		h -= businessHoursInAday
	}
	return d
}

// TODO unit test
func toPercentile(in []float64) [100]float64 {
	if len(in) < 100 {
		panic(fmt.Sprintf("toPercentile expected input len >= 100, len was %d", len(in)))
	}
	o := make([]float64, len(in))
	copy(o, in)
	sort.Float64s(o)
	pct := 99
	var o2 [100]float64
	// build o2 from largest to smallest values, so that as we pidgeonhole o into o2 we use the largest of each "eligbile" value for each percentile bucket, with the result that o2[i] means that (i+1)% of data <= that value.
	for i := len(o) - 1; i > -1; i-- {
		if 100*i/len(o) <= pct {
			o2[pct] = o[i]
			pct--
		}
	}
	o2[0] = o[0] // o2[i] means that (i+1)% of data <= that value. So o2[0] means 1% of data smaller than that value, which is correct. But, as a design decision, hardcode o2[0] = o[0], so that the first and last elements of o2 are the smallest and largest elements in o, fulfilling our goal of showing full spectrum of values.
	if pct != -1 {
		panic(fmt.Sprintf("toPercentile pct wasn't -1, it was %d", pct))
	}
	return o2
}

// Return an unsorted distribution of samples
// TODO unit test
func sampleDistribution(iterations int, rd *rand.Rand, historicalRatios []float64, toSamples []float64) []float64 {
	r := make([]float64, iterations)
	for i := 0; i < iterations; i++ {
		r[i] = samples(rd, historicalRatios, toSamples)
	}
	return r
}

// TODO unit test
func samples(rd *rand.Rand, historicalRatios []float64, toSamples []float64) float64 {
	var total float64
	for _, s := range toSamples {
		total += sample(rd, historicalRatios, s)
	}
	return total
}

// TODO unit test
func sample(rd *rand.Rand, historicalRatios []float64, toSample float64) float64 {
	return toSample / historicalRatios[rd.Intn(len(historicalRatios))]
}
