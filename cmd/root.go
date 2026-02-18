package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "basic",
	Short: "BASIC - Bank Account Statement Into CSV",
	Long: `BASIC converts bank account statement Excel files into CSV format
compatible with the Sure transaction import format.

Supported formats:
  tpbank-atm    TPBank ATM/debit account statement
  tpbank-visa   TPBank Visa credit card statement`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
