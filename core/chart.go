package core

import (
	"errors"
	"io/ioutil"
	"math"
	"time"

	"github.com/gizak/termui"
	"github.com/pkg/browser"
	chart "github.com/wcharczuk/go-chart"
	drawing "github.com/wcharczuk/go-chart/drawing"
)

// PredictedDeliveryDateChart renders a full-terminal chart, for predicted delivery
// dates of passed tasks, until user presses 'Q' to quit.
func PredictedDeliveryDateChart(predictedDaysInFuture []float64, ts tasks, pct []string) {
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
		rs[i+2] = ts[i].Name()
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

// AccuracyRatioChart shows how accuracy the user has estimated,
// providing a feedback loop to become a better estimator.
func AccuracyRatioChart(ars []AccuracyRatio, now time.Time) error {
	if len(ars) < 1 {
		return errors.New("no historical accuracy ratios")
	}
	return openChartInBrowser(makeAccuracyRatioChart(ars, now), "est-how-am-i-doing-")
}

func makeAccuracyRatioChart(ars []AccuracyRatio, now time.Time) *chart.Chart {
	// X axis: days since passed now
	// Y axis: accuracy ratio
	// Dot width: log of estimated hours (larger dot is larger estimate)
	xs := make([]float64, len(ars))
	ys := make([]float64, len(ars))
	dws := make([]float64, len(ars))
	minDaysAgo := math.MaxFloat64
	maxDaysAgo := 0.0
	for i := range ars {
		xs[i] = math.Floor(now.Sub(ars[i].time).Hours() / 24.0)
		if xs[i] < 0 {
			// disallow future dates; this is a retrospective chart and future days are unexpected.
			xs[i] = 0
		}
		if xs[i] < minDaysAgo {
			minDaysAgo = xs[i]
		}
		if xs[i] > maxDaysAgo {
			maxDaysAgo = xs[i]
		}
		ys[i] = ars[i].ratio
		h := ars[i].duration.Hours()
		baseDiameter := 2.0
		var w float64
		switch {
		case h <= 1:
			w = baseDiameter
		default:
			// Area of scatter dot increases linearly with estimate size
			w = baseDiameter + math.Sqrt(2*h)
		}
		dws[i] = w
	}
	dcFunc := func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
		ratio := y
		if ratio < 1 {
			ratio = 1.0 / ratio
		}
		switch {
		case ratio < 1.2:
			return drawing.ColorGreen
		case ratio < 1.4:
			return drawing.Color{R: 255, G: 255, B: 0, A: 255} // yellow
		case ratio < 2.0:
			return drawing.Color{R: 255, G: 140, B: 0, A: 255} // orange
		}
		return drawing.ColorRed
	}
	dwFunc := func(xrange, yrange chart.Range, index int, x, y float64) float64 {
		return dws[index]
	}
	// fmt.Printf("makeAccuracyRatiochart\nars=%+v\nnow=%v\nxs=%+v\nys=%+v\ndws=%+v\nmaxDaysAgo=%f", ars, now, xs, ys, dws, maxDaysAgo)
	c := &chart.Chart{
		XAxis: chart.XAxis{
			Name:      "Calendar Days Ago",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Name:      "Accuracy Ratio (estimate / actual)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    30,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				// Name: "See `est h -h` for explanation of chart",
				Style: chart.Style{
					Show:             true,
					StrokeWidth:      chart.Disabled,
					DotWidthProvider: dwFunc,
					// DotColor:         drawing.ColorBlue,
					DotColorProvider: dcFunc,
				},
				XValues: xs,
				YValues: ys,
			},
			chart.ContinuousSeries{
				Name: "Larger dots are larger estimates. The green line represents perfect estimates (i.e. accuracy ratio == 1.0)",
				Style: chart.Style{
					Show:        true,
					StrokeColor: drawing.ColorGreen,
				},
				XValues: []float64{minDaysAgo, maxDaysAgo},
				YValues: []float64{1.0, 1.0},
			},
		},
	}
	c.Elements = []chart.Renderable{
		chart.LegendThin(c),
	}
	return c
}

func openChartInBrowser(c *chart.Chart, chartName string) error {
	f, err := ioutil.TempFile("", chartName)
	if err != nil {
		return err
	}
	c.Render(chart.PNG, f)
	err = f.Close()
	if err != nil {
		return err
	}
	return browser.OpenFile(f.Name())
}
