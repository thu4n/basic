package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var dnseCmd = &cobra.Command{
	Use:     "dnse",
	Aliases: []string{"dnse-trade"},
	Short:   "Convert DNSE stock trade statement",
	Long: `Convert a DNSE stock trade Excel (.xlsx) or TSV/CSV statement into a CSV file
compatible with the Sure trade import format (trade_sample.csv).

The tool automatically detects date, ticker, order type (Buy/Sell), quantity,
and price columns from the input statement, skipping cancelled or unexecuted orders.`,
	RunE: runDNSE,
}

func init() {
	dnseCmd.Flags().StringP("input", "i", "dnse_sample.txt", "Input Excel (.xlsx) or text/CSV file path")
	dnseCmd.Flags().StringP("output", "o", "outputs/trade_output.csv", "Output CSV file path")
	dnseCmd.Flags().StringP("sheet", "s", "Sheet1", "Sheet name to read from (for Excel files)")
	dnseCmd.Flags().IntP("data-row", "r", 3, "Row number where data starts (1-indexed, auto-detected if set to 3)")
	dnseCmd.Flags().StringP("account", "a", "DNSE Trading Account", "Account name for Sure trade import")
	dnseCmd.Flags().StringP("currency", "c", "VND", "Currency code")
	dnseCmd.Flags().StringP("mic", "m", "XSTC", "Exchange operating MIC (e.g., XSTC for HOSE, HSTC for HNX)")
	rootCmd.AddCommand(dnseCmd)
}

func runDNSE(cmd *cobra.Command, args []string) error {
	inputFile, _ := cmd.Flags().GetString("input")
	outputFile, _ := cmd.Flags().GetString("output")
	sheetName, _ := cmd.Flags().GetString("sheet")
	dataRow, _ := cmd.Flags().GetInt("data-row")
	accountName, _ := cmd.Flags().GetString("account")
	currency, _ := cmd.Flags().GetString("currency")
	mic, _ := cmd.Flags().GetString("mic")

	fmt.Println("BASIC - Bank Account & Trade Statement Into CSV")
	fmt.Println("===============================================")
	fmt.Printf("Format: DNSE Stock Trade Import\n")
	fmt.Printf("Input:  %s\n", inputFile)
	fmt.Printf("Output: %s\n", outputFile)

	rows, err := readRowsFromFile(inputFile, sheetName)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("input file is empty")
	}

	// Detect column locations
	colMap := detectDNSEColumns(rows)

	if dir := filepath.Dir(outputFile); dir != "." && dir != "" {
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

	// Write header matching Sure trade import format (trade_sample.csv)
	header := []string{"date*", "ticker*", "exchange_operating_mic", "currency", "qty*", "price*", "account", "name"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	processedCount := 0
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}

		// Skip header rows or metadata rows unless auto-detected as a valid data row
		if i < dataRow-1 && !isDataRow(row, colMap) {
			continue
		}

		// Check if row is a valid data row
		if !isDataRow(row, colMap) {
			continue
		}

		// Extract & check status
		statusStr := getRowField(row, colMap.Status, -1)
		if isCancelledStatus(statusStr) {
			// Skip cancelled orders
			continue
		}

		// Extract Matched Quantity (KL khớp)
		qtyStr := getRowField(row, colMap.Qty, -1)
		qtyVal := parseNumber(qtyStr)

		// Check if matched quantity is 0 (unexecuted or cancelled order)
		if qtyVal <= 0 {
			if !isExecutedStatus(statusStr) {
				continue // Skip non-executed orders with 0 quantity
			}
			qtyVal = parseNumber(getRowField(row, colMap.OrderQty, -1))
			if qtyVal <= 0 {
				continue
			}
		}

		// Extract Action (Mua / Bán)
		actionStr := getRowField(row, colMap.Action, -1)
		if actionStr == "" {
			for _, cell := range row {
				if isActionString(cell) {
					actionStr = cell
					break
				}
			}
		}
		isSell := isSellAction(actionStr)
		if isSell {
			qtyVal = -qtyVal
		}

		// Extract Date
		rawDate := getRowField(row, colMap.Date, -1)
		timeStr := getRowField(row, colMap.Time, -1)

		if rawDate == "" && timeStr == "" {
			for c := 0; c < len(row) && c < 5; c++ {
				if isDateString(row[c]) {
					rawDate = row[c]
					break
				}
			}
		}
		if rawDate == "" && timeStr == "" {
			continue
		}
		formattedDate := parseDNSEDate(timeStr, rawDate)

		// Extract Ticker
		ticker := strings.ToUpper(getRowField(row, colMap.Ticker, -1))
		if ticker == "" {
			for c := 0; c < len(row) && c < 6; c++ {
				cand := strings.TrimSpace(row[c])
				if isTickerString(cand) {
					ticker = strings.ToUpper(cand) + ".VN"
					break
				}
			}
		}
		if ticker == "" {
			continue
		}

		// Extract Matched Price (Giá khớp)
		priceStr := getRowField(row, colMap.Price, colMap.OrderPrice)
		priceVal := parseNumber(priceStr)

		// Format output values
		qtyFormatted := strconv.FormatFloat(qtyVal, 'f', -1, 64)
		priceFormatted := strconv.FormatFloat(priceVal, 'f', -1, 64)

		tradeName := fmt.Sprintf("%s Purchase", ticker)
		if isSell {
			tradeName = fmt.Sprintf("%s Sale", ticker)
		}

		acc := accountName

		outputRow := []string{
			formattedDate,  // date*
			ticker,         // ticker*
			mic,            // exchange_operating_mic
			currency,       // currency
			qtyFormatted,   // qty*
			priceFormatted, // price*
			acc,            // account
			tradeName,      // name
		}

		if err := writer.Write(outputRow); err != nil {
			return fmt.Errorf("writing row %d: %w", i+1, err)
		}
		processedCount++
	}

	fmt.Printf("\nSuccessfully converted %d trades to %s\n", processedCount, outputFile)
	return nil
}

