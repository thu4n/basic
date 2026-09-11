package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "basic",
	Short: "BASIC - Bank Account & Trade Statement Into CSV",
	Long: `BASIC converts bank account statements and stock trade statements
into CSV formats compatible with Sure transaction and trade import standard.

Supported formats:
  tpbank-atm    TPBank ATM/debit account statement
  tpbank-visa   TPBank Visa credit card statement
  dnse          DNSE stock trade statement`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
