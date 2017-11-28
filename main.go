package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/rickar/cal"
	"github.com/spf13/viper"
)

type estConfig struct {
	Estfile string // est file name
}

type estFile struct {
	Version                    int
	Tasks                      []task
	FakeEstimateAccuracyRatios []float64
}

func fakeEstfile() *estFile {
	t0 := newTask()
	t0.Timeline = append(t0.Timeline, time.Now().Add(-time.Hour*24))
	t0.Timeline = append(t0.Timeline, time.Now())
	t0.EstimatedHours = 6
	t0.EstimatedAt = time.Now().Add(-time.Hour * 48)
	t1 := newTask()
	t1.IsDeleted = true
	t2 := newTask()
	t2.Timeline = append(t2.Timeline, time.Now().Add(-time.Hour*36))
	t2.EstimatedHours = 4
	t2.EstimatedAt = time.Now().Add(-time.Hour * 56)
	return &estFile{
		Version: 1,
		Tasks: []task{
			*newTask(),
			*newTask(),
			*t0,
			*t1,
			*t2,
		},
		FakeEstimateAccuracyRatios: fakeHistoricalVelocities,
	}
}

func fakeEstfileContents() string {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(*fakeEstfile()); err != nil {
		panic(err)
	}
	return buf.String()
}

type task struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	EstimatedAt    time.Time
	EstimatedHours float64
	Timeline       []time.Time // one Time per start, stop, start, stop, ... see isDone()
	IsDeleted      bool
}

func newTask() *task {
	return &task{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}
}

func (t *task) isDone() bool {
	return len(t.Timeline) > 0 && len(t.Timeline)%2 == 0
}

func createFileWithDefaultContentsIfNotExists(filename string, defaultContents string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// no-op, filename will be created
	} else if err != nil {
		return err
	} else {
		// filename exists, never overwrite
		return nil
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(defaultContents))
	return err
}

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

const estConfigDefaultContents string = `
# Your estfile stores your tasks and estimates. Some users may want to change this to a location with automatic backup, such as Dropbox or Google Drive.
estfile = "$HOME/.estfile.toml"
`

const estConfigDefaultFileNameNoSuffix string = ".estconfig"
const estConfigDefaultFileSuffix string = ".toml"
const estConfigDefaultFileName string = estConfigDefaultFileNameNoSuffix + estConfigDefaultFileSuffix

const estfileDefaultContents string = `
[[task]]
`

// getEstconfig returns the singleton estConfig for this process.
// Creates a config file if none found.
func getEstConfig() (estConfig, error) {
	if err := createFileWithDefaultContentsIfNotExists(os.Getenv("HOME")+"/"+estConfigDefaultFileName, estConfigDefaultContents); err != nil {
		return estConfig{}, fmt.Errorf("couldn't find or create %s: %s", estConfigDefaultFileName, err)
	}

	viper.SetConfigName(estConfigDefaultFileNameNoSuffix) // .toml suffix discovered automatically
	viper.AddConfigPath("$HOME")
	if err := viper.ReadInConfig(); err != nil {
		return estConfig{}, err
	}

	c := estConfig{}
	err := viper.Unmarshal(&c)
	return c, err
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

	// println("begin")
	// buf := new(bytes.Buffer)
	// fmt.Println(buf.String())
	// if err := toml.NewEncoder(os.Stdout).Encode(*fakeEstfile()); err != nil {
	// panic(err)
	// }
	// println("end")

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
