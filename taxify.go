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

type FilingStatus struct {
	brackets          []int
	rates             []float64
	standardDeduction int
	personalExemption int
}

type State struct {
	name                 string
	abbrev               string
	dependentExemption   int
	dependentIsCredit    bool
	stdDeductionIsCredit bool
	exemptionIsCredit    bool
	incomeTypesTaxed     []float32 // *[1] see below
	single               FilingStatus
	couple               FilingStatus
	effectiveRate        float64
	incomeTax            int
}

// *[1] {ordinary, capital gains, dividends/interest} *negative means special case
// if capital gains is negative, a deduction of x is applied to capital gains before adding it to taxableIncome

type FedFilingStatus struct {
	// if dividends are qualified, they get added to capital gains instead of income
	incomeBrackets       []int
	incomeRates          []float64
	capitalGainsBrackets []int
	capitalGainsRates    []float64
	standardDeduction    int
}

type Federal struct {
	name               string
	abbrev             string
	medicareRate       float64 // 0.0145
	socialSecurityRate float64 // 0.062
	socialSecurityCap  int     // $147,000
	single             FedFilingStatus
	couple             FedFilingStatus
	effectiveRate      float64
	incomeTax          int
}

func initializeStates(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) *[51]*State {
	// TODO: move `exemptionIsCredit` and `stdDeductionIsCredit` out of FilingStatus since it's the same for both
	states := [51]*State{
		// {
		// 	name:                 "Alabama",
		// 	abbrev:               "AL",
		// 	dependentExemption:   1000,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{-1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 500, 3000},
		// 		rates:             []float64{0.02, 0.03, 0.05},
		// 		standardDeduction: 2500,
		// 		personalExemption: 1500,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 1000, 6000},
		// 		rates:             []float64{0.02, 0.03, 0.05},
		// 		standardDeduction: 7500,
		// 		personalExemption: 3000,
		// 	},
		// },
		// {
		// 	name:                 "Alaska",
		// 	abbrev:               "AK",
		// 	dependentExemption:   0,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{0.0, 0.0, 0.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0},
		// 		standardDeduction: 0,
		// 		personalExemption: 0,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0},
		// 		standardDeduction: 0,
		// 		personalExemption: 0,
		// 	},
		// },
		// {
		// 	name:                 "Arizona",
		// 	abbrev:               "AZ",
		// 	dependentExemption:   100,
		// 	dependentIsCredit:    true,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 27808, 55615, 116843},
		// 		rates:             []float64{0.0259, 0.0334, 0.0417, 0.045},
		// 		standardDeduction: 12950,
		// 		personalExemption: 0,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 55615, 111229, 333684},
		// 		rates:             []float64{0.0259, 0.0334, 0.0417, 0.045},
		// 		standardDeduction: 25900,
		// 		personalExemption: 0,
		// 	},
		// },
		// {
		// 	name:                 "Arkansas",
		// 	abbrev:               "AR",
		// 	dependentExemption:   29,
		// 	dependentIsCredit:    true,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    true,
		// 	incomeTypesTaxed:     []float32{1.0, -0.5, 1.0}, // only 50% of capital gains are taxed
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 4300, 8500},
		// 		rates:             []float64{0.02, 0.04, 0.055},
		// 		standardDeduction: 2200,
		// 		personalExemption: 29,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 4300, 8500},
		// 		rates:             []float64{0.02, 0.04, 0.055},
		// 		standardDeduction: 4400,
		// 		personalExemption: 58,
		// 	},
		// },
		// {
		// 	name:                 "California",
		// 	abbrev:               "CA",
		// 	dependentExemption:   400,
		// 	dependentIsCredit:    true,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    true,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 9325, 22107, 34892, 48435, 61214, 312686, 375221, 625369, 1000000},
		// 		rates:             []float64{0.01, 0.02, 0.04, 0.06, 0.08, 0.093, 0.103, 0.113, 0.123, 0.133},
		// 		standardDeduction: 4803,
		// 		personalExemption: 129,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 18650, 44214, 69784, 96870, 122428, 625372, 750442, 1000000, 1250738},
		// 		rates:             []float64{0.01, 0.02, 0.04, 0.06, 0.08, 0.093, 0.103, 0.113, 0.123, 0.133},
		// 		standardDeduction: 9606,
		// 		personalExemption: 258,
		// 	},
		// },
		// {
		// 	name:                 "Colorado",
		// 	abbrev:               "CO",
		// 	dependentExemption:   400,
		// 	dependentIsCredit:    true,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0455},
		// 		standardDeduction: 12950,
		// 		personalExemption: 0,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0455},
		// 		standardDeduction: 25900,
		// 		personalExemption: 0,
		// 	},
		// },
		// {
		// 	name:                 "Connecticut",
		// 	abbrev:               "CT",
		// 	dependentExemption:   0,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 0.07, 1.0}, // flat rate of 7% on capital gains
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 10000, 50000, 100000, 200000, 250000, 500000},
		// 		rates:             []float64{0.03, 0.05, 0.055, 0.06, 0.065, 0.069, 0.0699},
		// 		standardDeduction: 0,
		// 		personalExemption: 15000,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 20000, 100000, 200000, 400000, 500000, 1000000},
		// 		rates:             []float64{0.03, 0.05, 0.055, 0.06, 0.065, 0.069, 0.0699},
		// 		standardDeduction: 0,
		// 		personalExemption: 24000,
		// 	},
		// },
		// {
		// 	name:                 "Delaware",
		// 	abbrev:               "DE",
		// 	dependentExemption:   110,
		// 	dependentIsCredit:    true,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    true,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{2000, 5000, 10000, 20000, 25000, 60000},
		// 		rates:             []float64{0.022, 0.039, 0.048, 0.052, 0.0555, 0.066},
		// 		standardDeduction: 3250,
		// 		personalExemption: 110,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{2000, 5000, 10000, 20000, 25000, 60000},
		// 		rates:             []float64{0.022, 0.039, 0.048, 0.052, 0.0555, 0.066},
		// 		standardDeduction: 6500,
		// 		personalExemption: 220,
		// 	},
		// },
		// {
		// 	name:                 "Florida",
		// 	abbrev:               "FL",
		// 	dependentExemption:   0,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{0.0, 0.0, 0.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0},
		// 		standardDeduction: 0,
		// 		personalExemption: 0,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0},
		// 		standardDeduction: 0,
		// 		personalExemption: 0,
		// 	},
		// },
		// {
		// 	name:                 "Georgia",
		// 	abbrev:               "GA",
		// 	dependentExemption:   3000,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 750, 2250, 3750, 5250, 7000},
		// 		rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.0575},
		// 		standardDeduction: 5400,
		// 		personalExemption: 2700,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 1000, 3000, 5000, 7000, 10000},
		// 		rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.0575},
		// 		standardDeduction: 7100,
		// 		personalExemption: 7400,
		// 	},
		// },
		// {
		// 	name:                 "Hawaii",
		// 	abbrev:               "HI",
		// 	dependentExemption:   1144,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 0.0725, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 2400, 4800, 9600, 14400, 19200, 24000, 36000, 48000, 150000, 175000, 200000},
		// 		rates:             []float64{0.014, 0.032, 0.055, 0.064, 0.068, 0.072, 0.076, 0.079, 0.0825, 0.09, 0.1, 0.11},
		// 		standardDeduction: 2200,
		// 		personalExemption: 1144,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 4800, 9600, 19200, 28800, 38400, 48000, 72000, 96000, 300000, 350000, 400000},
		// 		rates:             []float64{0.014, 0.032, 0.055, 0.064, 0.068, 0.072, 0.076, 0.079, 0.0825, 0.09, 0.1, 0.11},
		// 		standardDeduction: 4400,
		// 		personalExemption: 2288,
		// 	},
		// },
		// {
		// 	name:                 "Idaho",
		// 	abbrev:               "ID",
		// 	dependentExemption:   0,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 1588, 4763, 7939},
		// 		rates:             []float64{0.01, 0.03, 0.045, 0.06},
		// 		standardDeduction: 12950,
		// 		personalExemption: 0,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 3176, 9526, 15878},
		// 		rates:             []float64{0.01, 0.03, 0.045, 0.06},
		// 		standardDeduction: 25900,
		// 		personalExemption: 0,
		// 	},
		// },
		// {
		// 	name:                 "Illinois",
		// 	abbrev:               "IL",
		// 	dependentExemption:   2375,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0495},
		// 		standardDeduction: 0,
		// 		personalExemption: 2375,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0495},
		// 		standardDeduction: 0,
		// 		personalExemption: 4750,
		// 	},
		// },
		{
		// 	name:                 "Indiana",
		// 	abbrev:               "IN",
		// 	dependentExemption:   1000,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0323},
		// 		standardDeduction: 0,
		// 		personalExemption: 1000,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.0323},
		// 		standardDeduction: 0,
		// 		personalExemption: 2000,
		// 	},
		// },
		// {
		// 	name:                 "Iowa",
		// 	abbrev:               "IA",
		// 	dependentExemption:   40,
		// 	dependentIsCredit:    true,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    true,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 1743, 3486, 6972, 15687, 26145, 34860, 52290, 78435},
		// 		rates:             []float64{0.0033, 0.0067, 0.0225, 0.0414, 0.0563, 0.0596, 0.0625, 0.0744, 0.0853},
		// 		standardDeduction: 2210,
		// 		personalExemption: 40,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 1743, 3486, 6972, 15687, 26145, 34860, 52290, 78435},
		// 		rates:             []float64{0.0033, 0.0067, 0.0225, 0.0414, 0.0563, 0.0596, 0.0625, 0.0744, 0.0853},
		// 		standardDeduction: 5450,
		// 		personalExemption: 80,
		// 	},
		// },
		// {
		// 	name:                 "Kansas",
		// 	abbrev:               "KS",
		// 	dependentExemption:   2250,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 15000, 30000},
		// 		rates:             []float64{0.031, 0.0525, 0.057},
		// 		standardDeduction: 3500,
		// 		personalExemption: 2250,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 30000, 60000},
		// 		rates:             []float64{0.031, 0.0525, 0.057},
		// 		standardDeduction: 8000,
		// 		personalExemption: 4500,
		// 	},
		// },
		// {
		// 	name:                 "Kentucky",
		// 	abbrev:               "KY",
		// 	dependentExemption:   0,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.050},
		// 		standardDeduction: 2770,
		// 		personalExemption: 0,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0},
		// 		rates:             []float64{0.050},
		// 		standardDeduction: 5540,
		// 		personalExemption: 0,
		// 	},
		// },
		// {
		// 	name:                 "Louisiana",
		// 	abbrev:               "LA",
		// 	dependentExemption:   1000,
		// 	dependentIsCredit:    false,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 12500, 50000},
		// 		rates:             []float64{0.0185, 0.035, 0.0425},
		// 		standardDeduction: 0,
		// 		personalExemption: 4500,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 25000, 100000},
		// 		rates:             []float64{0.0185, 0.035, 0.0425},
		// 		standardDeduction: 0,
		// 		personalExemption: 9000,
		// 	},
		// },
		// {
		// 	name:                 "Maine",
		// 	abbrev:               "ME",
		// 	dependentExemption:   300,
		// 	dependentIsCredit:    true,
		// 	stdDeductionIsCredit: false,
		// 	exemptionIsCredit:    false,
		// 	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
		// 	single: FilingStatus{
		// 		brackets:          []int{0, 23000, 54450},
		// 		rates:             []float64{0.058, 0.0675, 0.0715},
		// 		standardDeduction: 12950,
		// 		personalExemption: 4450,
		// 	},
		// 	couple: FilingStatus{
		// 		brackets:          []int{0, 46000, 108900},
		// 		rates:             []float64{0.058, 0.0675, 0.0715},
		// 		standardDeduction: 25900,
		// 		personalExemption: 8900,
		// 	},
		// },
		{
			name:                 "Maryland",
			abbrev:               "MD",
			dependentExemption:   3200,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 1000, 2000, 3000, 100000, 125000, 150000, 250000},
				rates:             []float64{0.02, 0.03, 0.04, 0.0475, 0.05, 0.0525, 0.055, 0.0575},
				standardDeduction: 2350,
				personalExemption: 3200,
			},
			couple: FilingStatus{
				brackets:          []int{0, 1000, 2000, 3000, 150000, 175000, 225000, 300000},
				rates:             []float64{0.02, 0.03, 0.04, 0.0475, 0.05, 0.0525, 0.055, 0.0575},
				standardDeduction: 4700,
				personalExemption: 6400,
			},
		},
		{
			name:                 "Massachusetts",
			abbrev:               "MA",
			dependentExemption:   1000,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.05},
				standardDeduction: 0,
				personalExemption: 4400,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.05},
				standardDeduction: 0,
				personalExemption: 8800,
			},
		},
		{
			name:                 "Michigan",
			abbrev:               "MI",
			dependentExemption:   5000,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0425},
				standardDeduction: 0,
				personalExemption: 5000,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0425},
				standardDeduction: 0,
				personalExemption: 10000,
			},
		},
		{
			name:                 "Minnesota",
			abbrev:               "MN",
			dependentExemption:   4450,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 28080, 92230, 171220},
				rates:             []float64{0.0535, 0.068, 0.0785, 0.0985},
				standardDeduction: 12900,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0, 41050, 163060, 284810},
				rates:             []float64{0.0535, 0.068, 0.0785, 0.0985},
				standardDeduction: 25800,
				personalExemption: 0,
			},
		},
		{
			name:                 "Mississippi",
			abbrev:               "MS",
			dependentExemption:   1500,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{5000, 10000},
				rates:             []float64{0.04, 0.05},
				standardDeduction: 2300,
				personalExemption: 6000,
			},
			couple: FilingStatus{
				brackets:          []int{5000, 10000},
				rates:             []float64{0.04, 0.05},
				standardDeduction: 4600,
				personalExemption: 12000,
			},
		},
		{
			name:                 "Missouri",
			abbrev:               "MO",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{108, 1088, 2176, 3264, 4352, 5440, 6528, 7616, 8704},
				rates:             []float64{0.015, 0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.054},
				standardDeduction: 12950,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{108, 1088, 2176, 3264, 4352, 5440, 6528, 7616, 8704},
				rates:             []float64{0.015, 0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.054},
				standardDeduction: 25900,
				personalExemption: 0,
			},
		},
		{
			name:                 "Montana",
			abbrev:               "MT",
			dependentExemption:   2580,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0}, // 2% credit on capital gains (ignored for now)
			single: FilingStatus{
				brackets:          []int{0, 3100, 5500, 8400, 11400, 14600, 18800},
				rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.0675},
				standardDeduction: 4830,
				personalExemption: 2580,
			},
			couple: FilingStatus{
				brackets:          []int{0, 3100, 5500, 8400, 11400, 14600, 18800},
				rates:             []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.0675},
				standardDeduction: 9660,
				personalExemption: 5160,
			},
		},
		{
			name:                 "Nebraska",
			abbrev:               "NE",
			dependentExemption:   146,
			dependentIsCredit:    true,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    true,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3440, 20590, 33180},
				rates:             []float64{0.0246, 0.0351, 0.0501, 0.0684},
				standardDeduction: 7350,
				personalExemption: 146,
			},
			couple: FilingStatus{
				brackets:          []int{0, 6860, 41190, 66360},
				rates:             []float64{0.0246, 0.0351, 0.0501, 0.0684},
				standardDeduction: 14700,
				personalExemption: 292,
			},
		},
		{
			name:                 "Nevada",
			abbrev:               "NV",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
		},
		{
			name:                 "New Hampshire",
			abbrev:               "NH",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{0.0, 0.0, 0.05},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 2400,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 4800,
			},
		},
		{
			name:                 "New Jersey",
			abbrev:               "NJ",
			dependentExemption:   1500,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 20000, 35000, 40000, 75000, 500000, 1000000},
				rates:             []float64{0.014, 0.0175, 0.035, 0.05525, 0.0637, 0.0897, 0.1075},
				standardDeduction: 0,
				personalExemption: 1000,
			},
			couple: FilingStatus{
				brackets:          []int{0, 20000, 50000, 70000, 80000, 150000, 500000, 1000000},
				rates:             []float64{0.014, 0.0175, 0.0245, 0.035, 0.05525, 0.0637, 0.0897, 0.1075},
				standardDeduction: 0,
				personalExemption: 2000,
			},
		},
		{
			name:                 "New Mexico",
			abbrev:               "NM",
			dependentExemption:   4000,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, -0.4, 1.0}, // 40% deduction of capital gains
			single: FilingStatus{
				brackets:          []int{0, 5500, 11000, 16000, 210000},
				rates:             []float64{0.017, 0.032, 0.047, 0.049, 0.059},
				standardDeduction: 12950,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0, 8000, 16000, 24000, 315000},
				rates:             []float64{0.017, 0.032, 0.047, 0.049, 0.059},
				standardDeduction: 25900,
				personalExemption: 0,
			},
		},
		{
			name:                 "New York",
			abbrev:               "NY",
			dependentExemption:   1000,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 8500, 11700, 13900, 80650, 215400, 1077550, 5000000, 25000000},
				rates:             []float64{0.04, 0.045, 0.0525, 0.0585, 0.0625, 0.0685, 0.0965, 0.103, 0.109},
				standardDeduction: 8000,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0, 17150, 23600, 27900, 161550, 323200, 2155350, 5000000, 25000000},
				rates:             []float64{0.04, 0.045, 0.0525, 0.0585, 0.0625, 0.0685, 0.0965, 0.103, 0.109},
				standardDeduction: 16050,
				personalExemption: 0,
			},
		},
		{
			name:                 "North Carolina",
			abbrev:               "NC",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0499},
				standardDeduction: 12750,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0499},
				standardDeduction: 25500,
				personalExemption: 0,
			},
		},
		{
			name:                 "North Dakota",
			abbrev:               "ND",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, -0.4, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 40525, 98100, 204675, 445000},
				rates:             []float64{0.011, 0.0204, 0.0227, 0.0264, 0.029},
				standardDeduction: 12950,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0, 67700, 163550, 249150, 445000},
				rates:             []float64{0.011, 0.0204, 0.0227, 0.0264, 0.029},
				standardDeduction: 25900,
				personalExemption: 0,
			},
		},
		{
			name:                 "Ohio",
			abbrev:               "OH",
			dependentExemption:   2400,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{25000, 44250, 88450, 110650},
				rates:             []float64{0.02765, 0.03226, 0.03688, 0.0399},
				standardDeduction: 0,
				personalExemption: 2400,
			},
			couple: FilingStatus{
				brackets:          []int{25000, 44250, 88450, 110650},
				rates:             []float64{0.02765, 0.03226, 0.03688, 0.0399},
				standardDeduction: 0,
				personalExemption: 4800,
			},
		},
		{
			name:                 "Oklahoma",
			abbrev:               "OK",
			dependentExemption:   1000,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 1000, 2500, 3750, 4900, 7200},
				rates:             []float64{0.0025, 0.0075, 0.0175, 0.0275, 0.0375, 0.0475},
				standardDeduction: 6350,
				personalExemption: 1000,
			},
			couple: FilingStatus{
				brackets:          []int{0, 2000, 5000, 7500, 9800, 12200},
				rates:             []float64{0.0025, 0.0075, 0.0175, 0.0275, 0.0375, 0.0475},
				standardDeduction: 12700,
				personalExemption: 2000,
			},
		},
		{
			name:                 "Oregon",
			abbrev:               "OR",
			dependentExemption:   219,
			dependentIsCredit:    true,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    true,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3650, 9200, 125000},
				rates:             []float64{0.0475, 0.0675, 0.0875, 0.099},
				standardDeduction: 2420,
				personalExemption: 219,
			},
			couple: FilingStatus{
				brackets:          []int{0, 7300, 18400, 250000},
				rates:             []float64{0.0475, 0.0675, 0.0875, 0.099},
				standardDeduction: 4840,
				personalExemption: 436,
			},
		},
		{
			name:                 "Pennsylvania",
			abbrev:               "PA",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0307},
				standardDeduction: 0,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0307},
				standardDeduction: 0,
				personalExemption: 0,
			},
		},
		{
			name:                 "Rhode Island",
			abbrev:               "RI",
			dependentExemption:   4350,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 68200, 155050},
				rates:             []float64{0.0375, 0.0475, 0.0599},
				standardDeduction: 9300,
				personalExemption: 4350,
			},
			couple: FilingStatus{
				brackets:          []int{0, 68200, 155050},
				rates:             []float64{0.0375, 0.0475, 0.0599},
				standardDeduction: 18600,
				personalExemption: 8700,
			},
		},
		{
			name:                 "South Carolina",
			abbrev:               "SC",
			dependentExemption:   4300,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, -0.44, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3200, 6410, 9620, 12820, 16040},
				rates:             []float64{0.0, 0.03, 0.04, 0.05, 0.06, 0.07},
				standardDeduction: 12950,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0, 3200, 6410, 9620, 12820, 16040},
				rates:             []float64{0.0, 0.03, 0.04, 0.05, 0.06, 0.07},
				standardDeduction: 25900,
				personalExemption: 0,
			},
		},
		{
			name:                 "South Dakota",
			abbrev:               "SD",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
		},
		{
			name:               "Tennessee",
			abbrev:             "TN",
			dependentExemption: 0,
			// dependentIsCredit:  false,
			// stdDeductionIsCredit:  false,
			// exemptionIsCredit:  false,
			incomeTypesTaxed: []float32{0.0, 0.0, 0.06},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
		},
		{
			name:                 "Texas",
			abbrev:               "TX",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
		},
		{
			name:                 "Utah",
			abbrev:               "UT",
			dependentExemption:   1750,
			dependentIsCredit:    true,
			stdDeductionIsCredit: true,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0495},
				standardDeduction: 777,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0495},
				standardDeduction: 1554,
				personalExemption: 0,
			},
		},
		{
			name:                 "Vermont",
			abbrev:               "VT",
			dependentExemption:   4350,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0}, // there's a special case here too (ignored for now)
			single: FilingStatus{
				brackets:          []int{0, 40950, 99200, 206950},
				rates:             []float64{0.0335, 0.066, 0.076, 0.0875},
				standardDeduction: 6350,
				personalExemption: 4350,
			},
			couple: FilingStatus{
				brackets:          []int{0, 68400, 165350, 251950},
				rates:             []float64{0.0335, 0.066, 0.076, 0.0875},
				standardDeduction: 12700,
				personalExemption: 8700,
			},
		},
		{
			name:                 "Virginia",
			abbrev:               "VA",
			dependentExemption:   930,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 3000, 5000, 17000},
				rates:             []float64{0.02, 0.03, 0.05, 0.0575},
				standardDeduction: 4500,
				personalExemption: 930,
			},
			couple: FilingStatus{
				brackets:          []int{0, 3000, 5000, 17000},
				rates:             []float64{0.02, 0.03, 0.05, 0.0575},
				standardDeduction: 9000,
				personalExemption: 1860,
			},
		},
		{
			name:                 "Washington",
			abbrev:               "WA",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{0.0, 0.07, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 250000,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 250000,
				personalExemption: 0,
			},
		},
		{
			name:                 "West Virginia",
			abbrev:               "WV",
			dependentExemption:   2000,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 10000, 25000, 40000, 60000},
				rates:             []float64{0.03, 0.04, 0.045, 0.06, 0.065},
				standardDeduction: 0,
				personalExemption: 2000,
			},
			couple: FilingStatus{
				brackets:          []int{0, 10000, 25000, 40000, 60000},
				rates:             []float64{0.03, 0.04, 0.045, 0.06, 0.065},
				standardDeduction: 0,
				personalExemption: 4000,
			},
		},
		{
			name:                 "Wisconsin",
			abbrev:               "WI",
			dependentExemption:   700,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 12760, 25520, 280950},
				rates:             []float64{0.0354, 0.0465, 0.053, 0.0765},
				standardDeduction: 11790,
				personalExemption: 700,
			},
			couple: FilingStatus{
				brackets:          []int{0, 17010, 34030, 374030},
				rates:             []float64{0.0354, 0.0465, 0.053, 0.0765},
				standardDeduction: 21820,
				personalExemption: 1400,
			},
		},
		{
			name:                 "Wyoming",
			abbrev:               "WY",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{0.0, 0.0, 0.0},
			single: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0},
				rates:             []float64{0.0},
				standardDeduction: 0,
				personalExemption: 0,
			},
		},
		{
			name:                 "Washington D.C.",
			abbrev:               "DC",
			dependentExemption:   0,
			dependentIsCredit:    false,
			stdDeductionIsCredit: false,
			exemptionIsCredit:    false,
			incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
			single: FilingStatus{
				brackets:          []int{0, 10000, 40000, 60000, 250000, 500000, 1000000},
				rates:             []float64{0.04, 0.06, 0.065, 0.085, 0.0925, 0.0975, 0.1075},
				standardDeduction: 12950,
				personalExemption: 0,
			},
			couple: FilingStatus{
				brackets:          []int{0, 10000, 40000, 60000, 250000, 500000, 1000000},
				rates:             []float64{0.04, 0.06, 0.065, 0.085, 0.0925, 0.0975, 0.1075},
				standardDeduction: 25900,
				personalExemption: 0,
			},
		},
	}
	for _, state := range states {
		state.incomeTax, state.effectiveRate = state.calcIncomeTax(
			income, capitalGains, dividends, federalTax, numDependents, mfj)
	}
	return &states
}

