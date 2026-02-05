package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	ssoCmd := newSsoRootCmd()

	ssoCmd.AddCommand(newSsoLoginCmd())
	ssoCmd.AddCommand(newSsoLogoutCmd())

	rootCmd.AddCommand(ssoCmd)
}

func newSsoRootCmd() *cobra.Command {
	ssoCmd := &cobra.Command{
		Use:   "sso",
		Short: "Single sign-on (SSO) related operations",
		Long:  "Manage operations related to single sign-on (SSO), including login and logout",
	}

	ssoCmd.SetUsageTemplate(ssoUsageTemplate())

	return ssoCmd
}

func newSsoLoginCmd() *cobra.Command {
	ssoLoginCmd := &cobra.Command{
		Use:   "login",
		Short: "Perform SSO login operations",
		Long: `Login via SSO, obtain the access token and store it in the cache.
This command requires a configured profile, and the profile must be associated with a valid SSO session.
After a successful login, the system stores the access token for subsequent operations.`,
		Example: `  # Login to SSO using the specified profile
  bp sso login --profile my-sso-profile
  # Login to SSO using the specified sso-session
  bp sso login --sso-session my-sso-session`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ctx.config
			if cfg == nil {
				return fmt.Errorf("the configuration file cannot be loaded")
			}

			profileName := strings.TrimSpace(cmd.Flag("profile").Value.String())
			ssoSessionName := strings.TrimSpace(cmd.Flag("sso-session").Value.String())
			useDeviceCode := true
			noBrowser, err := cmd.Flags().GetBool("no-browser")
			if err != nil {
				return err
			}

			var sso *Sso
			var activeSessionName string

			if profileName != "" {
				profile, ok := cfg.Profiles[profileName]
				if !ok {
					return fmt.Errorf("the specified profile was not found: %s", profileName)
				}

				if strings.ToLower(strings.TrimSpace(profile.Mode)) != ModeSSO {
					return fmt.Errorf("the specified profile is not of sso type")
				}
				if strings.TrimSpace(profile.SsoSessionName) == "" {
					return fmt.Errorf("the specified profile does not have sso-session configured")
				}

				sso = &Sso{
					Profile:        profile,
					SsoSessionName: profile.SsoSessionName,
					Region:         profile.Region,
					UseDeviceCode:  useDeviceCode,
					NoBrowser:      noBrowser,
				}
				activeSessionName = profile.SsoSessionName
			} else if ssoSessionName != "" {
				ssoSession, ok := cfg.SsoSession[ssoSessionName]
				if !ok {
					return fmt.Errorf("the specified sso-session was not found: %s", ssoSessionName)
				}
				if ssoSession == nil {
					return fmt.Errorf("the specified sso-session is invalid: %s", ssoSessionName)
				}

				sso = &Sso{
					SsoSessionName: ssoSessionName,
					StartURL:       ssoSession.StartURL,
					Region:         ssoSession.Region,
					UseDeviceCode:  useDeviceCode,
					NoBrowser:      noBrowser,
				}
				activeSessionName = ssoSessionName
			} else {
				if len(cfg.SsoSession) == 0 {
					return fmt.Errorf("no sso-session configured")
				}
				if len(cfg.SsoSession) == 1 {
					for name, session := range cfg.SsoSession {
						if session == nil {
							return fmt.Errorf("the specified sso-session is invalid: %s", name)
						}
						sso = &Sso{
							SsoSessionName: name,
							StartURL:       session.StartURL,
							Region:         session.Region,
							UseDeviceCode:  useDeviceCode,
							NoBrowser:      noBrowser,
						}
						activeSessionName = name
						break
					}
				} else {
					options := buildSessionOptions(cfg.SsoSession)
					selectedName, selectedSession, err := selectExistingSession(options)
					if err != nil {
						return err
					}
					if selectedSession == nil {
						return fmt.Errorf("the specified sso-session is invalid: %s", selectedName)
					}
					sso = &Sso{
						SsoSessionName: selectedName,
						StartURL:       selectedSession.StartURL,
						Region:         selectedSession.Region,
						UseDeviceCode:  useDeviceCode,
						NoBrowser:      noBrowser,
					}
					activeSessionName = selectedName
				}
			}

			if err := sso.Login(); err != nil {
				if activeSessionName != "" {
					fmt.Printf("login failed for sso-session [%s]: %v\n", activeSessionName, err)
				}
				return err
			}

			if activeSessionName != "" {
				fmt.Printf("login successfully for sso-session [%s]\n", activeSessionName)
			} else {
				fmt.Println("login successfully")
			}
			return nil
		},
	}

	ssoLoginCmd.Flags().String("profile", "", "Specify the name of the configuration file to be used")
	ssoLoginCmd.Flags().String("sso-session", "", "Specify the SSO session to use when no profile is provided")
	ssoLoginCmd.Flags().Bool("no-browser", false, "Do not automatically open the browser during device authorization")

	ssoLoginCmd.SetUsageTemplate(ssoUsageTemplate())

	return ssoLoginCmd
}

