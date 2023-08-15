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
	p.Y.Label.Text = "Values"

	// Make a box plot of our data.
	boxMeanA, err := plotter.NewBoxPlot(vg.Length(20), 0, plotter.Values(dataA))
	if err != nil {
		return fmt.Errorf("Could not create boxplot: %v", err)
	}
	p.Add(boxMeanA)

	boxMeanB, err := plotter.NewBoxPlot(vg.Length(20), 1, plotter.Values(dataB))
	if err != nil {
		return fmt.Errorf("Could not create boxplot: %v", err)
	}
	p.Add(boxMeanB)

	boxMeanC, err := plotter.NewBoxPlot(vg.Length(20), 2, plotter.Values(dataC))
	if err != nil {
		return fmt.Errorf("Could not create boxplot: %v", err)
	}
	p.Add(boxMeanC)

	// p.Add(plotter.NewGrid())

	boxMeanA.FillColor = color.RGBA{139, 4, 221, 1}
	boxMeanB.FillColor = color.RGBA{144, 255, 153, 1}
	boxMeanC.FillColor = color.RGBA{241, 243, 245, 1}

	p.NominalX("AKS", "GKE", "C11n")

	// Save the plot to a PNG file.
	if err := p.Save(6*vg.Inch, 6*vg.Inch, fmt.Sprintf("%s/%s.png", PLOT_PREFIX, filename)); err != nil {
		return fmt.Errorf("Failed to save plot to file: %v", err)
	}

	return nil
}
