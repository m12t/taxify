/*
A simple CLI program for estimating one's state income tax in all 50 states at once

Resources:
https://taxfoundation.org/state-income-tax-rates-2022/
*/

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

type State struct {
	name          string
	abbrev        string
	brackets      []int
	rates         []float64
	effectiveRate float64
	incomeTax     int
}

func initializeStates(income *float64, numSteps *int) *[51]*State {
	states := [51]*State{
		{
			name:     "Alabama",
			abbrev:   "AL",
			brackets: []int{0, 500, 3000},
			rates:    []float64{0.02, 0.03, 0.05},
		},
		{
			name:     "Alaska",
			abbrev:   "AK",
			brackets: []int{0},
			rates:    []float64{0.0}, // like a flat tax of 0.0%
		},
		{
			name:     "Arizona",
			abbrev:   "AZ",
			brackets: []int{0, 27808, 55615, 116843},
			rates:    []float64{0.0259, 0.0334, 0.0417, 0.045},
		},
		{
			name:     "Arkansas",
			abbrev:   "AR",
			brackets: []int{0, 4300, 8500},
			rates:    []float64{0.02, 0.04, 0.055},
		},
		{
			name:     "California",
			abbrev:   "CA",
			brackets: []int{0, 9325, 22107, 34892, 48435, 61214, 312686, 375221, 625369, 1000000},
			rates:    []float64{0.01, 0.02, 0.04, 0.06, 0.08, 0.093, 0.103, 0.113, 0.123, 0.133},
		},
		{
			name:     "Colorado",
			abbrev:   "CO",
			brackets: []int{0},
			rates:    []float64{0.0455},
		},
		{
			name:     "Connecticut",
			abbrev:   "CT",
			brackets: []int{0, 10000, 50000, 100000, 200000, 250000, 500000},
			rates:    []float64{0.03, 0.05, 0.055, 0.06, 0.065, 0.069, 0.0699},
		},
		{
			name:     "Delaware",
			abbrev:   "DE",
			brackets: []int{2000, 5000, 10000, 20000, 25000, 60000},
			rates:    []float64{0.022, 0.039, 0.048, 0.052, 0.0555, 0.066},
		},
		{
			name:     "Florida",
			abbrev:   "FL",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "Georgia",
			abbrev:   "GA",
			brackets: []int{0, 750, 2250, 3750, 5250, 7000},
			rates:    []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.0575},
		},
		{
			name:     "Hawaii",
			abbrev:   "HI",
			brackets: []int{0, 2400, 4800, 9600, 14400, 19200, 24000, 36000, 48000, 150000, 175000, 200000},
			rates:    []float64{0.014, 0.032, 0.055, 0.064, 0.068, 0.072, 0.076, 0.079, 0.0825, 0.09, 0.1, 0.11},
		},
		{
			name:     "Idaho",
			abbrev:   "ID",
			brackets: []int{0, 1588, 4763, 7939},
			rates:    []float64{0.01, 0.03, 0.045, 0.06},
		},
		{
			name:     "Illinois",
			abbrev:   "IL",
			brackets: []int{0},
			rates:    []float64{0.0495},
		},
		{
			name:     "Indiana",
			abbrev:   "IN",
			brackets: []int{0},
			rates:    []float64{0.0323},
		},
		{
			name:     "Iowa",
			abbrev:   "IA",
			brackets: []int{0, 1743, 3486, 6972, 15687, 26145, 34860, 52290, 78435},
			rates:    []float64{0.0033, 0.0067, 0.0225, 0.0414, 0.0563, 0.0596, 0.0625, 0.0744, 0.0853},
		},
		{
			name:     "Kansas",
			abbrev:   "KS",
			brackets: []int{0, 15000, 30000},
			rates:    []float64{0.031, 0.0525, 0.057},
		},
		{
			name:     "Kentucky",
			abbrev:   "KY",
			brackets: []int{0},
			rates:    []float64{0.050},
		},
		{
			name:     "Louisiana",
			abbrev:   "LA",
			brackets: []int{0, 12500, 50000},
			rates:    []float64{0.0185, 0.035, 0.0425},
		},
		{
			name:     "Maine",
			abbrev:   "ME",
			brackets: []int{0, 23000, 54450},
			rates:    []float64{0.058, 0.0675, 0.0715},
		},
		{
			name:     "Maryland",
			abbrev:   "MD",
			brackets: []int{0, 1000, 2000, 3000, 100000, 125000, 150000, 250000},
			rates:    []float64{0.02, 0.03, 0.04, 0.0475, 0.05, 0.0525, 0.055, 0.0575},
		},
		{
			name:     "Massachusetts",
			abbrev:   "MA",
			brackets: []int{0},
			rates:    []float64{0.05},
		},
		{
			name:     "Michigan",
			abbrev:   "MI",
			brackets: []int{0},
			rates:    []float64{0.0425},
		},
		{
			name:     "Minnesota",
			abbrev:   "MN",
			brackets: []int{0, 28080, 92230, 171220},
			rates:    []float64{0.0535, 0.068, 0.0785, 0.0985},
		},
		{
			name:     "Mississippi",
			abbrev:   "MS",
			brackets: []int{5000, 10000},
			rates:    []float64{0.04, 0.05},
		},
		{
			name:     "Missouri",
			abbrev:   "MO",
			brackets: []int{108, 1088, 2176, 3264, 4352, 5440, 6528, 7616, 8704},
			rates:    []float64{0.015, 0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.054},
		},
		{
			name:     "Montana",
			abbrev:   "MT",
			brackets: []int{0, 3100, 5500, 8400, 11400, 14600, 18800},
			rates:    []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.0675},
		},
		{
			name:     "Nebraska",
			abbrev:   "NE",
			brackets: []int{0, 3440, 20590, 33180},
			rates:    []float64{0.0246, 0.0351, 0.0501, 0.0684},
		},
		{
			name:     "Nevada",
			abbrev:   "NV",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "New Hampshire",
			abbrev:   "NH",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "New Jersey",
			abbrev:   "NJ",
			brackets: []int{0, 20000, 35000, 40000, 75000, 500000, 1000000},
			rates:    []float64{0.014, 0.0175, 0.035, 0.05525, 0.0637, 0.0897, 0.1075},
		},
		{
			name:     "New Mexico",
			abbrev:   "NM",
			brackets: []int{0, 5500, 11000, 16000, 210000},
			rates:    []float64{0.017, 0.032, 0.047, 0.049, 0.059},
		},
		{
			name:     "New York",
			abbrev:   "NY",
			brackets: []int{0, 8500, 11700, 13900, 80650, 215400, 1077550, 5000000, 25000000},
			rates:    []float64{0.04, 0.045, 0.0525, 0.0585, 0.0625, 0.0685, 0.0965, 0.103, 0.109},
		},
		{
			name:     "North Carolina",
			abbrev:   "NYC",
			brackets: []int{0},
			rates:    []float64{0.0499},
		},
		{
			name:     "North Dakota",
			abbrev:   "ND",
			brackets: []int{0, 40525, 98100, 204675, 445000},
			rates:    []float64{0.011, 0.0204, 0.0227, 0.0264, 0.029},
		},
		{
			name:     "Ohio",
			abbrev:   "OH",
			brackets: []int{25000, 44250, 88450, 110650},
			rates:    []float64{0.02765, 0.03226, 0.03688, 0.0399},
		},
		{
			name:     "Oklahoma",
			abbrev:   "OK",
			brackets: []int{0, 1000, 2500, 3750, 4900, 7200},
			rates:    []float64{0.0025, 0.0075, 0.0175, 0.0275, 0.0375, 0.0475},
		},
		{
			name:     "Oregon",
			abbrev:   "OR",
			brackets: []int{0, 3650, 9200, 125000},
			rates:    []float64{0.0475, 0.0675, 0.0875, 0.099},
		},
		{
			name:     "Pennsylvania",
			abbrev:   "PA",
			brackets: []int{0},
			rates:    []float64{0.0307},
		},
		{
			name:     "Rhode Island",
			abbrev:   "RI",
			brackets: []int{0, 68200, 155050},
			rates:    []float64{0.0375, 0.0475, 0.0599},
		},
		{
			name:     "South Carolina",
			abbrev:   "SC",
			brackets: []int{0, 3200, 6410, 9620, 12820, 16040},
			rates:    []float64{0.0, 0.03, 0.04, 0.05, 0.06, 0.07},
		},
		{
			name:     "South Dakota",
			abbrev:   "SD",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "Tennessee",
			abbrev:   "TN",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "Texas",
			abbrev:   "TX",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "Utah",
			abbrev:   "UT",
			brackets: []int{0},
			rates:    []float64{0.0495},
		},
		{
			name:     "Vermont",
			abbrev:   "VT",
			brackets: []int{0, 40950, 99200, 206950},
			rates:    []float64{0.0335, 0.066, 0.076, 0.0875},
		},
		{
			name:     "Virginia",
			abbrev:   "VA",
			brackets: []int{0, 3000, 5000, 17000},
			rates:    []float64{0.02, 0.03, 0.05, 0.0575},
		},
		{
			name:     "Washington",
			abbrev:   "WA",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "West Virginia",
			abbrev:   "WV",
			brackets: []int{0, 10000, 25000, 40000, 60000},
			rates:    []float64{0.03, 0.04, 0.045, 0.06, 0.065},
		},
		{
			name:     "Wisconsin",
			abbrev:   "WI",
			brackets: []int{0, 12760, 25520, 280950},
			rates:    []float64{0.0354, 0.0465, 0.053, 0.0765},
		},
		{
			name:     "Wyoming",
			abbrev:   "WY",
			brackets: []int{0},
			rates:    []float64{0.0},
		},
		{
			name:     "Washington D.C.",
			abbrev:   "DC",
			brackets: []int{0, 10000, 40000, 60000, 250000, 500000, 1000000},
			rates:    []float64{0.04, 0.06, 0.065, 0.085, 0.0925, 0.0975, 0.1075},
		},
	}
	for _, state := range states {
		tax, effectiveRate := state.calcIncomeTax(income)
		state.incomeTax = tax
		state.effectiveRate = effectiveRate
	}
	return &states
}

