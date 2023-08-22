package main

import (
	"fmt"
	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func plotData(dataA, dataB, dataC []float64, title, filename string) error {
	p := plot.New()

	p.Title.Text = title
	p.Y.Label.Text = "Latency (s)"
	p.X.Label.Text = "Cloud Provider"

	// Make a box plot of our data.
	boxA, err := plotter.NewBoxPlot(vg.Length(20), 0, plotter.Values(dataA))
	if err != nil {
		return fmt.Errorf("Could not create boxplot: %v", err)
	}
	p.Add(boxA)

	boxB, err := plotter.NewBoxPlot(vg.Length(20), 1, plotter.Values(dataB))
	if err != nil {
		return fmt.Errorf("Could not create boxplot: %v", err)
	}
	p.Add(boxB)

	boxC, err := plotter.NewBoxPlot(vg.Length(20), 2, plotter.Values(dataC))
	if err != nil {
		return fmt.Errorf("Could not create boxplot: %v", err)
	}
	p.Add(boxC)

	p.Add(plotter.NewGrid())

	boxA.FillColor = color.RGBA{139, 4, 221, 1}
	boxB.FillColor = color.RGBA{144, 255, 153, 1}
	boxC.FillColor = color.RGBA{241, 243, 245, 1}

	p.NominalX("AKS", "GKE", "C11n")

	p.Y.Tick.Marker = myTicker{}

	// Save the plot to a PNG file.
	if err := p.Save(6*vg.Inch, 6*vg.Inch, fmt.Sprintf("%s/%s.png", PLOT_PREFIX, filename)); err != nil {
		return fmt.Errorf("Failed to save plot to file: %v", err)
	}

	return nil
}

type myTicker struct{}

func (myTicker) Ticks(min, max float64) []plot.Tick {
	values := equallySpacedValues(min, max, 10)
	ticks := []plot.Tick{}
	for _, v := range values {
		ticks = append(ticks, plot.Tick{Value: v, Label: fmt.Sprintf("%.2f", v)})
	}

	return ticks
}

func equallySpacedValues(min, max float64, numValues int) []float64 {
	values := make([]float64, numValues)

	for i := 0; i < numValues; i++ {
		values[i] = min + (float64(i)/float64(numValues-1))*(max-min)
	}

	return values
}