func initializeFederal(income, capitalGains, dividends *float64, mfj, qualified bool) *Federal {
	federal := Federal{
		name:               "Federal",
		abbrev:             "USA",
		medicareRate:       0.0145,
		socialSecurityRate: 0.062,
		socialSecurityCap:  147000, // of taxable income
		single: FedFilingStatus{
			incomeBrackets:       []int{0, 10275, 41775, 89075, 170050, 215950, 539900},
			incomeRates:          []float64{0.10, 0.12, 0.22, 0.24, 0.32, 0.35, 0.37},
			capitalGainsBrackets: []int{0, 41675, 459750},
			capitalGainsRates:    []float64{0.0, 0.15, 0.20},
			standardDeduction:    12950,
		},
		couple: FedFilingStatus{
			incomeBrackets:       []int{0, 20550, 83550, 178150, 340100, 431900, 647850},
			incomeRates:          []float64{0.10, 0.12, 0.22, 0.24, 0.32, 0.35, 0.37},
			capitalGainsBrackets: []int{0, 83350, 517200},
			capitalGainsRates:    []float64{0.0, 0.15, 0.20},
			standardDeduction:    25900,
		},
	}
	federal.incomeTax, federal.effectiveRate = federal.calcFederalIncomeTax(
		*income, *capitalGains, *dividends, mfj, qualified)
	return &federal
}