func initializeFederal(income *float64, numSteps *int) State {
	federal := State{
		name:     "Federal",
		brackets: []int{0, 10275, 41775, 89075, 170050, 215950, 539900},
		rates:    []float64{0.10, 0.12, 0.22, 0.24, 0.32, 0.35, 0.37},
	}
	tax, effectiveRate := federal.calcIncomeTax(income)
	federal.incomeTax = tax
	federal.effectiveRate = effectiveRate
	return federal
}

func main() {
	income := flag.Float64("income", 0, "Annual taxable income")
	ascending := flag.Bool("ascending", false, "Sort the output in ascending order?")
	toCSV := flag.Bool("csv", false, "Write the output to a CSV file?")
	numSteps := flag.Int("steps", 100, "The number of discrete points between 0 and income for CSV output")
	flag.Parse()

	states := initializeStates(income, numSteps)
	federal := initializeFederal(income, numSteps)

	if *ascending {
		sort.SliceStable(states[:], func(i, j int) bool {
			return states[i].incomeTax < states[j].incomeTax
		})
	} else {
		sort.SliceStable(states[:], func(i, j int) bool {
			return states[i].incomeTax > states[j].incomeTax
		})
	}

	printResults(income, &federal, states)

	if *toCSV {
		writeToCSV(income, numSteps, &federal, states)
	}
}

