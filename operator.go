package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	walletclient "github.com/bitcoin-sv/spv-wallet-go-client"
	"github.com/bitcoin-sv/spv-wallet-go-client/xpriv"
	"github.com/bitcoin-sv/spv-wallet/models"
)

const (
	atSign                   = "@"
	domainPrefix             = "https://"
	adminXPriv               = "xprv9s21ZrQH143K3CbJXirfrtpLvhT3Vgusdo8coBritQ3rcS7Jy7sxWhatuxG5h2y1Cqj8FKmPp69536gmjYRpfga2MJdsGyBsnB12E19CESK"
	adminXPub                = "xpub661MyMwAqRbcFgfmdkPgE2m5UjHXu9dj124DbaGLSjaqVESTWfCD4VuNmEbVPkbYLCkykwVZvmA8Pbf8884TQr1FgdG2nPoHR8aB36YdDQh"
	leaderPaymailAlias       = "leader"
	domainSuffixSharedConfig = "/v1/shared-config"
	minimalBalance           = 9

	clientOneURLEnvVar         = "CLIENT_ONE_URL"
	clientTwoURLEnvVar         = "CLIENT_TWO_URL"
	clientOneLeaderXPrivEnvVar = "CLIENT_ONE_LEADER_XPRIV"
	clientTwoLeaderXPrivEnvVar = "CLIENT_TWO_LEADER_XPRIV"

	masterInstanceURL   = "MASTER_INSTANCE_URL"
	masterInstanceXPriv = "MASTER_INSTANCE_XPRIV"
)

var (
	explicitHTTPURLRegex = regexp.MustCompile(`^https?://`)
)

type regressionTestConfig struct {
	clientOneURL         string
	clientTwoURL         string
	clientOneLeaderXPriv string
	clientTwoLeaderXPriv string
	masterURL            string
	masterXPriv          string
}

type regressionTestUser struct {
	XPriv   string `json:"xpriv"`
	XPub    string `json:"xpub"`
	Paymail string `json:"paymail"`
}

func main() {
	ctx := context.Background()
	config := loadConfig()

	leaderOne, err := createUser(ctx, config.clientOneURL, config.clientOneLeaderXPriv)
	if err != nil {
		fmt.Printf("Failed to create leader user for %v, error: %v\n", config.clientOneURL, err)
		os.Exit(1)
	}
	leaderTwo, err := createUser(ctx, config.clientTwoURL, config.clientTwoLeaderXPriv)
	if err != nil {
		fmt.Printf("Failed to create leader user for %v, error: %v\n", config.clientTwoURL, err)
		os.Exit(1)
	}

	masterBalance, err := getBalance(ctx, config.masterURL, config.masterXPriv)
	if err != nil {
		fmt.Printf("Failed to get balance for master instance, error: %v\n", err)
		os.Exit(1)
	}

	if masterBalance < 2*minimalBalance {
		fmt.Printf("Master instance has insufficient funds: %d\n", masterBalance)
		os.Exit(1)
	}

	_, err = sendFunds(ctx, config.masterURL, config.masterXPriv, leaderOne.Paymail, 10)
	if err != nil {
		fmt.Printf("Failed to send funds from master instance to leader instance %v, error: %v\n", leaderOne.Paymail, err)
		os.Exit(1)
	}

	leaderOneBalance, err := getBalance(ctx, config.clientOneURL, leaderOne.XPriv)
	if err != nil {
		fmt.Printf("Failed to get balance for master instance, error: %v\n", err)
		os.Exit(1)
	}

	if leaderOneBalance < minimalBalance {
		fmt.Printf("Leader instance %v has insufficient funds: %d\n", config.clientOneURL, leaderOneBalance)
		os.Exit(1)
	}

	_, err = sendFunds(ctx, config.masterURL, config.masterXPriv, leaderTwo.Paymail, 10)
	if err != nil {
		fmt.Printf("Failed to send funds from master instance to leader instance %v, error: %v\n", leaderTwo.Paymail, err)
		os.Exit(1)
	}

	leaderTwoBalance, err := getBalance(ctx, config.masterURL, config.masterXPriv)
	if err != nil {
		fmt.Printf("Failed to get balance for master instance, error: %v\n", err)
		os.Exit(1)
	}
	if leaderTwoBalance < minimalBalance {
		fmt.Printf("Leader instance %v has insufficient funds: %d\n", config.clientOneURL, leaderTwoBalance)
		os.Exit(1)
	}
}

func getPaymailDomain(xpub string, instanceURL string) (string, error) {
	apiURL := instanceURL + domainSuffixSharedConfig
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set(models.AuthHeader, xpub)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get shared config: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var configResponse models.SharedConfig
	if err := json.Unmarshal(body, &configResponse); err != nil {
		return "", err
	}

	if len(configResponse.PaymailDomains) != 1 {
		return "", fmt.Errorf("expected 1 paymail domain, got %d", len(configResponse.PaymailDomains))
	}
	return configResponse.PaymailDomains[0], nil
}

