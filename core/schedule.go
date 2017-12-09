package core

import (
	"fmt"
	"math/rand"
	"time"
)

// RenderDeliverySchedule returns a list of string
// delivery dates for the passed delivery dates percentile.
func RenderDeliverySchedule(dates [100]time.Time) [21]string {
	var ss [21]string
	ss[0] = fmt.Sprintf("%3d%% %s", 0, dates[0].Format("Jan 2"))
	for i := 1; i < 21; i++ {
		j := i*5 - 1
		ss[i] = fmt.Sprintf("%3d%% %s", j+1, dates[j].Format("Jan 2"))
	}
	return ss
}

// DeliverySchedule returns a predicted delivery schedule, as a time percentile,
// using the passed historical data as a basis for future work on the passed tasks.
func DeliverySchedule(historicalEstimateAccuracyRatios []float64, ts tasks) [100]time.Time {
	var toSamples []float64
	for i := range ts {
		toSamples = append(toSamples, ts[i].Estimated().Hours())
	}

	samples := sampleDistribution(100, rand.New(rand.NewSource(time.Now().UnixNano())), historicalEstimateAccuracyRatios, toSamples)

	pct := toPercentile(samples) // after writing toPercentile(), realized that the statistical significance of the distribution may change if the iterations in sampleDistribution() differ from 100. I.e. if you do 10k iterations, then pct[99] is going to represent a 1 in 10,000 case, which isn't what the user expects. So toPercentile() isn't too useful because the percentile result model only makes sense if 1% actually means 1 in 100. Right?

	timeSlice := futureBusinessHoursToTime(pct[:])
	var timeArray [100]time.Time
	numCopied := copy(timeArray[:], timeSlice)
	if numCopied != 100 {
		panic(fmt.Sprintf("expected to copy 100 elements when building time percentile, copied %d", numCopied))
	}
	return timeArray
}
