// A store for all tax codes encoded in functions
package codes

import (
	"math"
)

func taxEngine(income *float64, brackets *[]int, rates *[]float64) float64 {
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

/* --------------------- state codes --------------------- */

// Resources:
// https://itep.sfo2.digitaloceanspaces.com/pb51fedinc.pdf
func calcAlabamaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 500, 3000}
	rates := []float64{0.02, 0.03, 0.05}
	dependentExemption := 1000
	standardDeduction := 2500
	personalExemption := 1500
	if mfj {
		brackets = []int{0, 1000, 6000}
		standardDeduction = 7500
		personalExemption = 3000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)
	taxableIncome -= float64(federalTax) // can deduct 100% of federal tax paid from taxableIncome

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// No income tax of any kind
func calcAlaskaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	return 0, 0.0
}

func calcArizonaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 27808, 55615, 116843}
	rates := []float64{0.0259, 0.0334, 0.0417, 0.045}
	dependentExemption := 100
	standardDeduction := 12950
	if mfj {
		brackets = []int{0, 55615, 111229, 333684}
		standardDeduction = 25900
	}

	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0. the dependent credit may cause it to be negative
	return int(tax), tax / grossIncome
}

func calcArkansasTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) * 0.5 // 50% of gains are taxed as ordinary income
	taxableIncome += (*dividends)          // dividends are taxed as ordinary income
	grossIncome := taxableIncome           // capture gross income now
	brackets := []int{0, 4300, 8500}
	rates := []float64{0.02, 0.04, 0.055}
	dependentExemption := 29
	standardDeduction := 2200
	personalExemption := 29
	if mfj {
		standardDeduction = 4400
		personalExemption = 58
	}

	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)
	tax -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0. the dependent credit may cause it to be negative
	return int(tax), tax / grossIncome
}

func calcCaliforniaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 9325, 22107, 34892, 48435, 61214, 312686, 375221, 625369, 1000000}
	rates := []float64{0.01, 0.02, 0.04, 0.06, 0.08, 0.093, 0.103, 0.113, 0.123, 0.133}
	dependentExemption := 400
	standardDeduction := 4803
	personalExemption := 129
	if mfj {
		brackets = []int{0, 18650, 44214, 69784, 96870, 122428, 625372, 750442, 1000000, 1250738}
		standardDeduction = 9606
		personalExemption = 258
	}

	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)
	tax -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0. the dependent credit may cause it to be negative
	return int(tax), tax / grossIncome
}

func calcColoradoTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.0455}
	dependentExemption := 400
	standardDeduction := 12950
	if mfj {
		standardDeduction = 25900
	}

	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0. the dependent credit may cause it to be negative
	return int(tax), tax / grossIncome
}

func calcConnecticutTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	tax += (*capitalGains) * 0.07                  // flat 7% tax on capital gains
	taxableIncome += (*dividends)                  // dividends are taxed as ordinary income
	grossIncome := taxableIncome + (*capitalGains) // add dividends since it wasn't added above
	brackets := []int{0, 10000, 50000, 100000, 200000, 250000, 500000}
	rates := []float64{0.03, 0.05, 0.055, 0.06, 0.065, 0.069, 0.0699}
	personalExemption := 15000
	if mfj {
		brackets = []int{0, 20000, 100000, 200000, 400000, 500000, 1000000}
		personalExemption = 24000
	}

	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcDelawareTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{2000, 5000, 10000, 20000, 25000, 60000}
	rates := []float64{0.022, 0.039, 0.048, 0.052, 0.0555, 0.066}
	dependentExemption := 110
	standardDeduction := 3250
	personalExemption := 110
	if mfj {
		standardDeduction = 6500
		personalExemption = 220
	}

	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)
	tax -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0. the dependent credit may cause it to be negative
	return int(tax), tax / grossIncome
}

// No income tax of any kind
func calcFloridaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	return 0, 0.0
}

func calcGeorgiaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 750, 2250, 3750, 5250, 7000}
	rates := []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.0575}
	dependentExemption := 3000
	standardDeduction := 5400
	personalExemption := 2700
	if mfj {
		brackets = []int{0, 1000, 3000, 5000, 7000, 10000}
		standardDeduction = 7100
		personalExemption = 7400
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcHawaiiTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	tax += (*capitalGains) * 0.0725                //  flat 7.25% tax on gains
	taxableIncome += (*dividends)                  // dividends are taxed as ordinary income
	grossIncome := taxableIncome + (*capitalGains) // capture gross income now
	brackets := []int{0, 2400, 4800, 9600, 14400, 19200, 24000, 36000, 48000, 150000, 175000, 200000}
	rates := []float64{0.014, 0.032, 0.055, 0.064, 0.068, 0.072, 0.076, 0.079, 0.0825, 0.09, 0.1, 0.11}
	dependentExemption := 1144
	standardDeduction := 2200
	personalExemption := 1144
	if mfj {
		brackets = []int{0, 4800, 9600, 19200, 28800, 38400, 48000, 72000, 96000, 300000, 350000, 400000}
		standardDeduction = 4400
		personalExemption = 2288
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcIdahoTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 1588, 4763, 7939}
	rates := []float64{0.01, 0.03, 0.045, 0.06}
	standardDeduction := 12950
	if mfj {
		brackets = []int{0, 3176, 9526, 15878}
		standardDeduction = 25900
	}

	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcIllinoisTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.0495}
	personalExemption := 2375
	if mfj {
		personalExemption = 4750
	}

	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcIndianaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) //  flat 7.25% tax on gains
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.0323}
	dependentExemption := 1000
	personalExemption := 1000
	if mfj {
		personalExemption = 2000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// allows federal tax deduction. Released for 2023?
func calcIowaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 1743, 3486, 6972, 15687, 26145, 34860, 52290, 78435}
	rates := []float64{0.0033, 0.0067, 0.0225, 0.0414, 0.0563, 0.0596, 0.0625, 0.0744, 0.0853}
	dependentExemption := 40
	standardDeduction := 2210
	personalExemption := 40
	if mfj {
		standardDeduction = 5450
		personalExemption = 80
	}

	tax -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	tax -= float64(personalExemption)
	taxableIncome -= float64(federalTax)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax)
	return int(tax), tax / grossIncome
}

func calcKansasTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 15000, 30000}
	rates := []float64{0.031, 0.0525, 0.057}
	dependentExemption := 2250
	standardDeduction := 3500
	personalExemption := 2250
	if mfj {
		brackets = []int{0, 30000, 60000}
		standardDeduction = 8000
		personalExemption = 4500
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcKentuckyTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.050}
	standardDeduction := 2770
	if mfj {
		brackets = []int{0, 30000, 60000}
		standardDeduction = 5540
	}

	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}
{
	name:                 "Louisiana",
	abbrev:               "LA",
	dependentExemption:   1000,
	dependentIsCredit:    false,
	stdDeductionIsCredit: false,
	exemptionIsCredit:    false,
	incomeTypesTaxed:     []float32{1.0, 1.0, 1.0},
	single: FilingStatus{
		brackets:          []int{0, 12500, 50000},
		rates:             []float64{0.0185, 0.035, 0.0425},
		standardDeduction: 0,
		personalExemption: 4500,
	},
	couple: FilingStatus{
		brackets:          []int{0, 25000, 100000},
		rates:             []float64{0.0185, 0.035, 0.0425},
		standardDeduction: 0,
		personalExemption: 9000,
	},
},

// repealed the federal tax deduction
func calcLouisianaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets :=  []int{0, 12500, 50000}
	rates := []float64{0.0185, 0.035, 0.0425}
	dependentExemption := 1000
	personalExemption := 4500
	if mfj {
		brackets = []int{0, 25000, 100000}
		personalExemption = 9000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}