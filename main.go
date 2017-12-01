package main

import (
	"os"

	"github.com/ryanberckmans/est/core"
)

func main() {
	core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
		ts := ef.Tasks.NotDeleted().Estimated().NotStarted()
		dates := core.DeliverySchedule(ef.HistoricalEstimateAccuracyRatios(), ts)
		os.Stdout.WriteString(core.RenderDeliverySchedule(dates))
	}, func() {
		// failed to load estconfig or estfile. Err printed elsewhere.
	})
}
