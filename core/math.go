package core

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/rickar/cal"
	"github.com/ryanberckmans/est/core/worktimes"
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

// Return start of day for passed time, in local time.
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Return end of day for passed time, in local time.
func endOfDay(t time.Time) time.Time {
	return startOfDay(t.AddDate(0, 0, 1)).Add(-time.Nanosecond)
}

func businessHoursBetweenTimesOnSameDay(wt worktimes.WorkTimes, start, end time.Time) time.Duration {
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
func businessHoursBetweenTimes(wt worktimes.WorkTimes, start, end time.Time) time.Duration {
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
