package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/BurntSushi/toml"
)

type estFile struct {
	// TODO unexported dirty bool determines if estFile is written back
	Version int   // future use, to migrate old estfiles
	Tasks   tasks // a type alias for []task
	// Fake ratios, see historicalEstimateAccuracyRatios().
	// Fake ratios are saved to estFile so they are stable.
	FakeHistoricalEstimateAccuracyRatios []float64
}

// historicalEstimateAccuracyRatios returns the accuracy ratios for historical tasks
// in this estFile. Our definition of historical are tasks which are done and not
// deleted. This returned []float64 is the "evidence" in "evidence-based scheduling".
func (ef estFile) historicalEstimateAccuracyRatios() []float64 {
	/*
		TODO there is an argument to weight outcomes by magnitude of task estimate: larger estimates are often more important to a business and harder to get right. If an estimator's history was 90% accurate, but tasks which were estimated accurately are the smallest 90%, then it seems this estimator's history is less accurate than, say, someone who gets large estimates mostly accurate.

		Another arguement is that if all estimates are created equal, then the estimator has an incentive to break things down into smaller tasks which are easier to estimate. This seems a nice mechanism, but I wonder if it discourages folks from including "fuzzy work", such as research, in task definitions. E.g. if I spend N hours learning and planning so that I can then perfectly estimate a 2 hour task, am I really succeeding at improving my estimates and collecting evidence of historical estimates to predict future delivery dates? If N is untracked, large, or variable, then it seems I am not succeeding.

		My current bias is to try giving larger estimates more predictive weight. How could this be done? Ideas:

		1. the probability that a task is included in historical ratios is proportional to its estimate size, then smaller tasks have a larger chance of being excluded.

		2. normalize smaller task estimates closer to 1.0. Then smaller tasks appear to be "mostly on time" to downstream, which seems okay since small tasks _are_ mostly on time - it is not a significant business result to take pus or minus 30 minutes in a 2 hour task.

		3. use both (1) and (2). Then larger tasks would constitute a larger portion of ratio sample, and also larger tasks, being less normalized towards 1.0, would be increasingly responsible for predicted imperfect schedule.
			--> impl note, today FakeHistoricalEstimateAccuracyRatios are padded after real ones are calculated, but we probably don't want to pad with fakes after real ones are dropped due to sampling in (1). We probably want something like `sampledTasks, droppedTasks = sampleHistory(ef.Tasks); if sampledTasks < 20 pad with droppedTasks` and then only pad fakes at very end.

		Another argument is to match historical accuracy ratios of a certain size with future task estimates of a certain size. If an estimator is good or bad at estimating small tasks, let that reflect in small task predictions, and same for large. To impl this, we might use historicalEstimateAccuracyRatios :: [(EstimatedHours, Ratio)], so that downstream is able to weigh ratios with knowledge of the size of their estimates.
	*/
	ts := ef.Tasks.notDeleted().done()
	ars := make([]float64, len(ts))
	for i := range ts {
		ars[i] = ts[i].estimateAccuracyRatio()
	}

	// If real evidence is scarce, pad with fake ratios, which are expected to be fairly random, displaying a conservative lack of confidence in the estimating ability of our estimator.
	for i := 0; len(ars) < 20 && i < len(ef.FakeHistoricalEstimateAccuracyRatios); i++ {
		ars = append(ars, ef.FakeHistoricalEstimateAccuracyRatios[i])
	}
	return ars
}

func getEstFile(estFileName string) (estFile, error) {
	if err := createFileWithDefaultContentsIfNotExists(estFileName, fakeEstfileContents()); err != nil {
		return estFile{}, fmt.Errorf("couldn't find or create %s: %s", estFileName, err)
	}

	d, err := ioutil.ReadFile(estFileName)
	if err != nil {
		return estFile{}, err
	}

	ef := estFile{}
	_, err = toml.Decode(string(d), &ef)
	return ef, err
}

func fakeEstfile() *estFile {
	// Done task
	t0 := newTask()
	t0.Hours = []float64{6.0, 9.2}
	t0.Name = "organize imports in math.go"
	// t0.ShortName = "math.go imports"
	if !t0.isDone() {
		panic("done task wasn't done")
	}
	// Deleted task
	t1 := newTask()
	t1.Name = "this task was deleted"
	t1.IsDeleted = true
	// Started task
	t2 := newTask()
	t2.Hours = []float64{4.0}
	t2.StartedAt = time.Now()
	t2.Name = "optimize monte carlo functions"
	if !t2.isStarted() {
		panic("started task wasn't started")
	}
	// Estimated tasks
	t3 := newTask()
	t3.Hours = []float64{12}
	t3.Name = "impl est-rm"
	if !t3.isEstimated() || t3.isStarted() {
		panic("estimated task wasn't estimated or is started")
	}
	t4 := newTask()
	t4.Hours = []float64{16}
	t4.Name = "design shared predicted schedule for a team sharing estfiles"
	// t4.ShortName = "team schedule"
	if !t4.isEstimated() || t4.isStarted() {
		panic("estimated task wasn't estimated or is started")
	}
	t5 := newTask()
	t5.Hours = []float64{4.75}
	t5.Name = "#5 task"
	if !t5.isEstimated() || t5.isStarted() {
		panic("estimated task wasn't estimated or is started")
	}
	// More done tasks
	t6 := newTask()
	t6.Hours = []float64{3.0, 3.1}
	t6.Name = "#6 task"
	if !t6.isDone() {
		panic("done task wasn't done")
	}
	t7 := newTask()
	t7.Hours = []float64{8.0, 12.0}
	t7.Name = "#7 task"
	if !t7.isDone() {
		panic("done task wasn't done")
	}
	t8 := newTask()
	t8.Hours = []float64{0.5, 2}
	t8.Name = "#8 task"
	if !t8.isDone() {
		panic("done task wasn't done")
	}
	t9 := newTask()
	t9.Hours = []float64{8.0, 6.5}
	t9.Name = "#9 task"
	if !t9.isDone() {
		panic("done task wasn't done")
	}
	return &estFile{
		Version: 1,
		Tasks: []task{
			*t0,
			*t1,
			*t2,
			*t3,
			*t4,
			*t5,
			*t6,
			*t7,
			*t8,
			*t9,
		},
		FakeHistoricalEstimateAccuracyRatios: makeFakeHistoricalEstimateAccuracyRatios(),
	}
}

func makeFakeHistoricalEstimateAccuracyRatios() []float64 {
	c := 20
	fs := make([]float64, c)
	for i := 0; i < c; i++ {
		fs[i] = rand.NormFloat64()*0.2 + 0.8 // the average task for our fake ratios is delivered in 25% more time than estimated; one sigma of tasks are delivered on time or better (about 16% of tasks). Since fake ratios are used to pad predictions when real evidence is scarce, this is a conservative lack of confidence in a new estimator.
	}
	return fs
}

func fakeEstfileContents() string {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(*fakeEstfile()); err != nil {
		panic(err)
	}
	return buf.String()
}
