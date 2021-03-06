package core

import (
	"sort"
	"time"
)

// AccuracyRatio is an anonymous data point representing the accuracy of some
// (expected, actual) pair. The time is e.g. task done task
type AccuracyRatio struct {
	time     time.Time     // anonymous time of this data point, e.g. task done date
	duration time.Duration // anonymous duration of this data point, e.g. task estimate
	ratio    float64       // data point, ratio of (expected, actual)
}

// AccuracyRatios provides convenience functions.
type AccuracyRatios []AccuracyRatio

// SortByTimeAscending sorts this []AccuracyRatio in place by ascending ar.time.
func (ars AccuracyRatios) SortByTimeAscending() AccuracyRatios {
	sort.Sort(sortByTimeAscending(ars))
	return ars
}

// Ratios returns all []AccuracyRatio.ratio.
func (ars AccuracyRatios) Ratios() []float64 {
	rs := make([]float64, len(ars))
	for i := range ars {
		rs[i] = ars[i].ratio
	}
	return rs
}

// After returns subset of accuracy ratios which are after the passed time.
func (ars AccuracyRatios) After(t time.Time) AccuracyRatios {
	return filterAccuracyRatios(ars, func(ar *AccuracyRatio) bool {
		return ar.time.After(t)
	})
}

func filterAccuracyRatios(ars AccuracyRatios, fn func(t *AccuracyRatio) bool) AccuracyRatios {
	if ars == nil {
		return nil
	}
	var res AccuracyRatios
	for i := range ars {
		if fn(&ars[i]) {
			res = append(res, ars[i])
		}
	}
	return res
}

type sortByTimeAscending AccuracyRatios

func (ars sortByTimeAscending) Len() int {
	return len(ars)
}
func (ars sortByTimeAscending) Less(i, j int) bool {
	return ars[i].time.Before(ars[j].time)
}
func (ars sortByTimeAscending) Swap(i, j int) {
	tmp := ars[j]
	ars[j] = ars[i]
	ars[i] = tmp
}
