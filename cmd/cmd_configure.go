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
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	profileFlags    Profile
	ssoSessionFlags SsoSession
	ssoFlags        Profile
)

const defaultSsoRegion = "ap-southeast-1"

var defaultRegistrationScopes = []string{"cloudidentity:account:access", "offline_access"}
var allowedRegistrationScopes = []string{"cloudidentity:account:access", "offline_access"}
var allowedRegistrationScopesSet = map[string]struct{}{
	"cloudidentity:account:access": {},
	"offline_access":               {},
}

func init() {
	configureCmd := newConfigureRootCmd()

	configureCmd.AddCommand(newConfigureGetCmd())
	configureCmd.AddCommand(newConfigureListCmd())
	configureCmd.AddCommand(newConfigureDeleteCmd())
	configureCmd.AddCommand(newConfigureProfileCmd())
	configureCmd.AddCommand(newConfigureSetCmd())
	configureCmd.AddCommand(newConfigureSsoSessionCmd())
	configureCmd.AddCommand(newConfigureSsoCmd())

	rootCmd.AddCommand(configureCmd)
}

func newConfigureRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "configure",
		Args: cobra.MatchAll(cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	cmd.SetUsageTemplate(configureUsageTemplate())

	return cmd
}

func newConfigureGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := cmd.Flag("profile").Value.String()
			return getConfigProfile(profileName)
		},
		Short: "show target profile's information",
		Long: `Description:
  show target profile's information
  if no profile name specified, show default profile`,
		DisableFlagsInUseLine: true,
	}

	cmd.SetUsageTemplate(configureActionUsageTemplate())

	cmd.Flags().StringVar(&profileFlags.Name, "profile", "", "target profile name")
	cmd.Flags().BoolP("help", "h", false, "")

	return cmd
}

func newConfigureSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "set",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfigProfile(&profileFlags)
		},
		Short: "add new profile, or modify target profile",
		Long: `Description:
  add new profile, or modify target profile:
      1. if profile not exist, add new;
      2. if profile exist, modify target field`,
		DisableFlagsInUseLine: true,
	}

	cmd.SetUsageTemplate(configureActionUsageTemplate())

	cmd.Flags().StringVar(&profileFlags.Name, "profile", "", "target profile name")
	cmd.Flags().StringVar(&profileFlags.AccessKey, "access-key", "", "your access key(AK)")
	cmd.Flags().StringVar(&profileFlags.SecretKey, "secret-key", "", "your secret key(SK)")
	cmd.Flags().StringVar(&profileFlags.Region, "region", "", "your region")
	cmd.Flags().StringVar(&profileFlags.Endpoint, "endpoint", "", "endpoint bind with region")
	cmd.Flags().StringVar(&profileFlags.EndpointResolver, "endpoint-resolver", "", "endpoint resolver (auto-addressing)")
	cmd.Flags().StringVar(&profileFlags.SessionToken, "session-token", "", "your session token")
	cmd.Flags().StringVar(&profileFlags.SsoSessionName, "sso-session", "", "your sso session name")

	profileFlags.DisableSSL = cmd.Flags().Bool("disable-ssl", false, "disable ssl")
	profileFlags.UseDualStack = cmd.Flags().Bool("use-dual-stack", false, "use dual-stack endpoints")
	cmd.Flags().BoolP("help", "h", false, "")

	cmd.MarkFlagRequired("profile")
	cmd.MarkFlagRequired("region")

	return cmd
}

func newConfigureListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listConfigProfiles()
		},
		Short: "list all profiles",
		Long: `Description:
  list all profiles`,
		DisableFlagsInUseLine: true,
	}

	cmd.SetUsageTemplate(configureActionUsageTemplate())

	cmd.Flags().BoolP("help", "h", false, "")

	return cmd
}

func newConfigureDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := cmd.Flag("profile").Value.String()
			return deleteConfigProfile(profileName)
		},
		Short: "delete target profile",
		Long: `Description:
  delete target profile`,
		DisableFlagsInUseLine: true,
	}

	cmd.SetUsageTemplate(configureActionUsageTemplate())

	cmd.Flags().StringVar(&profileFlags.Name, "profile", "", "target profile name")
	cmd.Flags().BoolP("help", "h", false, "")

	cmd.MarkFlagRequired("profile")

	return cmd
}

func newConfigureProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := cmd.Flag("profile").Value.String()
			return changeConfigProfile(profileName)
		},
		Short: "change target profile",
		Long: `Description:
  change target profile`,
		DisableFlagsInUseLine: true,
	}

	cmd.SetUsageTemplate(configureActionUsageTemplate())

	cmd.Flags().StringVar(&profileFlags.Name, "profile", "", "target profile name")
	cmd.Flags().BoolP("help", "h", false, "")

	cmd.MarkFlagRequired("profile")

	return cmd
}

func newConfigureSsoSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "sso-session",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ctx.config
			if cfg == nil {
				cfg = &Configure{
					Profiles:   make(map[string]*Profile),
					SsoSession: make(map[string]*SsoSession),
				}
				ctx.config = cfg
			}
			if cfg.SsoSession == nil {
				cfg.SsoSession = make(map[string]*SsoSession)
			}

			var existingSession *SsoSession
			if strings.TrimSpace(ssoSessionFlags.Name) == "" {
				name, selected, err := promptSessionName(cfg, "")
				if err != nil {
					return err
				}
				ssoSessionFlags.Name = name
				existingSession = selected
			} else {
				ssoSessionFlags.Name = strings.TrimSpace(ssoSessionFlags.Name)
				existingSession = cfg.SsoSession[ssoSessionFlags.Name]
			}

			defaultStartURL := ""
			defaultRegion := defaultSsoRegion
			defaultScopes := []string(nil)
			if existingSession != nil {
				defaultStartURL = existingSession.StartURL
				defaultRegion = existingSession.Region
				defaultScopes = existingSession.RegistrationScopes
			}

			if err := promptForRequiredStringWithDefault(&ssoSessionFlags.StartURL, "Please enter SSO Start URL:", "SSO Start URL", defaultStartURL); err != nil {
				return err
			}
			if err := promptForRequiredStringWithDefault(&ssoSessionFlags.Region, "Please enter SSO region:", "SSO region", defaultRegion); err != nil {
				return err
			}

			var scopes []string
			var err error
			if len(ssoSessionFlags.RegistrationScopes) == 0 {
				showDefault := existingSession == nil
				scopes, err = promptForRegistrationScopesWithDefault(defaultScopes, showDefault)
			} else {
				scopes, err = normalizeRegistrationScopes(ssoSessionFlags.RegistrationScopes)
			}
			if err != nil {
				return err
			}
			ssoSessionFlags.RegistrationScopes = scopes

			if err := setSsoSession(&ssoSessionFlags); err != nil {
				return err
			}
			fmt.Printf("SSO session [%s] configured successfully.\n", ssoSessionFlags.Name)
			return nil
		},
		Short: "add or modify SSO session",
		Long: `Description:
  add new SSO session, or modify target SSO session:
      1. if SSO session not exist, add new;
      2. if SSO session exist, modify target field`,
		DisableFlagsInUseLine: true,
	}

	cmd.SetUsageTemplate(configureActionUsageTemplate())

	cmd.Flags().StringVar(&ssoSessionFlags.Name, "name", "", "SSO session name")
	cmd.Flags().StringVar(&ssoSessionFlags.StartURL, "start-url", "", "SSO start URL")
	cmd.Flags().StringVar(&ssoSessionFlags.Region, "region", defaultSsoRegion, "SSO region")
	cmd.Flags().StringSliceVar(&ssoSessionFlags.RegistrationScopes, "registration-scopes", defaultRegistrationScopes, "comma-separated SSO registration scopes (cloudidentity:account:access,offline_access)")
	cmd.Flags().BoolP("help", "h", false, "")

	return cmd
}

