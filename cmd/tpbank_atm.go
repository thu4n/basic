package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var tpbankATMCmd = &cobra.Command{
	Use:   "tpbank-atm",
	Short: "Convert TPBank ATM/debit account statement",
	Long: `Convert a TPBank ATM or debit account Excel statement (.xlsx)
into a CSV file compatible with the Sure transaction import format.

The Excel file is expected to have data starting at row 9 on the "VN" sheet,
with columns in the standard TPBank ATM export format.`,
	RunE: runTPBankATM,
}

func init() {
	tpbankATMCmd.Flags().StringP("input", "i", "tpb_test_transactions.xlsx", "Input Excel file path")
	tpbankATMCmd.Flags().StringP("output", "o", "outputs/output.csv", "Output CSV file path")
	rootCmd.AddCommand(tpbankATMCmd)
}

func runTPBankATM(cmd *cobra.Command, args []string) error {
	inputFile, _ := cmd.Flags().GetString("input")
	outputFile, _ := cmd.Flags().GetString("output")

	fmt.Println("BASIC - Bank Account Statement Into CSV")
	fmt.Println("========================================")
	fmt.Printf("Format: TPBank ATM\n")
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

	// Get all the rows in the default VN sheet
	rows, err := f.GetRows("VN")
	if err != nil {
		return fmt.Errorf("reading sheet 'VN': %w", err)
	}

	if dir := filepath.Dir(outputFile); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
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

	// Data starts from row 9 in Excel file
	for i, row := range rows[8:] {
		// Skip header row (first row after row 8)
		if i == 0 {
			continue
		}

		if len(row) < 9 {
			fmt.Printf("Warning: Row %d has insufficient columns, skipping\n", i+9)
			continue
		}

		// Extract columns from TPBank ATM format:
		// 0: Ngày thực hiện (datetime)
		// 1: Ngày hiệu lực
		// 2: Mô tả giao dịch (description)
		// 3: Ghi nợ (debit)
		// 4: Ghi có (credit)
		// 5: Số dư (balance)
		// 6: Tài khoản đối ứng
		// 7: Tên tài khoản
		// 8: Mã giao dịch

		dateTimeStr := row[0]
		datePart := strings.Split(dateTimeStr, " ")[0]

		parsedDate, err := time.Parse("02-01-2006", datePart)
		var formattedDate string
		if err == nil {
			formattedDate = parsedDate.Format("02-01-2006") // DD-MM-YYYY
		} else {
			formattedDate = datePart // Keep original if parsing fails
		}

		var amount string
		debit := strings.TrimSpace(row[3])  // Ghi nợ
		credit := strings.TrimSpace(row[4]) // Ghi có

		if debit != "" {
			// Remove comma from Vietnamese number format: "52,000" -> "52000"
			debit = strings.ReplaceAll(debit, ",", "")
			amount = "-" + debit
		} else if credit != "" {
			credit = strings.ReplaceAll(credit, ",", "")
			amount = credit
		} else {
			amount = "0"
		}

		description := row[2] // Mô tả giao dịch
		correspondentAccount := row[6]
		accountName := row[7]
		notes := fmt.Sprintf("%s | TKDU: %s | TTK: %s", description, correspondentAccount, accountName)

		outputRow := []string{
			formattedDate, // date*
			amount,        // amount*
			description,   // name
			"VND",         // currency
			"",            // category (empty)
			"",            // tags (empty)
			"TP Bank ATM", // account
			notes,         // notes
		}

		if err := writer.Write(outputRow); err != nil {
			return fmt.Errorf("writing row %d: %w", i+9, err)
		}
	}

	fmt.Printf("\nSuccessfully wrote data to %s\n", outputFile)
	return nil
}
