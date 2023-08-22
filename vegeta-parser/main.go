package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/montanaflynn/stats"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	AKS_DATA    = "../data/1300w/5replicas/aks"
	GKE_DATA    = "../data/1300w/5replicas/gke"
	C11N_DATA   = "../data/1300w/5replicas/c11n"
	PLOT_PREFIX = "../boxplots"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	resultsAKS, err := parseResults(AKS_DATA)
	if err != nil {
		return fmt.Errorf("parsing results: %w", err)
	}

	resultsGKE, err := parseResults(GKE_DATA)
	if err != nil {
		return fmt.Errorf("parsing results: %w", err)
	}

	resultsC11n, err := parseResults(C11N_DATA)
	if err != nil {
		return fmt.Errorf("parsing results: %w", err)
	}

	aksRun, err := getBasicStats(resultsAKS)
	if err != nil {
		return fmt.Errorf("getting aks stats: %w", err)
	}
	fmt.Println("========== Results AKS ==========")
	fmt.Println(aksRun)

	gkeRun, err := getBasicStats(resultsGKE)
	if err != nil {
		return fmt.Errorf("getting c11n stats: %w", err)
	}
	fmt.Println("========== Results GKE ==========")
	fmt.Println(gkeRun)

	c11nRun, err := getBasicStats(resultsC11n)
	if err != nil {
		return fmt.Errorf("getting c11n stats: %w", err)
	}
	fmt.Println("========== Results C11n ==========")
	fmt.Println(c11nRun)

	fmt.Println("========== AKS vs C11n ==========")
	fmt.Print(getDifference(aksRun, c11nRun))
	fmt.Println("========== GKE vs C11n ==========")
	fmt.Print(getDifference(gkeRun, c11nRun))

	if err := plotData(aksRun.MeanRaw, gkeRun.MeanRaw, c11nRun.MeanRaw, "Mean", "mean_latency"); err != nil {
		return fmt.Errorf("plotting mean: %w", err)
	}
	if err := plotData(aksRun.P99Raw, gkeRun.P99Raw, c11nRun.P99Raw, "99th Percentile", "p99_latency"); err != nil {
		return fmt.Errorf("plotting p99: %w", err)
	}
	if err := plotData(aksRun.MaxRaw, gkeRun.MaxRaw, c11nRun.MaxRaw, "Maximum", "max_latency"); err != nil {
		return fmt.Errorf("plotting max: %w", err)
	}
	if err := plotData(aksRun.MinRaw, gkeRun.MinRaw, c11nRun.MinRaw, "Minimum", "min_latency"); err != nil {
		return fmt.Errorf("plotting min: %w", err)
	}

	return nil
}

func getDifference(a, b runStats) string {
	meanDiff := (1 - (a.Mean.Mean / b.Mean.Mean)) * 100
	p99Diff := (1 - (a.P99.Mean / b.P99.Mean)) * 100
	maxDiff := (1 - (a.Max.Mean / b.Max.Mean)) * 100
	minDiff := (1 - (a.Min.Mean / b.Min.Mean)) * 100

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Mean:\t%s %%\n", writeSignedPercentage(meanDiff)))
	builder.WriteString(fmt.Sprintf("P99:\t%s %%\n", writeSignedPercentage(p99Diff)))
	builder.WriteString(fmt.Sprintf("Max:\t%s %%\n", writeSignedPercentage(maxDiff)))
	builder.WriteString(fmt.Sprintf("Min:\t%s %%\n", writeSignedPercentage(minDiff)))

	return builder.String()
}

func writeSignedPercentage(percentage float64) string {
	sign := ""
	if percentage > 0 {
		sign = "+"
	}

	return fmt.Sprintf("%s%2f", sign, percentage)
}

