/*
Copyright © 2024 KATO Kanryu<k.kanryu@gmail.com>
*/
package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/kanryu/validagen/generator"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [validators.toml]",
	Short: "Generate validators",
	Long: `Generate validators with [validators.toml]
For example:

validagen generate validators.toml
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initSlog(cmd)
		vp, err := generator.ParseToml(args[0])
		if err != nil {
			panic(err)
		}

		if tmpl, err := cmd.Flags().GetString("template"); err == nil {
			vp.Template = tmpl
		}
		err = vp.Generate()
		if err != nil {
			panic(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.
	generateCmd.PersistentFlags().StringP("loglevel", "l", "info", "Output log level (default is info)")

	generateCmd.Flags().StringP("template", "t", "", "Filepath of template for generate validators")
}

func initSlog(cmd *cobra.Command) {
	level := slog.LevelInfo
	if lvl, err := cmd.PersistentFlags().GetString("loglevel"); err == nil {
		switch strings.ToLower(lvl) {
		case "debug":
			level = slog.LevelDebug
		}
	} else {
		panic(err)
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}
