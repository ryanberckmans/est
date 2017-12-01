package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/rickar/cal"
)

// TODO probability mass function = pmf :: []Date -> map[Date]float64 s.t. 0 <= pmf(ds)[i] <= 1 && sum_i pmf(ds)[i] == 1   --> or, type DateChance struct, []DateChance
// TODO probability density function

// TODO unit test
func futureBusinessHoursToTime(bhs []float64) []time.Time {
	c := cal.NewCalendar()
	// TODO add Bread office holidays, configurable vacation, etc.
	ts := make([]time.Time, len(bhs))
	now := time.Now() // TODO inject now
	for i := range bhs {
		ts[i] = c.WorkdaysFrom(now, businessHoursToDays(bhs[i]))
	}
	return ts
}

// TODO unit test
func businessHoursToDays(h float64) int {
	businessHoursInAday := 8.0 // TODO golang seems to want to deafult to float64, maybe we should just use float64 ya? HOw will this affect serialization?
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
func sampleDistribution(iterations int, rand *rand.Rand, historicalRatios []float64, toSamples []float64) []float64 {
	r := make([]float64, iterations)
	for i := 0; i < iterations; i++ {
		r[i] = samples(rand, historicalRatios, toSamples)
	}
	return r
}

// TODO unit test
func samples(rand *rand.Rand, historicalRatios []float64, toSamples []float64) float64 {
	var total float64
	for _, s := range toSamples {
		total += sample(rand, historicalRatios, s)
	}
	return total
}

// TODO unit test
func sample(rand *rand.Rand, historicalRatios []float64, toSample float64) float64 {
	return toSample / historicalRatios[rand.Intn(len(historicalRatios))]
}