func createUser(ctx context.Context, instanceUrl string, userXpriv string) (*regressionTestUser, error) {
	keys, err := xpriv.FromString(userXpriv)
	if err != nil {
		return nil, err
	}

	paymailDomain, err := getPaymailDomain(adminXPub, instanceUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared config for %v: %w", paymailDomain, err)
	}

	user := &regressionTestUser{
		XPriv:   keys.XPriv(),
		XPub:    keys.XPub().String(),
		Paymail: preparePaymail(leaderPaymailAlias, paymailDomain),
	}

	adminClient := walletclient.NewWithAdminKey(addPrefixIfNeeded(instanceUrl), adminXPriv)

	if err := adminClient.AdminNewXpub(ctx, user.XPub, map[string]any{"some_metadata": "remove"}); err != nil {
		return nil, err
	}

	_, err = adminClient.AdminCreatePaymail(ctx, user.XPub, user.Paymail, "Regression tests", "")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func preparePaymail(paymailAlias string, domain string) string {
	if isValidURL(domain) {
		splitedDomain := strings.SplitAfter(domain, "//")
		domain = splitedDomain[1]
	}
	url := paymailAlias + atSign + domain
	return url
}

// isValidURL validates the URL if it has http or https prefix.
func isValidURL(rawURL string) bool {
	return explicitHTTPURLRegex.MatchString(rawURL)
}

// addPrefixIfNeeded adds the HTTPS prefix to the URL if it is missing.
func addPrefixIfNeeded(url string) string {
	if !isValidURL(url) {
		return domainPrefix + url
	}
	return url
}

// sendFunds sends funds from one paymail to another.
func sendFunds(ctx context.Context, fromInstance string, fromXPriv string, toPamail string, howMuch int) (*models.Transaction, error) {
	client := walletclient.NewWithXPriv(fromInstance, fromXPriv)

	balance, err := getBalance(ctx, fromInstance, fromXPriv)
	if err != nil {
		return nil, err
	}
	if balance < howMuch {
		return nil, fmt.Errorf("insufficient funds: %d", balance)
	}

	recipient := walletclient.Recipients{To: toPamail, Satoshis: uint64(howMuch)}
	recipients := []*walletclient.Recipients{&recipient}
	metadata := map[string]any{
		"description": "regression-test",
	}

	transaction, err := client.SendToRecipients(ctx, recipients, metadata)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func getBalance(ctx context.Context, fromInstance string, fromXPriv string) (int, error) {
	client := walletclient.NewWithXPriv(addPrefixIfNeeded(fromInstance), fromXPriv)

	xpubInfo, err := client.GetXPub(ctx)
	if err != nil {
		return -1, err
	}
	return int(xpubInfo.CurrentBalance), nil
}

func loadConfig() *regressionTestConfig {
	masterURL := os.Getenv(masterInstanceURL)
	if masterURL == "" {
		fmt.Println(masterInstanceURL, "environment variable is not set")
		os.Exit(1)
	}

	masterXPriv := os.Getenv(masterInstanceXPriv)
	if masterXPriv == "" {
		fmt.Println(masterInstanceXPriv, "environment variable is not set")
		os.Exit(1)
	}

	clientOneURL := os.Getenv(clientOneURLEnvVar)
	if clientOneURL == "" {
		fmt.Println(clientOneURLEnvVar, "environment variable is not set")
		os.Exit(1)
	}

	clientTwoURL := os.Getenv(clientTwoURLEnvVar)
	if clientTwoURL == "" {
		fmt.Println(clientTwoURLEnvVar, "environment variable is not set")
		os.Exit(1)
	}

	clientOneLeaderXPriv := os.Getenv(clientOneLeaderXPrivEnvVar)
	if clientOneLeaderXPriv == "" {
		fmt.Println(clientOneLeaderXPrivEnvVar, "environment variable is not set")
		os.Exit(1)
	}

	clientTwoLeaderXPriv := os.Getenv(clientTwoLeaderXPrivEnvVar)
	if clientTwoLeaderXPriv == "" {
		fmt.Println(clientTwoLeaderXPrivEnvVar, "environment variable is not set")
		os.Exit(1)
	}

	return &regressionTestConfig{
		clientOneURL:         addPrefixIfNeeded(clientOneURL),
		clientOneLeaderXPriv: clientOneLeaderXPriv,
		clientTwoURL:         addPrefixIfNeeded(clientTwoURL),
		clientTwoLeaderXPriv: clientTwoLeaderXPriv,
		masterURL:            addPrefixIfNeeded(masterURL),
		masterXPriv:          masterXPriv,
	}
}
