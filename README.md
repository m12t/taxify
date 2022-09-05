# a 50-State-at-once tax calculator
* This is a simple CLI tool for calculating state income tax in all 50 states at once for a given taxable income.
* To run the program, either build it beforehand and call the executable, or simply run: `go run taxify.go -income=xxxxxx`
* Results are returned in descending order by default, though this can be reversed by invoking the flag `-ascending=true`
* In addition to the report that will automatically print to the terminal, you can specify other command line arguments to shape the output:
    - `-plot=true` will run the plot with default values
    - `-top=x` (default value is `7`) will plot only the top x number of states. `top` here depends on the value for the flag `-ascending`, which defaults to `false`.
    - `-ascending=true` will rank the states from lowest to highest total income tax for the given income.
    - `-numSteps=x` is less important, and specifies the number of discrete calculations to be made between $0 and `income` to use when plotting. A higher value will lead to a smoother and more accurate plot, but there's diminishing returns. The default is 100, which works quite well.
An example plot:
![Plot of effective tax from $0 to $1M in ordinary income](https://github.com/m12t/taxify/blob/main/plots/plot.png)

Example output to the terminal:
![Terminal output](https://github.com/m12t/taxify/blob/main/plots/example_output.png)