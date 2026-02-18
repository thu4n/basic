# BASIC - Bank Account Statement Into CSV

A simple Go utility to convert TPBank account statements (Excel format) into a standardized CSV format compatible with **Sure** - a personal finance tracking application.

Supported formats:
- `tpbank-atm` — TPBank ATM/debit account statement
- `tpbank-visa` — TPBank Visa credit card statement

## What it does

- Reads TPBank account statement data from Excel files (`.xlsx`)
- Parses transaction dates, amounts, and descriptions
- Converts debit/credit columns into signed amounts (negative for debits, positive for credits)
- Outputs a clean CSV file ready for import into Sure

## Installation

Build the binary:

```bash
go build -o basic .
```

Or install it to your Go bin directory so you can run `basic` from anywhere:

```bash
go install
```

This installs the binary to `~/go/bin/basic`. Now you can use it globally:

```bash
basic tpbank-atm -i myfile.xlsx -o result.csv
```

## Usage


### Get Help

```bash
# List all subcommands
basic --help

# Help for a specific subcommand
basic tpbank-atm --help
basic tpbank-visa --help
```

### TPBank ATM / Debit

```bash
basic tpbank-atm -i mystatement.xlsx -o transactions.csv
```

### TPBank Visa Credit Card

```bash
basic tpbank-visa -i visa_statement.xlsx -o transactions.csv

# If the sheet name or data start row differs from the defaults:
basic tpbank-visa -i visa_statement.xlsx -o transactions.csv -s "Sheet1" -r 2
```

### Running from Source

If you haven't installed it yet, you can run directly:

```bash
go run . tpbank-atm -i mystatement.xlsx -o transactions.csv
```

### CLI Flags

All subcommands share these flags:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--input` | `-i` | Path to the input Excel file | subcommand-specific |
| `--output` | `-o` | Path to the output CSV file | `output.csv` |
| `--help` | `-h` | Show help | |

`tpbank-visa` also supports:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--sheet` | `-s` | Sheet name to read from | `Sheet1` |
| `--data-row` | `-r` | Row number where data starts (1-indexed) | `2` |

The program will:
1. Read `tpb_test_transactions.xlsx
2. Process transactions starting from row 9 (skipping headers)
3. Generate `output.csv` in the current directory

## Output Format

The generated CSV follows this schema:

```
date*,amount*,name,currency,category,tags,account,notes
```

- **date**: DD-MM-YYYY format
- **amount**: Negative for expenses (Ghi nợ), positive for income (Ghi có)
- **currency**: VND
- **account**: TP Bank ATM
- **notes**: Original transaction description

## Requirements

- Go 1.25.5 or higher
- TPBank account statement export file in Excel format

## Dependencies

```bash
go get github.com/xuri/excelize/v2
go get github.com/spf13/cobra
```