type dnseColMap struct {
	Date       int
	Time       int
	Action     int
	Ticker     int
	SubAccount int
	Qty        int
	OrderQty   int
	Price      int
	OrderPrice int
	Status     int
}

func detectDNSEColumns(rows [][]string) dnseColMap {
	m := dnseColMap{
		Time:       1,
		Date:       2,
		Action:     3,
		Ticker:     4,
		SubAccount: 5,
		OrderQty:   7,
		Qty:        8,
		OrderPrice: 9,
		Price:      10,
		Status:     16,
	}

	// Search up to row 20 for column headers
	limit := len(rows)
	if limit > 20 {
		limit = 20
	}

	for r := 0; r < limit; r++ {
		for c, val := range rows[r] {
			cleanVal := strings.ToLower(strings.TrimSpace(val))
			switch {
			case strings.Contains(cleanVal, "ngày giao dịch") || strings.Contains(cleanVal, "ngày gd"):
				m.Date = c
			case strings.Contains(cleanVal, "thời gian đặt"):
				m.Time = c
			case cleanVal == "lệnh" || strings.Contains(cleanVal, "loại lệnh") || cleanVal == "chiều":
				m.Action = c
			case cleanVal == "mã" || cleanVal == "mã ck" || cleanVal == "mã chứng khoán" || cleanVal == "ticker":
				m.Ticker = c
			case strings.Contains(cleanVal, "tiểu khoản"):
				m.SubAccount = c
			case cleanVal == "kl khớp" || cleanVal == "khối lượng khớp":
				m.Qty = c
			case cleanVal == "kl đặt" || cleanVal == "khối lượng đặt":
				m.OrderQty = c
			case cleanVal == "giá khớp":
				m.Price = c
			case cleanVal == "giá đặt":
				m.OrderPrice = c
			case strings.Contains(cleanVal, "trạng thái"):
				m.Status = c
			}
		}
	}

	return m
}

func isDataRow(row []string, m dnseColMap) bool {
	if len(row) < 4 {
		return false
	}
	action := getRowField(row, m.Action, -1)
	ticker := getRowField(row, m.Ticker, -1)
	if isActionString(action) && isTickerString(ticker) {
		return true
	}
	// Fallback scan across cells
	hasAction := false
	hasTicker := false
	for _, cell := range row {
		if isActionString(cell) {
			hasAction = true
		}
		if isTickerString(cell) {
			hasTicker = true
		}
	}
	return hasAction && hasTicker
}

func getRowField(row []string, primaryIdx int, fallbackIdx int) string {
	if primaryIdx >= 0 && primaryIdx < len(row) {
		val := strings.TrimSpace(row[primaryIdx])
		if val != "" {
			return val
		}
	}
	if fallbackIdx >= 0 && fallbackIdx < len(row) {
		val := strings.TrimSpace(row[fallbackIdx])
		if val != "" {
			return val
		}
	}
	return ""
}