func getBasicStats(results []result) (runStats, error) {
	mean, err := getMean(results)
	if err != nil {
		return runStats{}, fmt.Errorf("getting mean: %w", err)
	}
	p99, err := getP99(results)
	if err != nil {
		return runStats{}, fmt.Errorf("getting p99: %w", err)
	}
	max, err := getMax(results)
	if err != nil {
		return runStats{}, fmt.Errorf("getting max: %w", err)
	}
	min, err := getMin(results)
	if err != nil {
		return runStats{}, fmt.Errorf("getting min: %w", err)
	}

	meanStats, err := NewStatContainer(mean)
	if err != nil {
		return runStats{}, fmt.Errorf("calculating mean stats: %w", err)
	}
	p99Stats, err := NewStatContainer(p99)
	if err != nil {
		return runStats{}, fmt.Errorf("calculating p99 stats: %w", err)
	}
	maxStats, err := NewStatContainer(max)
	if err != nil {
		return runStats{}, fmt.Errorf("calculating max stats: %w", err)
	}
	minStats, err := NewStatContainer(min)
	if err != nil {
		return runStats{}, fmt.Errorf("calculating min stats: %w", err)
	}

	run := runStats{
		Mean:    meanStats,
		P99:     p99Stats,
		Max:     maxStats,
		Min:     minStats,
		MeanRaw: mean,
		P99Raw:  p99,
		MaxRaw:  max,
		MinRaw:  min,
	}

	return run, nil
}

// name: result
type runResult struct {
	Total   vegeta.Metrics `json:"total"`
	Decrypt vegeta.Metrics `json:"transit_decrypt_test_1"`
	Encrypt vegeta.Metrics `json:"transit_encrypt_test_1"`
	Sign    vegeta.Metrics `json:"transit_sign_test_1"`
	Verify  vegeta.Metrics `json:"transit_verify_test_1"`
}

type result struct {
	Metrics   runResult `json:"metrics"`
	TargetAdr string    `json:"target_addr"`
}

// runStats holds statContainers for multiple data dimensions of a single benchmark run.
type runStats struct {
	Mean    statContainer
	P99     statContainer
	Max     statContainer
	Min     statContainer
	MeanRaw []float64
	P99Raw  []float64
	MaxRaw  []float64
	MinRaw  []float64
}

func (r runStats) String() string {
	return fmt.Sprintf("Mean:\t%s\nP99:\t%s\nMax:\t%s\nMin:\t%s", r.Mean, r.P99, r.Max, r.Min)
}

// statContainer holds different measures for a single data dimension.
type statContainer struct {
	Mean     float64
	Variance float64
}

func NewStatContainer(data []float64) (statContainer, error) {
	mean, err := stats.Mean(data)
	if err != nil {
		return statContainer{}, fmt.Errorf("calculating mean: %w", err)
	}
	variance, err := stats.Variance(data)
	if err != nil {
		return statContainer{}, fmt.Errorf("calculating variance: %w", err)
	}

	return statContainer{mean, variance}, nil
}

func (s statContainer) String() string {
	return fmt.Sprintf("mean: %f, variance: %f", s.Mean, s.Variance)
}

func getMean(results []result) ([]float64, error) {
	var mean []float64

	for _, result := range results {
		mean = append(mean, result.Metrics.Total.Latencies.Mean.Seconds())
	}

	return mean, nil
}

func getP99(results []result) ([]float64, error) {
	var p99 []float64

	for _, result := range results {
		p99 = append(p99, result.Metrics.Total.Latencies.P99.Seconds())
	}

	return p99, nil
}

func getMax(results []result) ([]float64, error) {
	var max []float64

	for _, result := range results {
		max = append(max, result.Metrics.Total.Latencies.Max.Seconds())
	}

	return max, nil
}

func getMin(results []result) ([]float64, error) {
	var min []float64

	for _, result := range results {
		min = append(min, result.Metrics.Total.Latencies.Min.Seconds())
	}

	return min, nil
}

func parseResults(dataDir string) ([]result, error) {
	// Get a list of all files in the data directory
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return []result{}, fmt.Errorf("reading directory: %w", err)
	}

	// Create a results map to store the metrics for each file
	results := []result{}

	// Loop through each file and read its contents
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Read the contents of the file into a byte slice
		filename := filepath.Join(dataDir, file.Name())
		data, err := os.ReadFile(filename)
		if err != nil {
			return []result{}, fmt.Errorf("reading file: %w", err)
		}

		// Parse the metrics from the file contents
		metrics := result{}
		err = json.Unmarshal(data, &metrics)
		if err != nil {
			return []result{}, fmt.Errorf("parsing metrics from %s: %w", filename, err)
		}

		// Add the metrics to the results map
		results = append(results, metrics)
	}

	return results, nil
}
