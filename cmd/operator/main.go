package main

import (
	"context"

	"github.com/4chain-AG/spv-wallet-regression/internal/utils"
	"github.com/4chain-AG/spv-wallet-regression/internal/wallet"
)

const (
	adminXPriv         = "xprv9s21ZrQH143K3CbJXirfrtpLvhT3Vgusdo8coBritQ3rcS7Jy7sxWhatuxG5h2y1Cqj8FKmPp69536gmjYRpfga2MJdsGyBsnB12E19CESK"
	adminXPub          = "xpub661MyMwAqRbcFgfmdkPgE2m5UjHXu9dj124DbaGLSjaqVESTWfCD4VuNmEbVPkbYLCkykwVZvmA8Pbf8884TQr1FgdG2nPoHR8aB36YdDQh"
	leaderPaymailAlias = "leader"
)

func main() {
	ctx := context.Background()

	config, err := wallet.LoadConfig()
	if err != nil {
		utils.HandleErrorAndExit("Failed to load configuration: %v\n", err)
	}

	leaderOne, err := wallet.CreateUser(ctx, config.ClientOneURL, config.ClientOneLeaderXPriv, adminXPriv, adminXPub, leaderPaymailAlias)
	if err != nil {
		utils.HandleErrorAndExit("Failed to create leader user for %v, error: %v\n", config.ClientOneURL, err)
	}

	leaderTwo, err := wallet.CreateUser(ctx, config.ClientTwoURL, config.ClientTwoLeaderXPriv, adminXPriv, adminXPub, leaderPaymailAlias)
	if err != nil {
		utils.HandleErrorAndExit("Failed to create leader user for %v, error: %v\n", config.ClientTwoURL, err)
	}

	if _, err := wallet.SendFunds(ctx, config.MasterURL, config.MasterXPriv, leaderOne.Paymail, 10); err != nil {
		utils.HandleErrorAndExit("Failed to send funds to %v: %v\n", leaderOne.Paymail, err)
	}

	leaderOneBalance, err := wallet.GetBalance(ctx, config.ClientOneURL, leaderOne.XPriv)
	if err != nil {
		utils.HandleErrorAndExit("Failed to get balance for master instance, error: %v\n", err)
	}

	if leaderOneBalance < wallet.MinimalBalance {
		utils.HandleErrorAndExit("Leader instance %v has insufficient funds: %d\n", config.ClientOneURL, leaderOneBalance)
	}

	if _, err := wallet.SendFunds(ctx, config.MasterURL, config.MasterXPriv, leaderTwo.Paymail, 10); err != nil {
		utils.HandleErrorAndExit("Failed to send funds to %v: %v\n", leaderTwo.Paymail, err)
	}

	leaderTwoBalance, err := wallet.GetBalance(ctx, config.MasterURL, config.MasterXPriv)
	if err != nil {
		utils.HandleErrorAndExit("Failed to get balance for master instance, error: %v\n", err)
	}
	if leaderTwoBalance < wallet.MinimalBalance {
		utils.HandleErrorAndExit("Leader instance %v has insufficient funds: %d\n", config.ClientOneURL, leaderTwoBalance)
	}

	utils.PrintOutput("Setup complete!")
}
