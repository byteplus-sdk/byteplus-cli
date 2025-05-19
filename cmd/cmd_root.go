/*
 * // Copyright (c) 2024 Bytedance Ltd. and/or its affiliates
 * //
 * // Licensed under the Apache License, Version 2.0 (the "License");
 * // you may not use this file except in compliance with the License.
 * // You may obtain a copy of the License at
 * //
 * //	http://www.apache.org/licenses/LICENSE-2.0
 * //
 * // Unless required by applicable law or agreed to in writing, software
 * // distributed under the License is distributed on an "AS IS" BASIS,
 * // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * // See the License for the specific language governing permissions and
 * // limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "bp",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Usage()
		return nil
	},
	ValidArgs:     rootSupport.GetAllSvcCompatible(),
	SilenceErrors: true,
	SilenceUsage:  true,
}

func initRootCmd() {

	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.Flags().BoolP("help", "h", false, "")

	// todo enable color?
	rootCmd.SetUsageTemplate(rootUsageTemplate())

	rootCmd.AddCommand(&cobra.Command{
		Use: "enable-color",
		Run: func(cmd *cobra.Command, args []string) {
			config.EnableColor = true
			WriteConfigToFile(config)
		},
		Hidden: true,
	}, &cobra.Command{
		Use: "disable-color",
		Run: func(cmd *cobra.Command, args []string) {
			config.EnableColor = false
			WriteConfigToFile(config)
		},
		Hidden: true,
	})
}

func Execute() {
	initRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.CommandPath}} [service]{{end}} [action] [params] {{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [service] --help" for more information about a service.{{end}}
`
}