func promptForRequiredStringWithDefault(target *string, prompt, fieldName, defaultValue string) error {
	for {
		if target == nil || strings.TrimSpace(*target) == "" {
			if strings.TrimSpace(defaultValue) != "" {
				fmt.Printf("%s [%s]:", prompt, defaultValue)
				line, err := readLineAllowEmpty()
				if err != nil {
					return err
				}
				line = strings.TrimSpace(line)
				if line == "" {
					*target = defaultValue
				} else {
					*target = line
				}
			} else {
				fmt.Printf("%s", prompt)
				line, err := readLineAllowEmpty()
				if err != nil {
					return err
				}
				*target = strings.TrimSpace(line)
			}
		}
		*target = strings.TrimSpace(*target)
		if *target != "" {
			return nil
		}
		fmt.Printf("%s cannot be empty\n", fieldName)
		*target = ""
	}
}

func promptForRegistrationScopes(current []string) ([]string, error) {
	if len(current) == 0 {
		fmt.Printf("Please enter SSO registration scopes (comma-separated, allowed: %s) [%s]:", strings.Join(allowedRegistrationScopes, ", "), strings.Join(defaultRegistrationScopes, ","))
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			current = strings.Split(line, ",")
		}
	}
	return normalizeRegistrationScopes(current)
}

func promptForRegistrationScopesWithDefault(current []string, showDefault bool) ([]string, error) {
	defaultValue := strings.Join(current, ",")
	label := ""
	if showDefault {
		if defaultValue == "" {
			defaultValue = strings.Join(defaultRegistrationScopes, ",")
		}
		label = fmt.Sprintf("[%s]", defaultValue)
	} else if defaultValue != "" {
		label = fmt.Sprintf("[%s]", defaultValue)
	}
	fmt.Printf("Please enter SSO registration scopes (comma-separated, allowed: %s) %s:", strings.Join(allowedRegistrationScopes, ", "), label)
	line, err := readLineAllowEmpty()
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line != "" {
		current = strings.Split(line, ",")
	}
	return normalizeRegistrationScopes(current)
}

func normalizeRegistrationScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return append([]string(nil), defaultRegistrationScopes...), nil
	}
	seen := make(map[string]struct{})
	var normalized []string
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, ok := allowedRegistrationScopesSet[scope]; !ok {
			return nil, fmt.Errorf("invalid SSO registration scope %q, allowed values: %s", scope, strings.Join(allowedRegistrationScopes, ", "))
		}
		if _, exists := seen[scope]; !exists {
			seen[scope] = struct{}{}
			normalized = append(normalized, scope)
		}
	}
	if len(normalized) == 0 {
		return append([]string(nil), defaultRegistrationScopes...), nil
	}
	return normalized, nil
}

func newConfigureSsoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "sso",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ctx.config
			if cfg == nil {
				cfg = &Configure{
					Profiles:   make(map[string]*Profile),
					SsoSession: make(map[string]*SsoSession),
				}
				ctx.config = cfg
			}
			if cfg.Profiles == nil {
				cfg.Profiles = make(map[string]*Profile)
			}
			if cfg.SsoSession == nil {
				cfg.SsoSession = make(map[string]*SsoSession)
			}

			noBrowser, err := cmd.Flags().GetBool("no-browser")
			if err != nil {
				return err
			}

			if strings.TrimSpace(ssoFlags.Name) == "" {
				fmt.Print("Enter profile name (press Enter to use default: {sso-role-name}-{sso-account-id}): ")
				line, err := readLineAllowEmpty()
				if err != nil {
					return err
				}
				ssoFlags.Name = line
			}

			profile := &Profile{
				Name: ssoFlags.Name,
			}

			if inputProfile := cfg.Profiles[ssoFlags.Name]; inputProfile != nil {
				if strings.ToLower(strings.TrimSpace(inputProfile.Mode)) != ModeSSO {
					return fmt.Errorf("the profile [%v] already exists and is not an SSO profile. Overwriting a non-SSO profile is not permitted", ssoFlags.Name)
				}
				profile = inputProfile
			}

			var (
				name            string
				existingSession *SsoSession
			)
			if ssoFlags.SsoSessionName == "" {
				for {
					name, existingSession, err = promptSessionName(cfg, ssoFlags.SsoSessionName)
					if err == nil {
						break
					}
					if errors.Is(err, errSessionExists) {
						fmt.Println(err.Error())
						continue
					}
					return err
				}
				ssoFlags.SsoSessionName = name
			} else {
				existingSession = cfg.SsoSession[ssoFlags.SsoSessionName]
			}

			ssoSession := existingSession
			if ssoSession == nil {
				ssoSession, err = createSsoSessionInSso(ssoFlags.SsoSessionName, cfg)
				if err != nil {
					return err
				}
			}

			var sso SSOService = &Sso{
				Profile:        profile,
				SsoSessionName: ssoFlags.SsoSessionName,
				StartURL:       ssoSession.StartURL,
				Region:         ssoSession.Region,
				Scopes:         ssoSession.RegistrationScopes,
				UseDeviceCode:  true,
				NoBrowser:      noBrowser,
			}

			if err := sso.SetProfile(); err != nil {
				return err
			}
			fmt.Printf("SSO profile [%s] configured successfully.\n", profile.Name)
			return nil
		},
		Short: "configure SSO type profile",
		Long: `Description:
  configure SSO type profile with profile.mode=sso
  this command will guide you through the SSO authorization process
  and save the profile configuration to ~/.byteplus/config.json
  if the specified SSO session doesn't exist, it will be created automatically`,
		DisableFlagsInUseLine: true,
	}

	cmd.SetUsageTemplate(configureActionUsageTemplate())

	cmd.Flags().StringVar(&ssoFlags.Name, "profile", "", "profile name")
	cmd.Flags().StringVar(&ssoFlags.SsoSessionName, "sso-session", "", "SSO session name")
	cmd.Flags().Bool("no-browser", false, "Do not automatically open the browser during device authorization")
	cmd.Flags().BoolP("help", "h", false, "")

	return cmd
}

func readLineAllowEmpty() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

type sessionOption struct {
	Name    string
	Session *SsoSession
}

var errSessionExists = errors.New("SSO session already exists")

func promptSessionName(cfg *Configure, defaultName string) (string, *SsoSession, error) {
	if cfg == nil || len(cfg.SsoSession) == 0 {
		fmt.Print("Please enter SSO session name:")
		name, err := readLineAllowEmpty()
		if err != nil {
			return "", nil, err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return "", nil, fmt.Errorf("SSO session name cannot be empty")
		}
		return name, nil, nil
	}

	options := buildSessionOptions(cfg.SsoSession)
	selectedName, selectedSession, err := runSessionSelect(cfg, options, defaultName)
	if err != nil {
		return "", nil, err
	}

	return selectedName, selectedSession, nil
}

