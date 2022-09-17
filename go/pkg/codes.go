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

/* --------------------- federal tax code --------------------- */

func calcFederalTax(
	income, capitalGains, dividends float64,
	mfj, qualified bool) (int, float64) {
	medicareRate := 0.0145
	socialSecurityRate := 0.062
	socialSecurityCap := 147000
	ordinaryBrackets := []int{0, 10275, 41775, 89075, 170050, 215950, 539900}
	ordinaryRates := []float64{0.10, 0.12, 0.22, 0.24, 0.32, 0.35, 0.37}
	capitalGainsBrackets := []int{0, 41675, 459750}
	capitalGainsRates := []float64{0.0, 0.15, 0.20}
	standardDeduction := 12950
	if mfj {
		ordinaryBrackets = []int{0, 20550, 83550, 178150, 340100, 431900, 647850}
		capitalGainsBrackets = []int{0, 83350, 517200}
		standardDeduction = 25900
	}
	if qualified {
		capitalGains += dividends
	} else {
		income += dividends
	}
	tax := 0.0
	taxableIncome := income + capitalGains + dividends
	grossIncome := taxableIncome // freeze grossIncome now
	taxableIncome -= float64(standardDeduction)
	taxableIncome = math.Max(0.0, taxableIncome)
	tax += taxableIncome * medicareRate
	tax += taxEngine(&income, &ordinaryBrackets, &ordinaryRates)
	tax += taxEngine(&capitalGains, &capitalGainsBrackets, &capitalGainsRates)
	ssCappedIncome := math.Min(float64(socialSecurityCap), taxableIncome)
	tax += ssCappedIncome * socialSecurityRate
	return int(tax), tax / grossIncome
}

/* --------------------- state tax codes --------------------- */

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
		standardDeduction = 5540
	}

	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// repealed the federal tax deduction
func calcLouisianaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 12500, 50000}
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

func calcMaineTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 23000, 54450}
	rates := []float64{0.058, 0.0675, 0.0715}
	dependentExemption := 300
	standardDeduction := 12950
	personalExemption := 4450
	if mfj {
		brackets = []int{0, 46000, 108900}
		standardDeduction = 25900
		personalExemption = 8900
	}

	tax -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax)
	return int(tax), tax / grossIncome
}

func calcMarylandTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 1000, 2000, 3000, 100000, 125000, 150000, 250000}
	rates := []float64{0.02, 0.03, 0.04, 0.0475, 0.05, 0.0525, 0.055, 0.0575}
	dependentExemption := 3200
	standardDeduction := 2350
	personalExemption := 3200
	if mfj {
		brackets = []int{0, 1000, 2000, 3000, 150000, 175000, 225000, 300000}
		standardDeduction = 4700
		personalExemption = 6400
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcMassachusettsTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.05}
	dependentExemption := 1000
	personalExemption := 4400
	if mfj {
		personalExemption = 8800
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcMichiganTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.0425}
	dependentExemption := 5000
	personalExemption := 5000
	if mfj {
		personalExemption = 10000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcMinnesotaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 28080, 92230, 171220}
	rates := []float64{0.0535, 0.068, 0.0785, 0.0985}
	dependentExemption := 4450
	standardDeduction := 12900
	if mfj {
		brackets = []int{0, 41050, 163060, 284810}
		standardDeduction = 25800
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcMississippiTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{5000, 10000}
	rates := []float64{0.04, 0.05}
	dependentExemption := 1500
	standardDeduction := 2300
	personalExemption := 6000
	if mfj {
		standardDeduction = 4600
		personalExemption = 12000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// allows federal tax to be deducted up to $5000
func calcMissouriTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{108, 1088, 2176, 3264, 4352, 5440, 6528, 7616, 8704}
	rates := []float64{0.015, 0.02, 0.025, 0.03, 0.035, 0.04, 0.045, 0.05, 0.054}
	standardDeduction := 12950
	if mfj {
		standardDeduction = 25900
	}

	taxableIncome -= float64(standardDeduction)
	taxableIncome -= math.Min(5000, float64(federalTax)) // $5000 deduction cap

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// TODO: there's a 2% capital gains credit
func calcMontanaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 3100, 5500, 8400, 11400, 14600, 18800}
	rates := []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.0675}
	dependentExemption := 2580
	standardDeduction := 4830
	personalExemption := 2580
	fedDeductibilityCap := 5000
	if mfj {
		standardDeduction = 9660
		personalExemption = 5160
		fedDeductibilityCap = 10000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)
	taxableIncome -= math.Min(float64(fedDeductibilityCap), float64(federalTax)) // 5 or 10k max deduction

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcNebraskaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 3440, 20590, 33180}
	rates := []float64{0.0246, 0.0351, 0.0501, 0.0684}
	dependentExemption := 146
	standardDeduction := 7350
	personalExemption := 146
	if mfj {
		brackets = []int{0, 6860, 41190, 66360}
		standardDeduction = 14700
		personalExemption = 292
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
func calcNevadaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	return 0, 0.0
}

