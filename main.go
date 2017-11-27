package main

import (
	"math/rand"
	"time"

	"github.com/rickar/cal"
)

/*
// predict future predicts a cumulative density function
predict_future_cdf :: rand -> tasksToPredict -> historicalVelocities -> monteCarloIterationCount -> [(Completion Date, Cumulative Probability)]

predict_future_cdf_helper :: [predictedTotalHoursForTasksToPredict] -> [(Completion Date, Cumulative Probability)]
*/

// TODO perhaps "accuracy ratio" is better than velocity. Velocity implies a unit relationship and also that faster is better. In this case 1.0 is best.
var fakeHistoricalVelocities = []float32{
	1.0,
	1.3,
	0.7,
	0.5,
	0.4,
	1.6,
	0.8,
}

func monteCarloCDF(iterations int, rand *rand.Rand, historicalAccuracyRatios []float32, estimatesToPredict []float32) {
}

/*
	different ways to represent samples

	[float32]

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
func futureBusinessHoursToTime(bhs []float32) []time.Time {
	c := cal.NewCalendar()
	// TODO add Bread office holidays, configurable vacation, etc.
	// NEXT UP use businessHoursToDays() to convert this to business days, then use cal WorkdaysFrom(time.Now) to get the slice of Time
	return nil
}

// TODO unit test
func businessHoursToDays(h float32) int {
	var businessHoursInAday float32 = 8.0 // TODO golang seems to want to deafult to float64, maybe we should just use float64 ya? HOw will this affect serialization?
	d := 0
	for h > businessHoursInAday {
		d++
		h -= businessHoursInAday
	}
	return d
}

// Return an unsorted distribution of samples
func sampleDistribution(iterations int, rand *rand.Rand, historicalRatios []float32, toSamples []float32) []float32 {
	r := make([]float32, iterations)
	for i := 0; i < iterations; i++ {
		r[i] = samples(rand, historicalRatios, toSamples)
	}
	return r
}

func samples(rand *rand.Rand, historicalRatios []float32, toSamples []float32) float32 {
	var total float32
	for _, s := range toSamples {
		total += sample(rand, historicalRatios, s)
	}
	return total
}

func sample(rand *rand.Rand, historicalRatios []float32, toSample float32) float32 {
	return toSample / historicalRatios[rand.Intn(len(historicalRatios))]
}

func foo() {
}

func main() {
	// rand: The default Source is safe for concurrent use by multiple goroutines, but Sources created by NewSource are not.
	//  --> we should use default rand source
	rand.Seed(time.Now().UnixNano())
}
