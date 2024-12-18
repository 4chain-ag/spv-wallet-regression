package wallet

import (
	"fmt"

	"github.com/4chain-AG/spv-wallet-regression/internal/utils"
)

// Config holds the configuration for the regression test.
type Config struct {
	ClientOneURL         string
	ClientTwoURL         string
	ClientOneLeaderXPriv string
	ClientTwoLeaderXPriv string
	MasterURL            string
	MasterXPriv          string
}

// LoadConfig loads all required environment variables into a Config struct
func LoadConfig() (*Config, error) {

	requiredEnvVars := map[string]string{
		MASTER_INSTANCE_URL:     "",
		MASTER_INSTANCE_XPRIV:   "",
		CLIENT_ONE_URL:          "",
		CLIENT_TWO_URL:          "",
		CLIENT_ONE_LEADER_XPRIV: "",
		CLIENT_TWO_LEADER_XPRIV: "",
	}

	for env, _ := range requiredEnvVars {
		value, err := utils.GetEnv(env)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve environment variable %s: %w", env, err)
		}
		requiredEnvVars[env] = value
	}

	return &Config{
		MasterURL:            utils.AddPrefixIfNeeded(requiredEnvVars[MASTER_INSTANCE_URL]),
		MasterXPriv:          requiredEnvVars[MASTER_INSTANCE_XPRIV],
		ClientOneURL:         utils.AddPrefixIfNeeded(requiredEnvVars[CLIENT_ONE_URL]),
		ClientTwoURL:         utils.AddPrefixIfNeeded(requiredEnvVars[CLIENT_TWO_URL]),
		ClientOneLeaderXPriv: requiredEnvVars[CLIENT_ONE_LEADER_XPRIV],
		ClientTwoLeaderXPriv: requiredEnvVars[CLIENT_TWO_LEADER_XPRIV],
	}, nil
}
