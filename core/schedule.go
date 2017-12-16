package core

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ryanberckmans/est/core/worktimes"
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
func DeliverySchedule(wt worktimes.WorkTimes, now time.Time, historicalEstimateAccuracyRatios []float64, ts tasks) [100]time.Time {
	var toSamples []float64
	for i := range ts {
		toSamples = append(toSamples, ts[i].Estimated().Hours())
	}

	samples := sampleDistribution(100, rand.New(rand.NewSource(now.UnixNano())), historicalEstimateAccuracyRatios, toSamples)

	pct := toPercentile(samples) // after writing toPercentile(), realized that the statistical significance of the distribution may change if the iterations in sampleDistribution() differ from 100. I.e. if you do 10k iterations, then pct[99] is going to represent a 1 in 10,000 case, which isn't what the user expects. So toPercentile() isn't too useful because the percentile result model only makes sense if 1% actually means 1 in 100. Right?

	var timeArray [100]time.Time

	// wt.TimeAfter() is slow and this concurrency reduces runtime from 10s to 4s on my macbook.
	type pair struct {
		t time.Time
		i int
	}

	timeAfterChan := make(chan pair, len(pct))

	for i := range pct {
		d, err := time.ParseDuration(fmt.Sprintf("%fh", pct[i]))
		if err != nil {
			panic(err)
		}
		go func(t2 time.Time, d2 time.Duration, j int) {
			timeAfterChan <- pair{wt.TimeAfter(t2, d2), j}
		}(now, d, i)
	}
	for _ = range pct {
		p := <-timeAfterChan
		timeArray[p.i] = p.t
	}
	return timeArray
}
