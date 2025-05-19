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
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	generateServiceCommands()
}

func generateServiceCommands() {
	for svc, actionMeta := range rootSupport.SupportAction {
		apiMetas := rootSupport.SupportTypes[svc]
		svcCmd := &cobra.Command{
			Use: svc,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
			Args: cobra.MatchAll(cobra.OnlyValidArgs),
		}

		svcCmd.SetUsageTemplate(serviceUsageTemplate())
		svcCmd.ValidArgs = rootSupport.GetAllAction(svc)

		actionCmds := generateActionCmd(actionMeta, apiMetas)
		for i := 0; i < len(actionCmds); i++ {
			svcCmd.AddCommand(actionCmds[i])
		}

		svcCmd.Flags().BoolP("help", "h", false, "")

		rootCmd.AddCommand(svcCmd)

		for _, v := range compatible_support_cmd {
			if strings.ReplaceAll(v, "_", "") == svc {
				//copy a non ptr value from svcCmd for compatible svc cmd with _
				compatibleCmd := *svcCmd
				compatibleCmd.Use = v
				compatibleCmd.Hidden = true
				rootCmd.AddCommand(&compatibleCmd)
			}
		}
	}
}

func serviceUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.CommandPath}} [action]{{end}} [params] {{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Actions:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Use "{{.CommandPath}} [action] --help" for more information about a action.{{end}}
`
}