func readRowsFromFile(filePath string, sheetName string) ([][]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".xlsx" || ext == ".xls" {
		f, err := excelize.OpenFile(filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		targetSheet := sheetName
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets found in Excel file")
		}

		sheetExists := false
		for _, s := range sheets {
			if s == targetSheet {
				sheetExists = true
				break
			}
		}
		if !sheetExists {
			targetSheet = sheets[0] // fallback to first sheet
		}

		return f.GetRows(targetSheet)
	}

	// Fallback to text/CSV/TSV reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rows [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var cols []string
		if strings.Contains(line, "\t") {
			cols = strings.Split(line, "\t")
		} else {
			r := csv.NewReader(strings.NewReader(line))
			r.LazyQuotes = true
			parsed, err := r.Read()
			if err == nil {
				cols = parsed
			} else {
				cols = strings.Split(line, ",")
			}
		}
		rows = append(rows, cols)
	}
	return rows, scanner.Err()
}

func parseDNSEDate(primary string, secondary string) string {
	// First check if any candidate is an Excel serial number
	for _, raw := range []string{primary, secondary} {
		raw = strings.Trim(strings.TrimSpace(raw), ",")
		if raw == "" {
			continue
		}
		if floatVal, err := strconv.ParseFloat(raw, 64); err == nil && floatVal > 30000 && floatVal < 100000 {
			if t, err := excelize.ExcelDateToTime(floatVal, false); err == nil {
				return t.Format("2006-01-02")
			}
		}
	}

	for _, raw := range []string{primary, secondary} {
		raw = strings.Trim(strings.TrimSpace(raw), ",")
		if raw == "" {
			continue
		}

		// Extract date part if string contains time
		parts := strings.Fields(raw)
		datePart := raw
		for _, p := range parts {
			pClean := strings.Trim(p, ",")
			if strings.Contains(pClean, "/") || strings.Contains(pClean, "-") {
				datePart = pClean
				break
			}
		}

		layouts := []string{
			"02/01/2006",
			"2006-01-02",
			"02-01-2006",
			"2006/01/02",
			"1/2/2006",
			"02/01/06",
			"02-01-06",
			"06-01-02",
			"06/01/02",
			"02/01/2006 15:04:05",
			"2006-01-02 15:04:05",
		}

		for _, layout := range layouts {
			if t, err := time.Parse(layout, datePart); err == nil {
				if t.Year() < 100 {
					t = t.AddDate(2000, 0, 0)
				}
				return t.Format("2006-01-02")
			}
		}
	}

	res := strings.Trim(strings.TrimSpace(primary), ",")
	if res == "" {
		res = strings.Trim(strings.TrimSpace(secondary), ",")
	}
	return res
}

var dateRegex = regexp.MustCompile(`\d{1,4}[/-]\d{1,2}[/-]\d{1,4}`)
var tickerRegex = regexp.MustCompile(`^[A-Z]{3,5}$`)

func isDateString(s string) bool {
	s = strings.TrimSpace(s)
	if floatVal, err := strconv.ParseFloat(s, 64); err == nil && floatVal > 30000 && floatVal < 100000 {
		return true
	}
	return dateRegex.MatchString(s)
}

func isTickerString(s string) bool {
	s = strings.TrimSpace(s)
	return tickerRegex.MatchString(s) && s != "BUY" && s != "SELL" && s != "MUA" && s != "BAN" && s != "VND" && s != "USD"
}

func isActionString(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	return s == "MUA" || s == "BÁN" || s == "BAN" || s == "B" || s == "M" || s == "BUY" || s == "SELL"
}

func isSellAction(action string) bool {
	action = strings.ToUpper(strings.TrimSpace(action))
	return action == "BÁN" || action == "BAN" || action == "B" || action == "SELL" || action == "SALE" || action == "SELLING"
}

func isCancelledStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return false
	}
	return strings.Contains(status, "hủy") || strings.Contains(status, "huỷ") || strings.Contains(status, "cancel") || strings.Contains(status, "thất bại") || strings.Contains(status, "reject")
}

func isExecutedStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	return strings.Contains(status, "khớp") || strings.Contains(status, "done") || strings.Contains(status, "fill") || strings.Contains(status, "matched")
}

func parseNumber(str string) float64 {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, ",", "")
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return val
}
