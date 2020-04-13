package actions

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/foomo/configurd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use: "configurd",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			wdir, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			if flagDir != "" {
				flagDir = path.Join(wdir, flagDir)
			} else {
				flagDir = wdir
			}
			if cmd.Name() == "help" || cmd.Name() == "init" {
				return
			}
		},
	}

	flagTag       string
	flagDir       string
	flagVerbose   bool
	flagNamespace string
)

func mustNewConfigurd() configurd.Configurd {
	cnf, err := configurd.New(log, flagDir)
	if err != nil {
		log.Fatal(err)
	}
	return cnf
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagTag, "tag", "t", "latest", "Specifies the image tag")
	rootCmd.PersistentFlags().StringVarP(&flagDir, "dir", "d", "", "Specifies working directory")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Specifies should command output be displayed")
	rootCmd.AddCommand(buildCmd, installCmd, uninstallCmd, initCmd, versionCmd)
}

func Execute() {
	rootCmd.Execute()
}

func outputErrorf(output string, err error, format string, args ...interface{}) error {
	return fmt.Errorf("%v\nerror: %v\noutput: %v", fmt.Sprintf(format, args...), err, strings.ReplaceAll(output, "\n", " "))
}
