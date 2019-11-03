package client

import (
	"fmt"
	"github.com/smartcontractkit/external-initiator/store"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// main enters into the cobra command as shown and implemented below.
func Run() {
	if err := generateCmd().Execute(); err != nil {
		fmt.Println(err)
	}
}

func generateCmd() *cobra.Command {
	v := viper.New()
	newcmd := &cobra.Command{
		Use:  "external-initiator [required flags]",
		Args: cobra.MaximumNArgs(0),
		Long: "Monitors external blockchains and relays events to Chainlink node. ENV variables can be set by prefixing flag with EI_: EI_ACCESSKEY",
		Run:  func(_ *cobra.Command, args []string) { runCallback(v, args, startService) },
	}

	newcmd.Flags().String("databaseurl", "postgresql://postgres:password@localhost:5432/ei?sslmode=disable", "DatabaseURL configures the URL for chainlink to connect to. This must be a properly formatted URL, with a valid scheme (postgres://).")
	must(v.BindPFlag("databaseurl", newcmd.Flags().Lookup("databaseurl")))

	newcmd.Flags().String("chainlink", "localhost:6688", "The URL of the Chainlink Core service")
	must(v.BindPFlag("chainlink", newcmd.Flags().Lookup("chainlink")))

	newcmd.Flags().String("claccesskey", "", "The access key to identity the node to Chainlink")
	must(v.BindPFlag("claccesskey", newcmd.Flags().Lookup("claccesskey")))

	newcmd.Flags().String("clsecret", "", "The secret to authenticate the node to Chainlink")
	must(v.BindPFlag("clsecret", newcmd.Flags().Lookup("clsecret")))

	v.SetEnvPrefix("EI")
	v.AutomaticEnv()

	return newcmd
}

var requiredConfig = []string{
	"chainlink",
	"claccesskey",
	"clsecret",
	"databaseurl",
}

// runner type matches the function signature of synchronizeForever
type runner = func(Config, *store.Client)

func runCallback(v *viper.Viper, args []string, runner runner) {
	if err := validateParams(v, args, requiredConfig); err != nil {
		log.Warn(err.Error())
		return
	}

	config := newConfigFromViper(v)

	db, err := store.ConnectToDb(config.DatabaseURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	runner(config, db)
}

func validateParams(v *viper.Viper, args []string, required []string) error {
	var missing []string
	for _, k := range required {
		if v.GetString(k) == "" {
			msg := fmt.Sprintf("%s flag or EI_%s env must be set", k, strings.ToUpper(k))
			fmt.Println(msg)
			missing = append(missing, msg)
		}
	}
	if len(missing) > 0 {
		return errors.New(strings.Join(missing, ","))
	}
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
