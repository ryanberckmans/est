package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rickar/cal"
)

/*
// predict future predicts a cumulative density function
predict_future_cdf :: rand -> tasksToPredict -> historicalVelocities -> monteCarloIterationCount -> [(Completion Date, Cumulative Probability)]

predict_future_cdf_helper :: [predictedTotalHoursForTasksToPredict] -> [(Completion Date, Cumulative Probability)]
*/

// TODO perhaps "accuracy ratio" is better than velocity. Velocity implies a unit relationship and also that faster is better. In this case 1.0 is best.
var fakeHistoricalVelocities = []float64{
	1.0,
	1.3,
	0.7,
	0.5,
	0.4,
	1.6,
	0.8,
}

func monteCarloCDF(iterations int, rand *rand.Rand, historicalAccuracyRatios []float64, estimatesToPredict []float64) {
}

/*
	different ways to represent samples

	[float64]

	[Date] // where each Date has an equal chance of being the delivery date

	sorted [100]Date      -> for indexed % chance, Date is delivered on that date (I think this is sort of an inverse probability mass function https://en.wikipedia.org/wiki/Probability_mass_function)
				   -> same thing on or before that date, sort of inverse CDF

	A box plot needs 5 data points: 0, 25, 50, 75, 100 percentiles Date:
	sorted [5]Date
*/

/*
	NEXT UP

	  g :: unsorted biz days in future -> percentile
*/
// f :: unsorted distribution -> unsorted business days in future
// TODO unit test
func futureBusinessHoursToTime(bhs []float64) []time.Time {
	c := cal.NewCalendar()
	// TODO add Bread office holidays, configurable vacation, etc.
	ts := make([]time.Time, len(bhs))
	now := time.Now()
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
// toPercentile returns ...
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

func main() {
	// rand: The default Source is safe for concurrent use by multiple goroutines, but Sources created by NewSource are not.
	//  --> we should use default rand source
	// rand.Seed(time.Now().UnixNano())

	c, err := getEstConfig()
	if err != nil {
		fmt.Printf("fatal: %s", err)
		return
	}

	fmt.Printf("estConfig: %+v\n", c)

	f, err := getEstFile(strings.Replace(c.Estfile, "$HOME", os.Getenv("HOME"), -1)) // TODO support replacement of any env
	if err != nil {
		fmt.Printf("fatal: %s", err)
		return
	}

	fmt.Printf("estFile: %+v\n", f)

	toSamples := []float64{
		4,
		8,
		12,
		16,
	}

	var naiveSum float64
	for _, v := range toSamples {
		naiveSum += v
	}

	fmt.Printf("naive sum: %v naive end date: %v\n", naiveSum, futureBusinessHoursToTime([]float64{naiveSum}))

	bhs := sampleDistribution(100, rand.New(rand.NewSource(time.Now().UnixNano())), fakeHistoricalVelocities, toSamples)
	sort.Float64s(bhs)
	fmt.Printf("%+v\n", bhs)
	sampleDates := futureBusinessHoursToTime(bhs)
	// fmt.Printf("%+v\n", )
	fmt.Printf("  0%% %v\n", sampleDates[0])
	fmt.Printf(" 25%% %v\n", sampleDates[24])
	fmt.Printf(" 50%% %v\n", sampleDates[49])
	fmt.Printf(" 75%% %v\n", sampleDates[74])
	fmt.Printf("100%% %v\n", sampleDates[99])
}