func printResults(income *float64, federal *State, states *[51]*State) {
	fmt.Printf("\n50-State income tax report for income of $%.0f\n", *income)
	fmt.Println("    State                Tax       Effective Rate")
	fmt.Println("==================================================")
	fmt.Printf("*   %-20s $%-8d %.3f%%\n", federal.name, federal.incomeTax, 100*federal.effectiveRate)
	fmt.Println("==================================================")
	for i := 0; i < 51; i++ {
		fmt.Printf("%-3d %-20s $%-8d %.3f%%\n", i+1, states[i].name, states[i].incomeTax, 100*states[i].effectiveRate)
	}
	fmt.Println("==================================================")
}

func writeToCSV(income *float64, numSteps *int, federal *State, states *[51]*State) {
	// create an array of incomes sliced into `numSteps` steps
	incomeArray := *getIncomeArray(income, *numSteps)

	// create the 2D array at runtime with make()
	data := make([][]string, *numSteps+1)
	for i := range data {
		data[i] = make([]string, 53)
	}

	// add the label headers of income, [51]states+DC, Federal
	data[0][0] = "income"
	data[0][1] = "federal"
	for i := 0; i < len(states); i++ {
		// there are 53 columns: income + 50 states + DC + Federal
		data[0][i+2] = (*states)[i].abbrev
	}

	for i := 0; i < *numSteps; i++ {
		// add the income level for this row
		data[i+1][0] = strconv.FormatFloat(incomeArray[i], 'f', 2, 32)

		// add the federal effective rate for this income level
		_, rate := federal.calcIncomeTax(&incomeArray[i])
		data[i+1][1] = strconv.FormatFloat(rate, 'f', 6, 32)

		// add all 50 States' + DC's effective rate for this income level
		for j, state := range *states {
			_, rate := state.calcIncomeTax(&incomeArray[i])
			data[i+1][j+2] = strconv.FormatFloat(rate, 'f', 6, 32)
		}
	}
	path := fmt.Sprintf(
		"./output/csv/income=%.0f_steps=%d.csv", *income, *numSteps)
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file")
		panic(err)
	}
	w := csv.NewWriter(file)
	for _, record := range data {
		if err := w.Write(record); err != nil {
			panic(err)
		}
	}
	// Write any buffered data to the underlying writer (standard output).
	w.Flush()
	file.Close()
	if err := w.Error(); err != nil {
		panic(err)
	}

}

func getIncomeArray(income *float64, numSteps int) *[]float64 {
	stepSize := *income / float64(numSteps)
	incomes := make([]float64, numSteps)
	for i := 0; i < numSteps; i++ {
		incomes[i] = float64(stepSize * float64(i+1))
	}
	return &incomes
}

func (state *State) calcIncomeTax(income *float64) (int, float64) {
	numBrackets := len(state.brackets)
	tax := 0.0
	for i, bracket := range state.brackets {
		if i == numBrackets-1 {
			tax += math.Max(0, *income-float64(bracket)) * state.rates[i]
		} else {
			tax += math.Min(float64(state.brackets[i+1]-bracket), math.Max(0, *income-float64(bracket))) * state.rates[i]
		}
	}
	return int(tax), tax / float64(*income)
}