func main() {
	income := flag.Float64("income", 0, "Annual taxable income")
	capitalGains := flag.Float64("cg", 0, "Capital Gains earned")
	dividends := flag.Float64("interest", 0, "Dividends and interest earned")
	qualified := flag.Bool("qualified", false, "Are the dividends qualified? (default false)")
	toCSV := flag.Bool("csv", false, "Write the output to a CSV file?")
	numSteps := flag.Int("steps", 100, "The number of discrete points between 0 and income for CSV output")
	mfj := flag.Bool("joint", false, "Married filing jointly? (default false)")
	numDependents := flag.Int("dependents", 0, "number of dependents (default 0)")
	flag.Parse()

	federal := initializeFederal(income, capitalGains, dividends, *mfj, *qualified)
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
	filename := fmt.Sprintf(
		"./output/csv/income=%.0f_cg=%.0f_dividends=%.0f_qualified=%t_dependents=%d_mfj=%t_steps=%d.csv",
		income, capitalGains, dividends, qualified, numDependents, mfj, numSteps)
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

func (state *State) calcIncomeTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome, grossIncome := 0.0, (*income), ((*income) + (*capitalGains) + (*dividends))
	data := state.single
	if mfj {
		data = state.couple
	}

	dependentExemption := float64(state.dependentExemption * numDependents)
	if state.dependentIsCredit {
		// it's a direct credit. Subtract it from tax.
		// a negative is okay for now because it gets
		// checked in the second to last line of the func
		tax -= dependentExemption
	} else {
		taxableIncome -= dependentExemption
	}

	if state.stdDeductionIsCredit {
		tax -= float64(data.standardDeduction)
	} else {
		taxableIncome -= float64(data.standardDeduction)
	}

	if state.exemptionIsCredit {
		tax -= float64(data.personalExemption)
	} else {
		taxableIncome -= float64(data.personalExemption)
	}

	for i, val := range state.incomeTypesTaxed {
		// this deciphers the incomeTypesTaxed array and ensures that income, CG, and dividends
		// are correct for the given state after this runs.
		if val < float32(0) {
			// val is negative indicating a special case
			switch i {
			case 0:
				// it's one of 6 states where federal tax can be deducted from state income
				taxableIncome -= float64(federalTax)
			case 1:
				taxableIncome += (*capitalGains) * (1.0 - float64(val))
			}
		} else if val == float32(1) {
			// val is 1, meaning the category is taxed the same as ordinary income
			switch i {
			case 1:
				taxableIncome += (*capitalGains)
			case 2:
				taxableIncome += (*dividends)
			}
		} else {
			// there's a positive decimal value denoting a multiplier.
			// apply the multiple for the category and add it directly to the final tax.
			switch i {
			case 1:
				// we add to `tax`, not `taxableIncome` because these rates are specific
				tax += (*capitalGains) * float64(val)
			case 2:
				tax += (*dividends) * float64(val)
			}
		}
	}
	tax += taxEngine(&taxableIncome, &data.brackets, &data.rates)
	tax = math.Max(0, tax) // assert tax >= 0
	return int(tax), tax / grossIncome
}