// 5% flat tax on dividend income
func calcNewHampshireTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	grossIncome := (*income) + (*capitalGains) + (*dividends)
	taxableIncome := (*dividends)
	personalExemption := 2400
	if mfj {
		personalExemption = 4800
	}
	taxableIncome -= float64(personalExemption)
	taxableIncome = math.Max(0, taxableIncome) // assert >= 0
	tax := taxableIncome * 0.05
	return int(tax), tax / grossIncome
}

func calcNewJerseyTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 20000, 35000, 40000, 75000, 500000, 1000000}
	rates := []float64{0.014, 0.0175, 0.035, 0.05525, 0.0637, 0.0897, 0.1075}
	dependentExemption := 1500
	personalExemption := 1000
	if mfj {
		brackets = []int{0, 20000, 50000, 70000, 80000, 150000, 500000, 1000000}
		rates = []float64{0.014, 0.0175, 0.0245, 0.035, 0.05525, 0.0637, 0.0897, 0.1075}
		personalExemption = 2000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// allows a deduction for 40% of capital gains
func calcNewMexicoTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 5500, 11000, 16000, 210000}
	rates := []float64{0.017, 0.032, 0.047, 0.049, 0.059}
	dependentExemption := 4000
	standardDeduction := 12950
	if mfj {
		brackets = []int{0, 8000, 16000, 24000, 315000}
		standardDeduction = 25900
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= math.Max(1000, (*capitalGains)*0.4) // greater of 1000, 40% of gains

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcNewYorkTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 8500, 11700, 13900, 80650, 215400, 1077550, 5000000, 25000000}
	rates := []float64{0.04, 0.045, 0.0525, 0.0585, 0.0625, 0.0685, 0.0965, 0.103, 0.109}
	dependentExemption := 1000
	standardDeduction := 8000
	if mfj {
		brackets = []int{0, 17150, 23600, 27900, 161550, 323200, 2155350, 5000000, 25000000}
		standardDeduction = 16050
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcNorthCarolinaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.0499}
	standardDeduction := 12750
	if mfj {
		standardDeduction = 25500
	}

	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// allows a deduction for 40% of capital gains
func calcNorthDakotaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) * 0.6 // 40% deduction, taxed as ordinary income
	taxableIncome += (*dividends)          // dividends are taxed as ordinary income
	grossIncome := taxableIncome           // capture gross income now
	brackets := []int{0, 40525, 98100, 204675, 445000}
	rates := []float64{0.011, 0.0204, 0.0227, 0.0264, 0.029}
	standardDeduction := 12950
	if mfj {
		brackets = []int{0, 67700, 163550, 249150, 445000}
		standardDeduction = 25900
	}

	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcOhioTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{25000, 44250, 88450, 110650}
	rates := []float64{0.02765, 0.03226, 0.03688, 0.0399}
	dependentExemption := 2400
	personalExemption := 2400
	if mfj {
		personalExemption = 4800
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcOklahomaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 1000, 2500, 3750, 4900, 7200}
	rates := []float64{0.0025, 0.0075, 0.0175, 0.0275, 0.0375, 0.0475}
	dependentExemption := 1000
	standardDeduction := 6350
	personalExemption := 1000
	if mfj {
		brackets = []int{0, 2000, 5000, 7500, 9800, 12200}
		standardDeduction = 12700
		personalExemption = 2000
	}

	taxableIncome -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// federal tax deduction of up to 6950
func calcOregonTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 3650, 9200, 125000}
	rates := []float64{0.0475, 0.0675, 0.0875, 0.099}
	dependentExemption := 219
	standardDeduction := 2420
	personalExemption := 219
	if mfj {
		brackets = []int{0, 7300, 18400, 250000}
		standardDeduction = 4840
		personalExemption = 436
	}

	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	taxableIncome -= float64(standardDeduction)
	tax -= float64(personalExemption)
	taxableIncome -= math.Min(6950, float64(federalTax)) // $5000 deduction cap

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0
	return int(tax), tax / grossIncome
}

func calcPennsylvaniaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.0307}

	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcRhodeIslandTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 68200, 155050}
	rates := []float64{0.0375, 0.0475, 0.0599}
	dependentExemption := 4350
	standardDeduction := 9300
	personalExemption := 4350
	if mfj {
		standardDeduction = 18600
		personalExemption = 8700
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// allows a deduction for 44% of capital gains
func calcSouthCarolinaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) * 0.56 // 44% deduction, taxed as ordinary income
	taxableIncome += (*dividends)           // dividends are taxed as ordinary income
	grossIncome := taxableIncome            // capture gross income now
	brackets := []int{0, 3200, 6410, 9620, 12820, 16040}
	rates := []float64{0.0, 0.03, 0.04, 0.05, 0.06, 0.07}
	dependentExemption := 4300
	standardDeduction := 12950
	if mfj {
		standardDeduction = 25900
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// No income tax of any kind
func calcSouthDakotaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	return 0, 0.0
}

// No income tax of any kind
func calcTennesseeTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	return 0, 0.0
}

// No income tax of any kind
func calcTexasTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	return 0, 0.0
}

func calcUtahTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0}
	rates := []float64{0.0495}
	dependentExemption := 1750
	standardDeduction := 777
	if mfj {
		standardDeduction = 1554
	}

	tax -= float64(numDependents * dependentExemption) // is a credit, not a deduction
	tax -= float64(standardDeduction)

	tax += taxEngine(&taxableIncome, &brackets, &rates)
	tax = math.Max(0, tax) // assert tax >= 0. the dependent credit may cause it to be negative
	return int(tax), tax / grossIncome
}

func calcVermontTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 40950, 99200, 206950}
	rates := []float64{0.0335, 0.066, 0.076, 0.0875}
	dependentExemption := 4350
	standardDeduction := 6350
	personalExemption := 4350
	if mfj {
		brackets = []int{0, 68400, 165350, 251950}
		standardDeduction = 12700
		personalExemption = 8700
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcVirginiaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 3000, 5000, 17000}
	rates := []float64{0.02, 0.03, 0.05, 0.0575}
	dependentExemption := 930
	standardDeduction := 4500
	personalExemption := 930
	if mfj {
		standardDeduction = 9000
		personalExemption = 1860
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// 7% flat tax on capital gains
func calcWashingtonTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	grossIncome := (*income) + (*capitalGains) + (*dividends)
	taxableIncome := (*capitalGains)
	standardDeduction := 250000
	taxableIncome -= float64(standardDeduction)
	taxableIncome = math.Max(0, taxableIncome) // assert >= 0
	tax := taxableIncome * 0.07
	return int(tax), tax / grossIncome
}

func calcWestVirginiaTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 10000, 25000, 40000, 60000}
	rates := []float64{0.03, 0.04, 0.045, 0.06, 0.065}
	dependentExemption := 2000
	personalExemption := 2000
	if mfj {
		personalExemption = 4000
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

func calcWisconsinTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 12760, 25520, 280950}
	rates := []float64{0.0354, 0.0465, 0.053, 0.0765}
	dependentExemption := 700
	standardDeduction := 11790
	personalExemption := 700
	if mfj {
		brackets = []int{0, 17010, 34030, 374030}
		standardDeduction = 21820
		personalExemption = 1400
	}

	taxableIncome -= float64(numDependents * dependentExemption)
	taxableIncome -= float64(standardDeduction)
	taxableIncome -= float64(personalExemption)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}

// No income tax of any kind
func calcWyomingTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	return 0, 0.0
}

func calcDCTax(income, capitalGains, dividends *float64,
	federalTax, numDependents int, mfj bool) (int, float64) {
	tax, taxableIncome := 0.0, (*income)
	taxableIncome += (*capitalGains) // gains are taxed as ordinary income
	taxableIncome += (*dividends)    // dividends are taxed as ordinary income
	grossIncome := taxableIncome     // capture gross income now
	brackets := []int{0, 10000, 40000, 60000, 250000, 500000, 1000000}
	rates := []float64{0.04, 0.06, 0.065, 0.085, 0.0925, 0.0975, 0.1075}
	standardDeduction := 12950
	if mfj {
		standardDeduction = 25900
	}

	taxableIncome -= float64(standardDeduction)

	taxableIncome = math.Max(0, taxableIncome) // assert taxableIncome >= 0
	tax += taxEngine(&taxableIncome, &brackets, &rates)
	return int(tax), tax / grossIncome
}