package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var tpbankVisaCmd = &cobra.Command{
	Use:   "tpbank-visa",
	Short: "Convert TPBank Visa credit card statement",
	Long: `Convert a TPBank Visa credit card Excel statement (.xlsx) into a CSV file compatible with the Sure transaction import format.
The Excel file is expected to have data starting at row 9 on the "VN" sheet,
with columns in the standard TPBank ATM export format.`,
	RunE: runTPBankVisa,
}

func init() {
	tpbankVisaCmd.Flags().StringP("input", "i", "tpb_visa_transactions.xlsx", "Input Excel file path")
	tpbankVisaCmd.Flags().StringP("output", "o", "outputs/output.csv", "Output CSV file path")
	tpbankVisaCmd.Flags().StringP("sheet", "s", "VN", "Sheet name to read from")
	tpbankVisaCmd.Flags().IntP("data-row", "r", 9, "Row number where data starts (1-indexed)")
	rootCmd.AddCommand(tpbankVisaCmd)
}

func runTPBankVisa(cmd *cobra.Command, args []string) error {
	inputFile, _ := cmd.Flags().GetString("input")
	outputFile, _ := cmd.Flags().GetString("output")
	sheetName, _ := cmd.Flags().GetString("sheet")
	dataRow, _ := cmd.Flags().GetInt("data-row")

	fmt.Println("BASIC - Bank Account Statement Into CSV")
	fmt.Println("========================================")
	fmt.Printf("Format: TPBank Visa Credit Card\n")
	fmt.Printf("Input:  %s\n", inputFile)
	fmt.Printf("Output: %s\n", outputFile)

	f, err := excelize.OpenFile(inputFile)
	if err != nil {
		return fmt.Errorf("opening input file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("reading sheet '%s': %w", sheetName, err)
	}

	csvFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write header matching sure_transaction format
	header := []string{"date*", "amount*", "name", "currency", "category", "tags", "account", "notes"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	// Data starts from the specified row (convert to 0-indexed)
	startIdx := dataRow - 1
	if startIdx < 0 || startIdx >= len(rows) {
		return fmt.Errorf("data-row %d is out of range (file has %d rows)", dataRow, len(rows))
	}

	for i, row := range rows[startIdx:] {
		// Skip header row (first row of the data range)
		if i == 0 {
			continue
		}

		if len(row) == 0 {
			continue // skip empty rows
		}

		// TPBank Visa columns:
		// 0: Ngày giao dịch (transaction date)
		// 1: Ngày cập nhật (update date)
		// 2: Số tiền giao dịch (transaction amount)
		// 3: Ghi nợ (debit)
		// 4: Ghi có (credit)
		// 5: Mô tả giao dịch (description)

		if len(row) < 6 {
			fmt.Printf("Warning: Row %d has insufficient columns, skipping\n", startIdx+i+1)
			continue
		}

		dateStr := strings.TrimSpace(row[0])
		parsedDate, err := time.Parse("02/01/2006", dateStr)
		var formattedDate string
		if err == nil {
			formattedDate = parsedDate.Format("02-01-2006") // DD-MM-YYYY
		} else {
			formattedDate = dateStr // Keep original if parsing fails
		}

		debit := strings.TrimSpace(row[3])  // Ghi nợ
		credit := strings.TrimSpace(row[4]) // Ghi có

		var amount string
		if debit != "" {
			debit = strings.ReplaceAll(debit, ",", "")
			amount = "-" + debit // expense
		} else if credit != "" {
			credit = strings.ReplaceAll(credit, ",", "")
			amount = credit // income / refund
		} else {
			amount = "0"
		}

		description := strings.TrimSpace(row[5]) // Mô tả giao dịch

		outputRow := []string{
			formattedDate,  // date*
			amount,         // amount*
			description,    // name
			"VND",          // currency
			"",             // category (empty)
			"",             // tags (empty)
			"TP Bank Visa", // account
			description,    // notes
		}

		if err := writer.Write(outputRow); err != nil {
			return fmt.Errorf("writing row %d: %w", startIdx+i+1, err)
		}
	}

	fmt.Printf("\nSuccessfully wrote data to %s\n", outputFile)
	return nil
}
