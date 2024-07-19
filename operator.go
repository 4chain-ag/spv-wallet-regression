package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bitcoin-sv/spv-wallet/models"
)

const (
	adminXPriv               = "xprv9s21ZrQH143K3CbJXirfrtpLvhT3Vgusdo8coBritQ3rcS7Jy7sxWhatuxG5h2y1Cqj8FKmPp69536gmjYRpfga2MJdsGyBsnB12E19CESK"
	adminXPub                = "xpub661MyMwAqRbcFgfmdkPgE2m5UjHXu9dj124DbaGLSjaqVESTWfCD4VuNmEbVPkbYLCkykwVZvmA8Pbf8884TQr1FgdG2nPoHR8aB36YdDQh"
	leaderPaymailAlias       = "leader"
	domainSuffixSharedConfig = "/v1/shared-config"
	minimalBalance           = 100

	clientOneURLEnvVar         = "CLIENT_ONE_URL"
	clientTwoURLEnvVar         = "CLIENT_TWO_URL"
	clientOneLeaderXPrivEnvVar = "CLIENT_ONE_LEADER_XPRIV"
	clientTwoLeaderXPrivEnvVar = "CLIENT_TWO_LEADER_XPRIV"
)

type regressionTestConfig struct {
	clientOneURL           string
	clientTwoURL           string
	clientOnePaymailDomain string
	clientTwoPaymailDomain string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: operator <sqlite_url> <postgres_url>")
		os.Exit(1)
	}

	config := loadConfig()
	paymailDomainClientOne, err := getPaymailDomain(adminXPub, config.clientOneURL)
	if err != nil {
		fmt.Println("Failed to get shared config for client one:", err)
		os.Exit(1)
	}

	paymailDomainClientTwo, err := getPaymailDomain(adminXPub, config.clientTwoURL)
	if err != nil {
		fmt.Println("Failed to get shared config for client two:", err)
		os.Exit(1)
	}

	config.clientOnePaymailDomain = paymailDomainClientOne
	config.clientTwoPaymailDomain = paymailDomainClientTwo

	// create leader accounts
	// send them some money ->> repository spv-wallet-regression tests should have master instance url ENV set + xpriv env from we can get money
	// create this account and send money here: https://spv-wallet.test.4chain.space/
	// check balance
	// set envs
	// end

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

func loadConfig() *regressionTestConfig {
	return &regressionTestConfig{
		clientOneURL: os.Args[1],
		clientTwoURL: os.Args[2],
	}
}
