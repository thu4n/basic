package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDNSEDate(t *testing.T) {
	tests := []struct {
		primary   string
		secondary string
		expected  string
	}{
		{"15/05/2024", "", "2024-05-15"},
		{"2024-05-16", "", "2024-05-16"},
		{"10:15:30 17/05/2024", "", "2024-05-17"},
		{"46213.502492835651", "", "2026-07-10"},
		{"07-10-26", "46213.502492835651", "2026-07-10"},
	}

	for _, tt := range tests {
		got := parseDNSEDate(tt.primary, tt.secondary)
		if got != tt.expected {
			t.Errorf("parseDNSEDate(%q, %q) = %q; want %q", tt.primary, tt.secondary, got, tt.expected)
		}
	}
}

func TestIsSellAction(t *testing.T) {
	if !isSellAction("BÁN") || !isSellAction("BAN") || !isSellAction("B") || !isSellAction("SELL") {
		t.Errorf("expected true for sell actions")
	}
	if isSellAction("MUA") || isSellAction("BUY") {
		t.Errorf("expected false for buy actions")
	}
}

func TestIsCancelledStatus(t *testing.T) {
	if !isCancelledStatus("Đã huỷ") || !isCancelledStatus("Đã hủy") || !isCancelledStatus("Cancelled") {
		t.Errorf("expected true for cancelled statuses")
	}
	if isCancelledStatus("Đã khớp") || isCancelledStatus("Matched") {
		t.Errorf("expected false for executed statuses")
	}
}

func TestDNSETradeConversionEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test_out.csv")

	// Test text file conversion
	cmd := dnseCmd
	cmd.Flags().Set("input", "../dnse_sample.txt")
	cmd.Flags().Set("output", outPath)

	err := runDNSE(cmd, nil)
	if err != nil {
		t.Fatalf("runDNSE failed: %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 5 { // 1 header + 4 rows
		t.Fatalf("expected 5 lines in output CSV, got %d:\n%s", len(lines), string(content))
	}

	expectedHeader := "date*,ticker*,exchange_operating_mic,currency,qty*,price*,account,name"
	if lines[0] != expectedHeader {
		t.Errorf("header = %q; want %q", lines[0], expectedHeader)
	}
}
