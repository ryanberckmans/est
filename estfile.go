package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
)

type estFile struct {
	Version                    int
	Tasks                      tasks // a type alias for []task
	FakeEstimateAccuracyRatios []float64
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

	// future tasks with estimates
	t3 := newTask()
	t3.EstimatedHours = 12
	t3.EstimatedAt = time.Now().Add(-time.Hour * 56)
	t4 := newTask()
	t4.EstimatedHours = 16
	t4.EstimatedAt = time.Now().Add(-time.Hour * 20)
	t5 := newTask()
	t5.EstimatedHours = 4.75
	t5.EstimatedAt = time.Now().Add(-time.Hour * 20)
	return &estFile{
		Version: 1,
		Tasks: []task{
			*newTask(),
			*newTask(),
			*t0,
			*t1,
			*t2,
			*t3,
			*t4,
			*t5,
		},
		FakeEstimateAccuracyRatios: []float64{
			1.0,
			1.3,
			0.7,
			0.5,
			0.4,
			1.6,
			0.8,
		},
	}
}

func fakeEstfileContents() string {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(*fakeEstfile()); err != nil {
		panic(err)
	}
	return buf.String()
}