func selectExistingSession(options []sessionOption) (string, *SsoSession, error) {
	if len(options) == 0 {
		return "", nil, fmt.Errorf("no sso-session configured")
	}

	searcher := func(input string, index int) bool {
		if index < 0 || index >= len(options) {
			return false
		}
		rawInput := strings.TrimSpace(input)
		lowerInput := strings.ToLower(rawInput)
		item := options[index]
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
		Active:   "> {{ .Name | cyan }}   {{ sessionRegion .Session }}   {{ sessionStart .Session }}",
		Inactive: "  {{ .Name | faint }}   {{ sessionRegion .Session }}   {{ sessionStart .Session }}",
		Selected: "[*] {{ .Name }}",
		Details: `
--------- SSO Session ----------
Name:   {{ .Name }}
Region: {{ sessionRegion .Session }}
URL:    {{ sessionStart .Session }}
Scopes: {{ sessionScopes .Session }}`,
		FuncMap: buildPromptFuncMap(),
	}

	sel := promptui.Select{
		Label:             "Select SSO session (type to filter, Enter to choose)",
		Items:             options,
		Searcher:          searcher,
		Templates:         templates,
		StartInSearchMode: true,
		Size:              10,
	}

	idx, _, err := sel.Run()
	if err != nil {
		return "", nil, err
	}

	chosen := options[idx]
	return chosen.Name, chosen.Session, nil
}

const allSessionsLabel = "All SSO sessions"

