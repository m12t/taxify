import argparse


def main(args: list) -> None:
    with open(get_filepath(args.income, args.steps), "r") as f:
        print(f.read())


def get_filepath(income: int, steps: int) -> str:
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
