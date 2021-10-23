package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	downCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var downCmd = &cobra.Command{
	Use:     "down [UNIT...]",
	Short:   "uninstalls the squadron or given units",
	Example: "  squadron down frontend backend --namespace demo",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return down(cmd.Context(), args, cwd, flagNamespace, flagFiles)
	},
}

func down(ctx context.Context, args []string, cwd, namespace string, files []string) error {
	sq := squadron.New(cwd, namespace, files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)
	units, err := parseUnitArgs(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	return sq.Down(ctx, units, helmArgs)
}