func selectSessionOrAll(options []sessionOption) (string, *SsoSession, bool, error) {
	if len(options) == 0 {
		return "", nil, false, fmt.Errorf("no sso-session configured")
	}

	choices := make([]sessionOption, 0, len(options)+1)
	choices = append(choices, options...)
	choices = append(choices, sessionOption{Name: allSessionsLabel, Session: nil})

	searcher := func(input string, index int) bool {
		if index < 0 || index >= len(choices) {
			return false
		}
		rawInput := strings.TrimSpace(input)
		lowerInput := strings.ToLower(rawInput)
		item := choices[index]
		if item.Name == allSessionsLabel {
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
		Active:   "{{if isAll .}}> {{ .Name | yellow }}{{else}}> {{ .Name | cyan }}   {{ sessionRegion .Session }}   {{ sessionStart .Session }}{{end}}",
		Inactive: "{{if isAll .}}  {{ .Name | faint }}{{else}}  {{ .Name | faint }}   {{ sessionRegion .Session }}   {{ sessionStart .Session }}{{end}}",
		Selected: "[*] {{ .Name }}",
		Details: `
--------- SSO Session ----------
Name:   {{ .Name }}
Region: {{ sessionRegion .Session }}
URL:    {{ sessionStart .Session }}
Scopes: {{ sessionScopes .Session }}`,
		FuncMap: func() map[string]interface{} {
			fm := buildPromptFuncMap()
			fm["isAll"] = func(opt sessionOption) bool {
				return opt.Name == allSessionsLabel
			}
			return fm
		}(),
	}

	sel := promptui.Select{
		Label:             "Select SSO session to logout (type to filter, Enter to choose)",
		Items:             choices,
		Searcher:          searcher,
		Templates:         templates,
		StartInSearchMode: true,
		Size:              10,
	}

	idx, _, err := sel.Run()
	if err != nil {
		return "", nil, false, err
	}

	chosen := choices[idx]
	if chosen.Name == allSessionsLabel {
		return "", nil, true, nil
	}
	return chosen.Name, chosen.Session, false, nil
}

func logoutAllSessions(cfg *Configure) error {
	if cfg == nil {
		return fmt.Errorf("the configuration file cannot be loaded")
	}

	sessionNames := make([]string, 0, len(cfg.SsoSession))
	for name := range cfg.SsoSession {
		sessionNames = append(sessionNames, name)
	}
	sort.Strings(sessionNames)

	var failures []string
	for _, name := range sessionNames {
		session := cfg.SsoSession[name]
		if session == nil {
			continue
		}
		sso := &Sso{
			SsoSessionName: name,
			StartURL:       session.StartURL,
			Region:         session.Region,
		}
		if err := sso.Logout(); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", name, err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("failed to logout some sso sessions: %s", strings.Join(failures, "; "))
	}

	return nil
}

func newSsoLogoutCmd() *cobra.Command {
	ssoLogoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Perform SSO logout operations",
		Long:  `Logout from SSO by revoking the cached token and clearing local credentials.`,
		Example: `  # Logout SSO by profile
  bp sso logout --profile my-sso-profile
  # Logout SSO by sso-session
  bp sso logout --sso-session my-sso-session`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ctx.config
			if cfg == nil {
				return fmt.Errorf("the configuration file cannot be loaded")
			}

			ssoSessionName := strings.TrimSpace(cmd.Flag("sso-session").Value.String())

			if ssoSessionName != "" {
				session, ok := cfg.SsoSession[ssoSessionName]
				if !ok {
					return fmt.Errorf("the specified sso-session was not found: %s", ssoSessionName)
				}
				sso := &Sso{
					SsoSessionName: ssoSessionName,
					StartURL:       session.StartURL,
					Region:         session.Region,
				}
				if err := sso.Logout(); err != nil {
					return err
				}
				fmt.Println("logout successfully")
				return nil
			}

			if len(cfg.SsoSession) == 0 {
				return fmt.Errorf("no sso-session configured")
			}
			if len(cfg.SsoSession) == 1 {
				for name, session := range cfg.SsoSession {
					if session == nil {
						return fmt.Errorf("the specified sso-session is invalid: %s", name)
					}
					sso := &Sso{
						SsoSessionName: name,
						StartURL:       session.StartURL,
						Region:         session.Region,
					}
					if err := sso.Logout(); err != nil {
						return err
					}
					fmt.Println("logout successfully")
					return nil
				}
			}

			options := buildSessionOptions(cfg.SsoSession)
			selectedName, selectedSession, logoutAll, err := selectSessionOrAll(options)
			if err != nil {
				return err
			}
			if logoutAll {
				if err := logoutAllSessions(cfg); err != nil {
					return err
				}
				fmt.Println("logout successfully")
				return nil
			}
			if selectedSession == nil {
				return fmt.Errorf("the specified sso-session is invalid: %s", selectedName)
			}

			sso := &Sso{
				SsoSessionName: selectedName,
				StartURL:       selectedSession.StartURL,
				Region:         selectedSession.Region,
			}
			if err := sso.Logout(); err != nil {
				return err
			}
			fmt.Println("logout successfully")
			return nil
		},
	}

	ssoLogoutCmd.Flags().String("sso-session", "", "Specify the SSO session to log out")

	ssoLogoutCmd.SetUsageTemplate(ssoUsageTemplate())

	return ssoLogoutCmd
}

func ssoUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

