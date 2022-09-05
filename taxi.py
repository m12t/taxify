"""
A Python script for manipulating the CSV data returned by `taxify.go`.

* This script accepts CLI arguments in the same format as the Go file.
  As such, it can decode them in the same way and access the exact CSV
  file that was created by the Go program, given the same inputs.

"""
import argparse
import pandas as pd
import matplotlib.pyplot as plt
from matplotlib.animation import FuncAnimation, FFMpegFileWriter


def main(args: list) -> None:
    df = pd.read_csv(get_filepath(args.income, args.steps))
    plot(df)


def plot(df: pd.DataFrame) -> None:
    for i in range(1, df.shape[0]):
        for j in range(1, df.shape[1]):
            plt.plot(df["income"].iloc[:i], df.iloc[:i, j])
        plt.xlim([0, df["income"].iloc[-1]])
        plt.ylim([0, 0.38])
        plt.title(df["income"].iloc[i])
        plt.xlabel("Income")
        plt.ylabel("Effective tax rate")
        # plt.savefig(f"./output/plots/demo/{i}.png")
    plt.show()


def get_filepath(income: int, steps: int) -> str:
    """generate the dynamic filepath givent the inputs"""
    return f"./output/csv/income={income}_steps={steps}.csv"


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-income",
        type=int,
        required=True,
        help="income (type <int>), digits only, used in the CSV",
    )
    parser.add_argument(
        "-steps",
        type=int,
        required=True,
        help="number of steps (type <int>) used in the CSV",
    )
    args = parser.parse_args()
    main(args)
