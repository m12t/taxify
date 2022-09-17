/*
A simple CLI program for estimating one's state income tax in all 50 states at once

Resources:
https://taxfoundation.org/state-income-tax-rates-2022/
*/

package main

import (
	"crypto/md5"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
)

/*
type State struct {
	name    string
	handler func()
}

func a() {
	fmt.Println("hi")
}

func main() {
	ca := State{}
	ca.name = "CA"
	ca.handler = a
	ca.handler()
	tx := State{
		name:    "TX",
		handler: a,
	}
	tx.handler()
*/

type State struct {
	name          string
	abbrev        string
	calcTax       func()
	effectiveRate float64
	incomeTax     int
}

func initializeStates(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) *[51]*State {
	// TODO: move `exemptionIsCredit` and `stdDeductionIsCredit` out of FilingStatus since it's the same for both
	states := [51]*State{
		{
			name:    "Alabama",
			abbrev:  "AL",
			calcTax: calcAlabamaTax,
		},
		{
			name:    "Alaska",
			abbrev:  "AK",
			calcTax: calcAlaskaTax,
		},
		{
			name:    "Arizona",
			abbrev:  "AZ",
			calcTax: calcArizonaTax,
		},
		{
			name:    "Arkansas",
			abbrev:  "AR",
			calcTax: calcArkansasTax,
		},
		{
			name:    "California",
			abbrev:  "CA",
			calcTax: calcCaliforniaTax,
		},
		{
			name:    "Colorado",
			abbrev:  "CO",
			calcTax: calcColoradoTax,
		},
		{
			name:    "Connecticut",
			abbrev:  "CT",
			calcTax: calcConnecticutTax,
		},
		{
			name:   "Delaware",
			abbrev: "DE",
		},
		{
			name:   "Florida",
			abbrev: "FL",
		},
		{
			name:   "Georgia",
			abbrev: "GA",
		},
		{
			name:   "Hawaii",
			abbrev: "HI",
		},
		{
			name:   "Idaho",
			abbrev: "ID",
		},
		{
			name:   "Illinois",
			abbrev: "IL",
		},
		{
			name:   "Indiana",
			abbrev: "IN",
		},
		{
			name:   "Iowa",
			abbrev: "IA",
		},
		{
			name:   "Kansas",
			abbrev: "KS",
		},
		{
			name:   "Kentucky",
			abbrev: "KY",
		},
		{
			name:   "Louisiana",
			abbrev: "LA",
		},
		{
			name:   "Maine",
			abbrev: "ME",
		},
		{
			name:   "Maryland",
			abbrev: "MD",
		},
		{
			name:   "Massachusetts",
			abbrev: "MA",
		},
		{
			name:   "Michigan",
			abbrev: "MI",
		},
		{
			name:   "Minnesota",
			abbrev: "MN",
		},
		{
			name:   "Mississippi",
			abbrev: "MS",
		},
		{
			name:   "Missouri",
			abbrev: "MO",
		},
		{
			name:   "Montana",
			abbrev: "MT",
		},
		{
			name:   "Nebraska",
			abbrev: "NE",
		},
		{
			name:   "Nevada",
			abbrev: "NV",
		},
		{
			name:   "New Hampshire",
			abbrev: "NH",
		},
		{
			name:   "New Jersey",
			abbrev: "NJ",
		},
		{
			name:   "New Mexico",
			abbrev: "NM",
		},
		{
			name:   "New York",
			abbrev: "NY",
		},
		{
			name:   "North Carolina",
			abbrev: "NC",
		},
		{
			name:   "North Dakota",
			abbrev: "ND",
		},
		{
			name:   "Ohio",
			abbrev: "OH",
		},
		{
			name:   "Oklahoma",
			abbrev: "OK",
		},
		{
			name:   "Oregon",
			abbrev: "OR",
		},
		{
			name:   "Pennsylvania",
			abbrev: "PA",
		},
		{
			name:   "Rhode Island",
			abbrev: "RI",
		},
		{
			name:   "South Carolina",
			abbrev: "SC",
		},
		{
			name:   "South Dakota",
			abbrev: "SD",
		},
		{
			name:   "Tennessee",
			abbrev: "TN",
		},
		{
			name:   "Texas",
			abbrev: "TX",
		},
		{
			name:   "Utah",
			abbrev: "UT",
		},
		{
			name:   "Vermont",
			abbrev: "VT",
		},
		{
			name:   "Virginia",
			abbrev: "VA",
		},
		{
			name:   "Washington",
			abbrev: "WA",
		},
		{
			name:   "West Virginia",
			abbrev: "WV",
		},
		{
			name:   "Wisconsin",
			abbrev: "WI",
		},
		{
			name:   "Wyoming",
			abbrev: "WY",
		},
		{
			name:   "Washington D.C.",
			abbrev: "DC",
		},
	}
	for _, state := range states {
		state.incomeTax, state.effectiveRate = state.calcTax(
			income, capitalGains, dividends, federalTax, numDependents, mfj)
	}
	return &states
}

