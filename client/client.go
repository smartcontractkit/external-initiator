// Package client provides the core functionality
// to Run an External Initiator.
package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Run enters into the cobra command to start the external initiator.
func Run() {
	if err := generateCmd().Execute(); err != nil {
		logger.Error(err)
	}
}

func generateCmd() *cobra.Command {
	v := viper.New()
	newcmd := &cobra.Command{
		Use:  "external-initiator [endpoint configs]",
		Args: cobra.MinimumNArgs(0),
		Long: "Monitors external blockchains and relays events to Chainlink node. Supplying endpoint configs as args will delete all other stored configs. ENV variables can be set by prefixing flag with EI_: EI_ACCESSKEY",
		Run:  func(_ *cobra.Command, args []string) { runCallback(v, args, startService) },
	}

	newcmd.Flags().String("databaseurl", "postgresql://postgres:password@localhost:5432/ei?sslmode=disable", "DatabaseURL configures the URL for external initiator to connect to. This must be a properly formatted URL, with a valid scheme (postgres://).")
	must(v.BindPFlag("databaseurl", newcmd.Flags().Lookup("databaseurl")))

	newcmd.Flags().String("chainlinkurl", "localhost:6688", "The URL of the Chainlink Core Service")
	must(v.BindPFlag("chainlinkurl", newcmd.Flags().Lookup("chainlinkurl")))

	newcmd.Flags().String("ic_accesskey", "", "The Chainlink access key, used for traffic flowing from this Service to Chainlink")
	must(v.BindPFlag("ic_accesskey", newcmd.Flags().Lookup("ic_accesskey")))

	newcmd.Flags().String("ic_secret", "", "The Chainlink secret, used for traffic flowing from this Service to Chainlink")
	must(v.BindPFlag("ic_secret", newcmd.Flags().Lookup("ic_secret")))

	newcmd.Flags().String("ci_accesskey", "", "The External Initiator access key, used for traffic flowing from Chainlink to this Service")
	must(v.BindPFlag("ci_accesskey", newcmd.Flags().Lookup("ci_accesskey")))

	newcmd.Flags().String("ci_secret", "", "The External Initiator secret, used for traffic flowing from Chainlink to this Service")
	must(v.BindPFlag("ci_secret", newcmd.Flags().Lookup("ci_secret")))

	newcmd.Flags().Bool("mock", false, "Set to true if the External Initiator should expect mock events from the blockchains")
	must(v.BindPFlag("mock", newcmd.Flags().Lookup("mock")))

	newcmd.Flags().Duration("cl_timeout", 5*time.Second, "The timeout for job run triggers to the Chainlink node")
	must(v.BindPFlag("cl_timeout", newcmd.Flags().Lookup("cl_timeout")))

	newcmd.Flags().Uint("cl_retry_attempts", 3, "The maximum number of attempts that will be made for job run triggers")
	must(v.BindPFlag("cl_retry_attempts", newcmd.Flags().Lookup("cl_retry_attempts")))

	newcmd.Flags().Duration("cl_retry_delay", 1*time.Second, "The delay between attempts for job run triggers")
	must(v.BindPFlag("cl_retry_delay", newcmd.Flags().Lookup("cl_retry_delay")))

	newcmd.Flags().Int64("keeper_block_cooldown", 3, "Number of blocks to cool down before triggering a new run for a Keeper job")
	must(v.BindPFlag("keeper_block_cooldown", newcmd.Flags().Lookup("keeper_block_cooldown")))

	v.SetEnvPrefix("EI")
	v.AutomaticEnv()

	return newcmd
}

var requiredConfig = []string{
	"chainlinkurl",
	"ic_accesskey",
	"ic_secret",
	"databaseurl",
	"ci_accesskey",
	"ci_secret",
	"cl_timeout",
	"cl_retry_attempts",
	"cl_retry_delay",
}

// runner type matches the function signature of synchronizeForever
type runner = func(Config, *store.Client, []string)

func runCallback(v *viper.Viper, args []string, runner runner) {
	err := validateParams(v, args, requiredConfig)
	if err != nil {
		logger.Error(err)
	}

	config := newConfigFromViper(v)

	db, err := store.ConnectToDb(config.DatabaseURL)
	if err != nil {
		logger.Error(err)
		return
	}
	defer db.Close()

	runner(config, db, args)
}

func validateParams(v *viper.Viper, args []string, required []string) error {
	var missing []string
	for _, k := range required {
		if v.GetString(k) == "" {
			msg := fmt.Sprintf("%s flag or EI_%s env must be set", k, strings.ToUpper(k))
			logger.Error(msg)
			missing = append(missing, msg)
		}
	}
	if len(missing) > 0 {
		return errors.New(strings.Join(missing, ","))
	}

	for _, a := range args {
		var config store.Endpoint
		err := json.Unmarshal([]byte(a), &config)
		if err != nil {
			msg := fmt.Sprintf("Invalid endpoint configuration provided: %v", a)
			logger.Error(msg)
			return errors.Wrap(err, msg)
		}
	}

	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
