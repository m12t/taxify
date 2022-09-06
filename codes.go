// A store for all tax codes encoded in functions
package codes

import (
	"math"
)

// Try to do 5 states per day.

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
	taxableIncome += *capitalGains // gains are taxed as ordinary income
	taxableIncome += *dividends    // gains are taxed as ordinary income
	grossIncome := taxableIncome   // capture gross income now
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

	// deductions
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
	taxableIncome += *capitalGains // gains are taxed as ordinary income
	taxableIncome += *dividends    // gains are taxed as ordinary income
	grossIncome := taxableIncome   // capture gross income now
	brackets := []int{0, 27808, 55615, 116843}
	rates := []float64{0.0259, 0.0334, 0.0417, 0.045}
	dependentExemption := 100
	standardDeduction := 12950
	if mfj {
		brackets = []int{0, 55615, 111229, 333684}
		standardDeduction = 25900
	}

	// deductions
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
	taxableIncome += *dividends            // gains are taxed as ordinary income
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

	// deductions
	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)
	tax -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0. the dependent credit may cause it to be negative
	return int(tax), tax / grossIncome
}