func (federal *Federal) calcFederalIncomeTax(
	income, capitalGains, dividends float64,
	mfj, qualified bool) (int, float64) {
	tax := 0.0
	data := federal.single
	if mfj {
		data = federal.couple
	}
	if qualified {
		capitalGains += dividends
	} else {
		income += dividends
	}
	grossIncome := income + capitalGains + dividends
	income -= float64(data.standardDeduction)
	income = math.Max(0.0, income)
	tax += income * federal.medicareRate
	tax += taxEngine(&income, &data.incomeBrackets, &data.incomeRates)
	tax += taxEngine(&capitalGains, &data.capitalGainsBrackets, &data.capitalGainsRates)
	ssCappedIncome := math.Min(float64(federal.socialSecurityCap), income)
	tax += ssCappedIncome * federal.socialSecurityRate
	return int(tax), tax / grossIncome
}

func taxEngine(income *float64, brackets *[]int, rates *[]float64) float64 {
	// todo take deductions and credits into account...
	tax := 0.0
	numBrackets := len(*brackets)
	for i, bracket := range *brackets {
		if i == numBrackets-1 {
			tax += math.Max(0, (*income)-float64(bracket)) * (*rates)[i]
		} else {
			tax += math.Min(float64((*brackets)[i+1]-bracket), math.Max(0, (*income)-float64(bracket))) * (*rates)[i]
		}
	}
	return tax
}
