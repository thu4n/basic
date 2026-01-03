# BASIC - Bank Account Statement Into CSV

A simple Go utility to convert TPBank account statements (Excel format) into a standardized CSV format compatible with **Sure** - a personal finance tracking application.

## What it does

- Reads TPBank account statement data from Excel files (`.xlsx`)
- Parses transaction dates, amounts, and descriptions
- Converts debit/credit columns into signed amounts (negative for debits, positive for credits)
- Outputs a clean CSV file ready for import into Sure

## Installation

Install the tool to your Go bin directory so you can run `basic` from anywhere:

```bash
go install
```

This installs the binary to `~/go/bin/basic`. Now you can use it globally:

```bash
basic -input myfile.xlsx -output result.csv
```

## Usage


### Get Help

```bash
basic -h
```

### Custom Input/Output Files

```bash
basic -input mystatement.xlsx -output transactions.csv
```

### Running from Source

If you haven't installed it yet, you can run directly:

```bash
go run main.go

# Or with custom files
go run main.go -input mystatement.xlsx -output transactions.csv
```

### CLI Flags

- `-input`: Path to the input Excel file (default: `tpb_test_transactions.xlsx`)
- `-output`: Path to the output CSV file (default: `output.csv`)
- `-h`: Show help

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
```
