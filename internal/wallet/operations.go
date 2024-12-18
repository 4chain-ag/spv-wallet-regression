package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	walletclient "github.com/bitcoin-sv/spv-wallet-go-client"
	"github.com/bitcoin-sv/spv-wallet-go-client/xpriv"
	"github.com/bitcoin-sv/spv-wallet/models"
)

// User represents a wallet user with key and paymail info.
type User struct {
	XPriv   string
	XPub    string
	Paymail string
}

// CreateUser creates a wallet user and sets up a paymail.
func CreateUser(ctx context.Context, instanceURL, userXPriv, adminXPriv, adminXPub, alias string) (*User, error) {
	keys, err := xpriv.FromString(userXPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XPriv: %w", err)
	}

	paymailDomain, err := getPaymailDomain(instanceURL, adminXPub)
	if err != nil {
		return nil, fmt.Errorf("failed to get paymail domain for %v: %w", paymailDomain, err)
	}

	user := &User{
		XPriv:   keys.XPriv(),
		XPub:    keys.XPub().String(),
		Paymail: fmt.Sprintf("%s@%s", alias, paymailDomain),
	}

	adminClient, err := walletclient.NewWithAdminKey(instanceURL, adminXPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}

	if err := adminClient.AdminNewXpub(ctx, user.XPub, map[string]any{"some_metadata": "remove"}); err != nil {
		return nil, fmt.Errorf("failed to create XPub: %w", err)
	}

	if _, err := adminClient.AdminCreatePaymail(ctx, user.XPub, user.Paymail, "Regression Test", ""); err != nil {
		return nil, fmt.Errorf("failed to create paymail: %w", err)
	}

	return user, nil
}

// GetBalance retrieves the current balance.
func GetBalance(ctx context.Context, instanceURL, fromXPriv string) (int, error) {
	client, err := walletclient.NewWithXPriv(instanceURL, fromXPriv)
	if err != nil {
		return -1, fmt.Errorf("failed to create client: %w", err)
	}

	xPubInfo, err := client.GetXPub(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve XPub: %w", err)
	}

	return int(xPubInfo.CurrentBalance), nil
}

// SendFunds transfers funds to a specified paymail.
func SendFunds(ctx context.Context, fromURL, fromXPriv, toPaymail string, amount int) (*models.Transaction, error) {
	client, err := walletclient.NewWithXPriv(fromURL, fromXPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	balance, err := GetBalance(ctx, fromURL, fromXPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if balance < amount {
		return nil, fmt.Errorf("insufficient funds: %d", balance)
	}

	recipient := walletclient.Recipients{To: toPaymail, Satoshis: uint64(amount)}
	metadata := map[string]any{
		"description": "regression-test",
	}
	tx, err := client.SendToRecipients(ctx, []*walletclient.Recipients{&recipient}, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to send funds: %w", err)
	}
	return tx, nil
}

func getPaymailDomain(instanceURL, adminXPub string) (string, error) {
	apiURL, err := url.JoinPath(instanceURL, sharedConfigURI)
	if err != nil {
		return "", fmt.Errorf("failed to join URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set(models.AuthHeader, adminXPub)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get paymail domain: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get shared config: %s", resp.Status)
	}

	var config models.SharedConfig
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &config); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(config.PaymailDomains) != 1 {
		return "", fmt.Errorf("expected 1 paymail domain, got %d - [%v]", len(config.PaymailDomains), config.PaymailDomains)
	}

	return config.PaymailDomains[0], nil
}
