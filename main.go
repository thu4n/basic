package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("tpb_test_transactions.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Get all the rows in the default VN sheet
	rows, err := f.GetRows("VN")
	if err != nil {
		fmt.Println(err)
		return
	}

	csvFile, err := os.Create("output.csv")
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write header matching sure_transaction format
	header := []string{"date*", "amount*", "name", "currency", "category", "tags", "account", "notes"}
	if err := writer.Write(header); err != nil {
		fmt.Println("Error writing header:", err)
		return
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

		// Extract columns from TPBank format:
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

		outputRow := []string{
			formattedDate, // date*
			amount,        // amount*
			"",            // name (empty)
			"VND",         // currency
			"",            // category (empty)
			"",            // tags (empty)
			"TP Bank ATM", // account
			description,   // notes
		}

		if err := writer.Write(outputRow); err != nil {
			fmt.Println("Error writing row:", err)
			return
		}
	}

	fmt.Println("Successfully wrote data to output.csv")
}
