package main

import (
	"context"

	"github.com/4chain-AG/spv-wallet-regression/internal/utils"
	"github.com/4chain-AG/spv-wallet-regression/internal/wallet"
)

func main() {
	ctx := context.Background()

	instanceURL, err := utils.GetEnv(wallet.MASTER_INSTANCE_URL)
	if err != nil {
		utils.HandleErrorAndExit("Error: %s environment variable is not set: %v\n", wallet.MASTER_INSTANCE_URL, err)

	}
	xpriv, err := utils.GetEnv(wallet.MASTER_INSTANCE_XPRIV)
	if err != nil {
		utils.HandleErrorAndExit("Error: %s environment variable is not set: %v\n", wallet.MASTER_INSTANCE_XPRIV, err)
	}

	instanceURL = utils.AddPrefixIfNeeded(instanceURL)

	balance, err := wallet.GetBalance(ctx, instanceURL, xpriv)
	if err != nil {
		utils.HandleErrorAndExit("Error: Failed to check balance: %v\n", err)
	}

	utils.PrintOutput("Current balance: %d satoshis\n", balance)

	if balance < 2*wallet.MinimalBalance {
		utils.HandleErrorAndExit("Insufficient funds! Required: %d, Available: %d\n", wallet.MinimalBalance, balance)
	}

	utils.PrintOutput("Balance check passed! Sufficient funds available.")
}
