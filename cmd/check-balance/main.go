package main

import (
	"context"
	"fmt"
	"os"

	"github.com/4chain-AG/spv-wallet-regression/internal/utils"
	"github.com/4chain-AG/spv-wallet-regression/internal/wallet"
)

const (
	minimalBalance = 20
)

func main() {
	ctx := context.Background()

	instanceURL, err := utils.GetEnv(wallet.MASTER_INSTANCE_URL)
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Error: %s environment variable is not set: %v\n", wallet.MASTER_INSTANCE_URL, err)
		os.Exit(1)
	}
	xpriv, err := utils.GetEnv(wallet.MASTER_INSTANCE_XPRIV)
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Error: %s environment variable is not set: %v\n", wallet.MASTER_INSTANCE_XPRIV, err)
		os.Exit(1)
	}

	instanceURL = utils.AddPrefixIfNeeded(instanceURL)

	balance, err := wallet.GetBalance(ctx, instanceURL, xpriv)
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Error: Failed to check balance: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(utils.StdOut, "Current balance: %d satoshis\n", balance)

	if balance < minimalBalance {
		fmt.Fprintf(utils.StdOut, "Insufficient funds! Required: %d, Available: %d\n", minimalBalance, balance)
		os.Exit(1)
	}

	fmt.Fprintln(utils.StdOut, "Balance check passed! Sufficient funds available.")
}
