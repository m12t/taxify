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
	name               string
	abbrev             string
	dependentExemption int
	dependentIsCredit  bool
	incomeTypesTaxed   []float32 // ** see below
	single             FilingStatus
	couple             FilingStatus
	effectiveRate      float64
	incomeTax          int
}

// ** {ordinary, capital gains, dividends/interest} *negative means special case
// if ordinary is negative, it's one of 6 states that allow federal taxes to be deducted from state
// if capital gains is negative, a deduction of x is applied to capital gains before adding it to income

type FilingStatus struct {
	brackets          []int
	rates             []float64
	standardDeduction int
	deductionIsCredit bool
	personalExemption int
	exemptionIsCredit bool
}

func initializeStates(income, dividends, capitalGains *float64, numSteps *int, mfj bool) *[51]*State {
	states := [51]*State{
		{
			name:               "Alabama",
			abbrev:             "AL",
			dependentExemption: 1000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{-1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 500, 3000},
				rates:             []float64{0.02, 0.03, 0.05},
				standardDeduction: 2500,
				deductionIsCredit: false,
				personalExemption: 1500,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 1000, 6000},
				rates:             []float64{0.02, 0.03, 0.05},
				standardDeduction: 7500,
				deductionIsCredit: false,
				personalExemption: 3000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Alaska",
			abbrev:             "AK",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Arizona",
			abbrev:             "AZ",
			dependentExemption: 100,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 27808, 55615, 116843},
				rates:             []float64{0.0259, 0.0334, 0.0417, 0.045},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 55615, 111229, 333684},
				rates:             []float64{0.0259, 0.0334, 0.0417, 0.045},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Arkansas",
			abbrev:             "AR",
			dependentExemption: 29,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, -0.5, 1.0}, // only 50% of capital gains are taxed
			single: FilingStatus{
				brackets:          []int{0, 4300, 8500},
				rates:             []float64{0.02, 0.04, 0.055},
				standardDeduction: 2200,
				deductionIsCredit: false,
				personalExemption: 29,
				exemptionIsCredit: true,
			},
			couple: FilingStatus{
				brackets:          []int{0, 4300, 8500},
				rates:             []float64{0.02, 0.04, 0.055},
				standardDeduction: 4400,
				deductionIsCredit: false,
				personalExemption: 58,
				exemptionIsCredit: true,
			},
		},
		{
			name:               "California",
			abbrev:             "CA",
			dependentExemption: 400,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 9325, 22107, 34892, 48435, 61214, 312686, 375221, 625369, 1000000},
				rates:             []float64{0.01, 0.02, 0.04, 0.06, 0.08, 0.093, 0.103, 0.113, 0.123, 0.133},
				standardDeduction: 4803,
				deductionIsCredit: false,
				personalExemption: 129,
				exemptionIsCredit: true,
			},
			couple: FilingStatus{
				brackets:          []int{0, 18650, 44214, 69784, 96870, 122428, 625372, 750442, 1000000, 1250738},
				rates:             []float64{0.01, 0.02, 0.04, 0.06, 0.08, 0.093, 0.103, 0.113, 0.123, 0.133},
				standardDeduction: 9606,
				deductionIsCredit: false,
				personalExemption: 258,
				exemptionIsCredit: true,
			},
		},
		{
			name:               "Colorado",
			abbrev:             "CO",
			dependentExemption: 400,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0455},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0455},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Connecticut",
			abbrev:             "CT",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 0.07, 1.0}, // flat rate of 7% on capital gains
			single: FilingStatus{
				brackets:          []int{0, 10000, 50000, 100000, 200000, 250000, 500000},
				rates:             []float64{0.03, 0.05, 0.055, 0.06, 0.065, 0.069, 0.0699},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 15000,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 20000, 100000, 200000, 400000, 500000, 1000000},
				rates:             []float64{0.03, 0.05, 0.055, 0.06, 0.065, 0.069, 0.0699},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 24000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Delaware",
			abbrev:             "DE",
			dependentExemption: 110,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{2000, 5000, 10000, 20000, 25000, 60000},
				rates:             []float64{0.022, 0.039, 0.048, 0.052, 0.0555, 0.066},
				standardDeduction: 3250,
				deductionIsCredit: false,
				personalExemption: 110,
				exemptionIsCredit: true,
			},
			couple: FilingStatus{
				brackets:          []int{2000, 5000, 10000, 20000, 25000, 60000},
				rates:             []float64{0.022, 0.039, 0.048, 0.052, 0.0555, 0.066},
				standardDeduction: 6500,
				deductionIsCredit: false,
				personalExemption: 220,
				exemptionIsCredit: true,
			},
		},
		{
			name:               "Florida",
			abbrev:             "FL",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Georgia",
			abbrev:             "GA",
			dependentExemption: 3000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 750, 2250, 3750, 5250, 7000},
				rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.0575},
				standardDeduction: 5400,
				deductionIsCredit: false,
				personalExemption: 2700,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 1000, 3000, 5000, 7000, 10000},
				rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.0575},
				standardDeduction: 7100,
				deductionIsCredit: false,
				personalExemption: 7400,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Hawaii",
			abbrev:             "HI",
			dependentExemption: 1144,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 0.0725, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 2400, 4800, 9600, 14400, 19200, 24000, 36000, 48000, 150000, 175000, 200000},
				rates:             []float64{0.014, 0.032, 0.055, 0.064, 0.068, 0.072, 0.076, 0.079, 0.0825, 0.09, 0.1, 0.11},
				standardDeduction: 2200,
				deductionIsCredit: false,
				personalExemption: 1144,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 4800, 9600, 19200, 28800, 38400, 48000, 72000, 96000, 300000, 350000, 400000},
				rates:             []float64{0.014, 0.032, 0.055, 0.064, 0.068, 0.072, 0.076, 0.079, 0.0825, 0.09, 0.1, 0.11},
				standardDeduction: 4400,
				deductionIsCredit: false,
				personalExemption: 2288,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Idaho",
			abbrev:             "ID",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 1588, 4763, 7939},
				rates:             []float64{0.01, 0.03, 0.045, 0.06},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 3176, 9526, 15878},
				rates:             []float64{0.01, 0.03, 0.045, 0.06},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Illinois",
			abbrev:             "IL",
			dependentExemption: 2375,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0495},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 2375,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0495},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 4750,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Indiana",
			abbrev:             "IN",
			dependentExemption: 1000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0323},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 1000,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0323},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 2000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Iowa",
			abbrev:             "IA",
			dependentExemption: 40,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{-1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 1743, 3486, 6972, 15687, 26145, 34860, 52290, 78435},
				rates:             []float64{0.0033, 0.0067, 0.0225, 0.0414, 0.0563, 0.0596, 0.0625, 0.0744, 0.0853},
				standardDeduction: 2210,
				deductionIsCredit: false,
				personalExemption: 40,
				exemptionIsCredit: true,
			},
			couple: FilingStatus{
				brackets:          []int{0, 1743, 3486, 6972, 15687, 26145, 34860, 52290, 78435},
				rates:             []float64{0.0033, 0.0067, 0.0225, 0.0414, 0.0563, 0.0596, 0.0625, 0.0744, 0.0853},
				standardDeduction: 5450,
				deductionIsCredit: false,
				personalExemption: 80,
				exemptionIsCredit: true,
			},
		},
		{
			name:               "Kansas",
			abbrev:             "KS",
			dependentExemption: 2250,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 15000, 30000},
				rates:             []float64{0.031, 0.0525, 0.057},
				standardDeduction: 3500,
				deductionIsCredit: false,
				personalExemption: 2250,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 30000, 60000},
				rates:             []float64{0.031, 0.0525, 0.057},
				standardDeduction: 8000,
				deductionIsCredit: false,
				personalExemption: 4500,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Kentucky",
			abbrev:             "KY",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.050},
				standardDeduction: 2770,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.050},
				standardDeduction: 5540,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Louisiana",
			abbrev:             "LA",
			dependentExemption: 1000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{-1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 12500, 50000},
				rates:             []float64{0.0185, 0.035, 0.0425},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 4500,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 25000, 100000},
				rates:             []float64{0.0185, 0.035, 0.0425},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 9000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Maine",
			abbrev:             "ME",
			dependentExemption: 300,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 23000, 54450},
				rates:             []float64{0.058, 0.0675, 0.0715},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 4450,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 46000, 108900},
				rates:             []float64{0.058, 0.0675, 0.0715},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 8900,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Maryland",
			abbrev:             "MD",
			dependentExemption: 3200,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 1000, 2000, 3000, 100000, 125000, 150000, 250000},
				rates:             []float64{0.02, 0.03, 0.04, 0.0475, 0.05, 0.0525, 0.055, 0.0575},
				standardDeduction: 2350,
				deductionIsCredit: false,
				personalExemption: 3200,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 1000, 2000, 3000, 150000, 175000, 225000, 300000},
				rates:             []float64{0.02, 0.03, 0.04, 0.0475, 0.05, 0.0525, 0.055, 0.0575},
				standardDeduction: 4700,
				deductionIsCredit: false,
				personalExemption: 6400,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Massachusetts",
			abbrev:             "MA",
			dependentExemption: 1000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.05},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 4400,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.05},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 8800,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Michigan",
			abbrev:             "MI",
			dependentExemption: 5000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0425},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 5000,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0425},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 10000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Minnesota",
			abbrev:             "MN",
			dependentExemption: 4450,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 28080, 92230, 171220},
				rates:             []float64{0.0535, 0.068, 0.0785, 0.0985},
				standardDeduction: 12900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 41050, 163060, 284810},
				rates:             []float64{0.0535, 0.068, 0.0785, 0.0985},
				standardDeduction: 25800,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Mississippi",
			abbrev:             "MS",
			dependentExemption: 1500,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{5000, 10000},
				rates:             []float64{0.04, 0.05},
				standardDeduction: 2300,
				deductionIsCredit: false,
				personalExemption: 6000,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{5000, 10000},
				rates:             []float64{0.04, 0.05},
				standardDeduction: 4600,
				deductionIsCredit: false,
				personalExemption: 12000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Missouri",
			abbrev:             "MO",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{-1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{108, 1088, 2176, 3264, 4352, 5440, 6528, 7616, 8704},
				rates:             []float64{0.015, 0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.054},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{108, 1088, 2176, 3264, 4352, 5440, 6528, 7616, 8704},
				rates:             []float64{0.015, 0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.054},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Montana",
			abbrev:             "MT",
			dependentExemption: 2580,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{-1.0, 1.0, 1.0}, // 2% credit on capital gains (ignored for now)
			single: FilingStatus{
				brackets:          []int{0, 3100, 5500, 8400, 11400, 14600, 18800},
				rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.0675},
				standardDeduction: 4830,
				deductionIsCredit: false,
				personalExemption: 2580,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 3100, 5500, 8400, 11400, 14600, 18800},
				rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.0675},
				standardDeduction: 9660,
				deductionIsCredit: false,
				personalExemption: 5160,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Nebraska",
			abbrev:             "NE",
			dependentExemption: 146,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3440, 20590, 33180},
				rates:             []float64{0.0246, 0.0351, 0.0501, 0.0684},
				standardDeduction: 7350,
				deductionIsCredit: false,
				personalExemption: 146,
				exemptionIsCredit: true,
			},
			couple: FilingStatus{
				brackets:          []int{0, 6860, 41190, 66360},
				rates:             []float64{0.0246, 0.0351, 0.0501, 0.0684},
				standardDeduction: 14700,
				deductionIsCredit: false,
				personalExemption: 292,
				exemptionIsCredit: true,
			},
		},
		{
			name:               "Nevada",
			abbrev:             "NV",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "New Hampshire",
			abbrev:             "NH",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.05},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 2400,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 4800,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "New Jersey",
			abbrev:             "NJ",
			dependentExemption: 1500,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 20000, 35000, 40000, 75000, 500000, 1000000},
				rates:             []float64{0.014, 0.0175, 0.035, 0.05525, 0.0637, 0.0897, 0.1075},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 1000,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 20000, 50000, 70000, 80000, 150000, 500000, 1000000},
				rates:             []float64{0.014, 0.0175, 0.0245, 0.035, 0.05525, 0.0637, 0.0897, 0.1075},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 2000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "New Mexico",
			abbrev:             "NM",
			dependentExemption: 4000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, -0.4, 1.0}, // 40% deduction of capital gains
			single: FilingStatus{
				brackets:          []int{0, 5500, 11000, 16000, 210000},
				rates:             []float64{0.017, 0.032, 0.047, 0.049, 0.059},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 8000, 16000, 24000, 315000},
				rates:             []float64{0.017, 0.032, 0.047, 0.049, 0.059},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "New York",
			abbrev:             "NY",
			dependentExemption: 1000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 8500, 11700, 13900, 80650, 215400, 1077550, 5000000, 25000000},
				rates:             []float64{0.04, 0.045, 0.0525, 0.0585, 0.0625, 0.0685, 0.0965, 0.103, 0.109},
				standardDeduction: 8000,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 17150, 23600, 27900, 161550, 323200, 2155350, 5000000, 25000000},
				rates:             []float64{0.04, 0.045, 0.0525, 0.0585, 0.0625, 0.0685, 0.0965, 0.103, 0.109},
				standardDeduction: 16050,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "North Carolina",
			abbrev:             "NC",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0499},
				standardDeduction: 12750,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0499},
				standardDeduction: 25500,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "North Dakota",
			abbrev:             "ND",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, -0.4, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 40525, 98100, 204675, 445000},
				rates:             []float64{0.011, 0.0204, 0.0227, 0.0264, 0.029},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 67700, 163550, 249150, 445000},
				rates:             []float64{0.011, 0.0204, 0.0227, 0.0264, 0.029},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Ohio",
			abbrev:             "OH",
			dependentExemption: 2400,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{25000, 44250, 88450, 110650},
				rates:             []float64{0.02765, 0.03226, 0.03688, 0.0399},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 2400,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{25000, 44250, 88450, 110650},
				rates:             []float64{0.02765, 0.03226, 0.03688, 0.0399},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 4800,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Oklahoma",
			abbrev:             "OK",
			dependentExemption: 1000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 1000, 2500, 3750, 4900, 7200},
				rates:             []float64{0.0025, 0.0075, 0.0175, 0.0275, 0.0375, 0.0475},
				standardDeduction: 6350,
				deductionIsCredit: false,
				personalExemption: 1000,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 2000, 5000, 7500, 9800, 12200},
				rates:             []float64{0.0025, 0.0075, 0.0175, 0.0275, 0.0375, 0.0475},
				standardDeduction: 12700,
				deductionIsCredit: false,
				personalExemption: 2000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Oregon",
			abbrev:             "OR",
			dependentExemption: 219,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{-1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3650, 9200, 125000},
				rates:             []float64{0.0475, 0.0675, 0.0875, 0.099},
				standardDeduction: 2420,
				deductionIsCredit: false,
				personalExemption: 219,
				exemptionIsCredit: true,
			},
			couple: FilingStatus{
				brackets:          []int{0, 7300, 18400, 250000},
				rates:             []float64{0.0475, 0.0675, 0.0875, 0.099},
				standardDeduction: 4840,
				deductionIsCredit: false,
				personalExemption: 436,
				exemptionIsCredit: true,
			},
		},
		{
			name:               "Pennsylvania",
			abbrev:             "PA",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0307},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0307},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Rhode Island",
			abbrev:             "RI",
			dependentExemption: 4350,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 68200, 155050},
				rates:             []float64{0.0375, 0.0475, 0.0599},
				standardDeduction: 9300,
				deductionIsCredit: false,
				personalExemption: 4350,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 68200, 155050},
				rates:             []float64{0.0375, 0.0475, 0.0599},
				standardDeduction: 18600,
				deductionIsCredit: false,
				personalExemption: 8700,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "South Carolina",
			abbrev:             "SC",
			dependentExemption: 4300,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, -0.44, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3200, 6410, 9620, 12820, 16040},
				rates:             []float64{0.0, 0.03, 0.04, 0.05, 0.06, 0.07},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 3200, 6410, 9620, 12820, 16040},
				rates:             []float64{0.0, 0.03, 0.04, 0.05, 0.06, 0.07},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "South Dakota",
			abbrev:             "SD",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Tennessee",
			abbrev:             "TN",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.06},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Texas",
			abbrev:             "TX",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Utah",
			abbrev:             "UT",
			dependentExemption: 1750,
			dependentIsCredit:  true,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0495},
				standardDeduction: 777,
				deductionIsCredit: true,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0495},
				standardDeduction: 1554,
				deductionIsCredit: true,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Vermont",
			abbrev:             "VT",
			dependentExemption: 4350,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0}, // there's a special case here too (ignored for now)
			single: FilingStatus{
				brackets:          []int{0, 40950, 99200, 206950},
				rates:             []float64{0.0335, 0.066, 0.076, 0.0875},
				standardDeduction: 6350,
				deductionIsCredit: false,
				personalExemption: 4350,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 68400, 165350, 251950},
				rates:             []float64{0.0335, 0.066, 0.076, 0.0875},
				standardDeduction: 12700,
				deductionIsCredit: false,
				personalExemption: 8700,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Virginia",
			abbrev:             "VA",
			dependentExemption: 930,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3000, 5000, 17000},
				rates:             []float64{0.02, 0.03, 0.05, 0.0575},
				standardDeduction: 4500,
				deductionIsCredit: false,
				personalExemption: 930,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 3000, 5000, 17000},
				rates:             []float64{0.02, 0.03, 0.05, 0.0575},
				standardDeduction: 9000,
				deductionIsCredit: false,
				personalExemption: 1860,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Washington",
			abbrev:             "WA",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.07, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 250000,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 250000,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "West Virginia",
			abbrev:             "WV",
			dependentExemption: 2000,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 10000, 25000, 40000, 60000},
				rates:             []float64{0.03, 0.04, 0.045, 0.06, 0.065},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 2000,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 10000, 25000, 40000, 60000},
				rates:             []float64{0.03, 0.04, 0.045, 0.06, 0.065},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 4000,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Wisconsin",
			abbrev:             "WI",
			dependentExemption: 700,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 12760, 25520, 280950},
				rates:             []float64{0.0354, 0.0465, 0.053, 0.0765},
				standardDeduction: 11790,
				deductionIsCredit: false,
				personalExemption: 700,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 17010, 34030, 374030},
				rates:             []float64{0.0354, 0.0465, 0.053, 0.0765},
				standardDeduction: 21820,
				deductionIsCredit: false,
				personalExemption: 1400,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Wyoming",
			abbrev:             "WY",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
		{
			name:               "Washington D.C.",
			abbrev:             "DC",
			dependentExemption: 0,
			dependentIsCredit:  false,
			incomeTypesTaxed:   []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 10000, 40000, 60000, 250000, 500000, 1000000},
				rates:             []float64{0.04, 0.06, 0.065, 0.085, 0.0925, 0.0975, 0.1075},
				standardDeduction: 12950,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
			couple: FilingStatus{
				brackets:          []int{0, 10000, 40000, 60000, 250000, 500000, 1000000},
				rates:             []float64{0.04, 0.06, 0.065, 0.085, 0.0925, 0.0975, 0.1075},
				standardDeduction: 25900,
				deductionIsCredit: false,
				personalExemption: 0,
				exemptionIsCredit: false,
			},
		},
	}
	for _, state := range states {
		tax, effectiveRate := state.calcIncomeTax(income, dividends, capitalGains)
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
	toCSV := flag.Bool("csv", false, "Write the output to a CSV file?")
	numSteps := flag.Int("steps", 100, "The number of discrete points between 0 and income for CSV output")
	flag.Parse()

	states := initializeStates(income, numSteps)
	federal := initializeFederal(income, numSteps)

	sort.SliceStable(states[:], func(i, j int) bool {
		return states[i].effectiveRate > states[j].effectiveRate
	})

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
	file, err := os.Create(fmt.Sprintf("./output/csv/income=%.0f_steps=%d.csv", *income, *numSteps))
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

func getIncomeArray(income *float64, numSteps int) *[]float64 {
	stepSize := *income / float64(numSteps)
	incomes := make([]float64, numSteps)
	for i := 0; i < numSteps; i++ {
		incomes[i] = float64(stepSize * float64(i+1))
	}
	return &incomes
}

func (state *State) calcIncomeTax(income, dividends, capitalGains *float64, mfj bool) (int, float64) {
	data := state.single
	if mfj {
		data = state.couple
	}
	numBrackets := len(data.brackets)
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