func buildSessionOptions(all map[string]*SsoSession) []sessionOption {
	var keys []string
	for name := range all {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	var out []sessionOption
	for _, name := range keys {
		out = append(out, sessionOption{
			Name:    name,
			Session: all[name],
		})
	}
	return out
}

const addNewSessionLabel = "<Create new session>"

func runSessionSelect(cfg *Configure, options []sessionOption, defaultName string) (string, *SsoSession, error) {
	choices := make([]sessionOption, 0, len(options)+1)
	choices = append(choices, options...)
	choices = append(choices, sessionOption{Name: addNewSessionLabel, Session: nil})

	var lastSearchInput string
	searcher := func(input string, index int) bool {
		if index < 0 || index >= len(choices) {
			return false
		}
		rawInput := strings.TrimSpace(input)
		lastSearchInput = rawInput
		lowerInput := strings.ToLower(rawInput)
		item := choices[index]
		if item.Name == addNewSessionLabel {
			return true
		}
		content := strings.ToLower(item.Name)
		if item.Session != nil {
			content += " " + strings.ToLower(item.Session.Region) + " " + strings.ToLower(item.Session.StartURL) + " " + strings.ToLower(strings.Join(item.Session.RegistrationScopes, ","))
		}
		if lowerInput == "" {
			return true
		}
		return strings.Contains(content, lowerInput)
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{if isNew .}}> {{ .Name | green }}{{else}}> {{ .Name | cyan }}   {{ sessionRegion .Session }}   {{ sessionStart .Session }}{{end}}",
		Inactive: "{{if isNew .}}  {{ .Name | faint }}{{else}}  {{ .Name | faint }}   {{ sessionRegion .Session }}   {{ sessionStart .Session }}{{end}}",
		Selected: "* {{ .Name }}",
		Details: `
--------- SSO Session ----------
Name:   {{ .Name }}
Region: {{ sessionRegion .Session }}
URL:    {{ sessionStart .Session }}
Scopes: {{ sessionScopes .Session }}`,
		FuncMap: buildPromptFuncMap(),
	}

	sel := promptui.Select{
		Label:             "Select or create SSO session (type to filter, Enter to choose)",
		Items:             choices,
		Searcher:          searcher,
		Templates:         templates,
		StartInSearchMode: true,
		Size:              10,
	}

	idx, _, err := sel.Run()
	if err != nil {
		return "", nil, err
	}

	chosen := choices[idx]
	if chosen.Name == addNewSessionLabel {
		defaultNewName := strings.TrimSpace(defaultName)
		if defaultNewName == "" {
			defaultNewName = lastSearchInput
		}
		newNamePrompt := promptui.Prompt{
			Label:     "Enter new SSO session name",
			Default:   defaultNewName,
			AllowEdit: true,
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("SSO session name cannot be empty")
				}
				if _, ok := cfg.SsoSession[input]; ok {
					return fmt.Errorf("%w: %s", errSessionExists, input)
				}
				return nil
			},
		}
		newName, err := newNamePrompt.Run()
		if err != nil {
			return "", nil, err
		}
		return strings.TrimSpace(newName), nil, nil
	}

	return chosen.Name, chosen.Session, nil
}

func buildPromptFuncMap() template.FuncMap {
	fm := template.FuncMap{}
	for k, v := range promptui.FuncMap {
		fm[k] = v
	}
	fm["isNew"] = func(opt sessionOption) bool {
		return opt.Session == nil && opt.Name == addNewSessionLabel
	}
	fm["sessionRegion"] = func(s *SsoSession) string {
		if s == nil {
			return ""
		}
		return s.Region
	}
	fm["sessionStart"] = func(s *SsoSession) string {
		if s == nil {
			return ""
		}
		return s.StartURL
	}
	fm["sessionScopes"] = func(s *SsoSession) string {
		if s == nil || len(s.RegistrationScopes) == 0 {
			return "-"
		}
		return strings.Join(s.RegistrationScopes, ",")
	}
	return fm
}

func createSsoSessionInSso(sessionName string, cfg *Configure) (*SsoSession, error) {
	newSession := &SsoSession{
		Name: sessionName,
	}

	if err := promptForRequiredStringWithDefault(&newSession.StartURL, "Please enter SSO start URL:", "SSO start URL", ""); err != nil {
		return nil, err
	}
	if err := promptForRequiredStringWithDefault(&newSession.Region, "Please enter SSO region:", "SSO region", defaultSsoRegion); err != nil {
		return nil, err
	}

	scopes, err := promptForRegistrationScopes(newSession.RegistrationScopes)
	if err != nil {
		return nil, err
	}
	newSession.RegistrationScopes = scopes

	cfg.SsoSession[sessionName] = newSession

	if err := WriteConfigToFile(cfg); err != nil {
		return nil, fmt.Errorf("failed to save SSO session configuration: %v", err)
	}

	return newSession, nil
}

func configureUsageTemplate() string {
	return `Usage:{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

func configureActionUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}} [params]{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}
