package core

import (
	"github.com/gizak/termui"
)

// PredictedDeliveryDateChart renders a full-terminal chart, for predicted delivery
// dates of passed tasks, until user presses 'Q' to quit.
func PredictedDeliveryDateChart(predictedDaysInFuture []float64, ts []Task, pct []string) {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	lc0 := termui.NewLineChart()
	lc0.BorderLabel = "Predicted delivery date for unstarted, estimated tasks"
	// lc0.Mode = "dot"
	lc0.Data = predictedDaysInFuture
	lc0.Height = 24
	lc0.AxesColor = termui.ColorWhite
	lc0.LineColor = termui.ColorGreen | termui.AttrBold

	xAxisLabel := termui.NewPar("percentiles 0 to 100")
	xAxisLabel.Height = 1
	xAxisLabel.Border = false

	yAxisLabel := termui.NewPar("predicted\ndays in\nfuture\n\nQ to quit")
	yAxisLabel.Height = 20
	yAxisLabel.PaddingTop = 10
	yAxisLabel.PaddingLeft = 5
	yAxisLabel.Border = false

	rs := make([]string, len(ts)+2)
	rs[0] = "Tasks in schedule"
	// rs[1] is newline
	for i := range ts {
		rs[i+2] = ts[i].Name
	}
	taskList := termui.NewList()
	taskList.Border = false
	taskList.Items = rs
	taskList.PaddingTop = 2
	taskList.Height = len(rs) + taskList.PaddingTop

	dds := make([]string, len(pct)+2)
	dds[0] = "Predicted dates"
	// dds[1] is newline
	for i := range pct {
		dds[i+2] = pct[i]
	}
	deliveryDateList := termui.NewList()
	deliveryDateList.Border = false
	deliveryDateList.Items = dds
	deliveryDateList.PaddingTop = 1
	deliveryDateList.PaddingLeft = 1
	deliveryDateList.Height = len(dds) + deliveryDateList.PaddingTop

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(2, 0, yAxisLabel),
			termui.NewCol(8, 0, lc0),
			termui.NewCol(2, 0, deliveryDateList),
		),
		termui.NewRow(termui.NewCol(12, 4, xAxisLabel)),
		termui.NewRow(termui.NewCol(12, 0, taskList)),
	)
	termui.Body.Align()
	termui.Render(termui.Body)
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/wnd/resize", func(e termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})
	termui.Loop()
}
