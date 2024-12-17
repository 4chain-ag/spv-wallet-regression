package main

import (
	"context"
	"fmt"
	"os"

	"github.com/4chain-AG/spv-wallet-regression/internal/utils"
	"github.com/4chain-AG/spv-wallet-regression/internal/wallet"
)

const (
	adminXPriv         = "xprv9s21ZrQH143K3CbJXirfrtpLvhT3Vgusdo8coBritQ3rcS7Jy7sxWhatuxG5h2y1Cqj8FKmPp69536gmjYRpfga2MJdsGyBsnB12E19CESK"
	adminXPub          = "xpub661MyMwAqRbcFgfmdkPgE2m5UjHXu9dj124DbaGLSjaqVESTWfCD4VuNmEbVPkbYLCkykwVZvmA8Pbf8884TQr1FgdG2nPoHR8aB36YdDQh"
	leaderPaymailAlias = "leader"
	minimalBalance     = 9
)

func main() {
	ctx := context.Background()

	config, err := wallet.LoadConfig()
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	leaderOne, err := wallet.CreateUser(ctx, config.ClientOneURL, config.ClientOneLeaderXPriv, adminXPriv, adminXPub, leaderPaymailAlias)
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Failed to create leader user for %v, error: %v\n", config.ClientOneURL, err)
		os.Exit(1)
	}

	leaderTwo, err := wallet.CreateUser(ctx, config.ClientTwoURL, config.ClientTwoLeaderXPriv, adminXPriv, adminXPub, leaderPaymailAlias)
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Failed to create leader user for %v, error: %v\n", config.ClientTwoURL, err)
		os.Exit(1)
	}

	if _, err := wallet.SendFunds(ctx, config.MasterURL, config.MasterXPriv, leaderOne.Paymail, 10); err != nil {
		fmt.Fprintf(utils.StdErr, "Failed to send funds to %v: %v\n", leaderOne.Paymail, err)
		os.Exit(1)
	}

	leaderOneBalance, err := wallet.GetBalance(ctx, config.ClientOneURL, leaderOne.XPriv)
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Failed to get balance for master instance, error: %v\n", err)
		os.Exit(1)
	}

	if leaderOneBalance < minimalBalance {
		fmt.Fprintf(utils.StdErr, "Leader instance %v has insufficient funds: %d\n", config.ClientOneURL, leaderOneBalance)
		os.Exit(1)
	}

	if _, err := wallet.SendFunds(ctx, config.MasterURL, config.MasterXPriv, leaderTwo.Paymail, 10); err != nil {
		fmt.Fprintf(utils.StdErr, "Failed to send funds to %v: %v\n", leaderTwo.Paymail, err)
		os.Exit(1)
	}

	leaderTwoBalance, err := wallet.GetBalance(ctx, config.MasterURL, config.MasterXPriv)
	if err != nil {
		fmt.Fprintf(utils.StdErr, "Failed to get balance for master instance, error: %v\n", err)
		os.Exit(1)
	}
	if leaderTwoBalance < minimalBalance {
		fmt.Fprintf(utils.StdErr, "Leader instance %v has insufficient funds: %d\n", config.ClientOneURL, leaderTwoBalance)
		os.Exit(1)
	}

	fmt.Println("Setup complete!")
}