func main() {
	income := flag.Float64("income", 0, "Annual taxable income")
	capitalGains := flag.Float64("cg", 0, "Capital Gains earned")
	dividends := flag.Float64("interest", 0, "Dividends and interest earned")
	qualified := flag.Bool("qualified", false, "Are the dividends qualified? (federal only, default false)")
	toCSV := flag.Bool("csv", false, "Write the output to a CSV file?")
	numSteps := flag.Int("steps", 100, "The number of discrete points between 0 and income for CSV output")
	mfj := flag.Bool("joint", false, "Married filing jointly? (default false)")
	numDependents := flag.Int("dependents", 0, "number of dependents (default 0)")
	flag.Parse()

	federal := initializeFederal(income, capitalGains, dividends, *mfj, *qualified)
	federal := State{
		name:    "Federal",
		abbrev:  "FED",
		calcTax: calcFederalTax,
	}
	states := initializeStates(income, capitalGains, dividends, federal.incomeTax, *numDependents, *mfj)

	sort.SliceStable(states[:], func(i, j int) bool {
		return states[i].effectiveRate > states[j].effectiveRate
	})

	printResults(income, federal, states)

	if *toCSV {
		writeToCSV(*income, *capitalGains, *dividends, *numSteps,
			*numDependents, federal, states, *mfj, *qualified)
	}
}

func printResults(income *float64, federal *Federal, states *[51]*State) {
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

func writeToCSV(income, capitalGains, dividends float64,
	numSteps, numDependents int, federal *Federal, states *[51]*State, mfj, qualified bool) {
	// create an array of incomes sliced into `numSteps` steps
	incomeArray := *getIncomeArray(income, numSteps)
	capitalGainsArray := *getIncomeArray(capitalGains, numSteps)
	dividendsArray := *getIncomeArray(dividends, numSteps)

	// create the 2D array at runtime with make()
	data := make([][]string, numSteps+1)
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

	for i := 0; i < numSteps; i++ {
		// add the income level for this row
		data[i+1][0] = strconv.FormatFloat(incomeArray[i], 'f', 2, 32)

		// add the federal effective rate for this income level
		// calcFederalIncomeTax(income, capitalGains, dividends *float64, mfj, qualified bool) (int, float64)
		federalTax, rate := federal.calcFederalIncomeTax(incomeArray[i], capitalGains, dividends, mfj, qualified)
		data[i+1][1] = strconv.FormatFloat(rate, 'f', 6, 32)

		// add all 50 States' + DC's effective rate for this income level
		for j, state := range *states {
			_, rate := state.calcIncomeTax(
				&incomeArray[i], &capitalGainsArray[i], &dividendsArray[i],
				federalTax, numDependents, mfj)
			data[i+1][j+2] = strconv.FormatFloat(rate, 'f', 6, 32)
		}
	}
	// filenames are getting long... could use some encoding to reduce this... md5 checksum?
	rawStr := fmt.Sprintf("%.0f_%.0f_%.0f_%t_%d_%t_%d", income, capitalGains, dividends, qualified, numDependents, mfj, numSteps)
	msg := []byte(rawStr)
	hash := md5.New()
	filename := fmt.Sprintf("%s.csv", string(hash.Sum(msg)[:8]))
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	for _, record := range data {
		if err := w.Write(record); err != nil {
			panic(err)
		}
	}
	// Write any buffered data to the underlying writer (standard output).
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
}

func getIncomeArray(income float64, numSteps int) *[]float64 {
	stepSize := income / float64(numSteps)
	incomes := make([]float64, numSteps)
	for i := 0; i < numSteps; i++ {
		incomes[i] = float64(stepSize * float64(i+1))
	}
	return &incomes
}
