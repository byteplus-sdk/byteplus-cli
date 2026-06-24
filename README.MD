## Byteplus Command Line Tools

### Overview

1. Byteplus CLI Tool
2. Minimum Go version required: 1.5+, recommended: 1.12+

### Installing `bp`

#### Get the Client via Release

1. Download the latest version from: [https://github.com/byteplus-sdk/byteplus-cli/releases](https://github.com/byteplus-sdk/byteplus-cli/releases)
2. Select the version for your OS, extract and use directly

#### Build the Client Yourself

1. Use `build.sh` to compile the client

```shell
# Build for the current OS/arch
sh build.sh

# Cross compile for a specific target
sh build.sh darwin arm64
sh build.sh windows amd64
sh build.sh linux amd64
```

#### Install from npm

```shell
npm install -g @byteplus/cli
```

The npm package downloads the matching `bp` release asset during installation.
Set `BYTEPLUS_CLI_DOWNLOAD_BASE_URL` when you need to use an internal mirror.

### Configuring `bp`

Add `bp` to your system `PATH`

1. Check if `$PATH` includes `/usr/local/bin`. If not, add it accordingly.
2. Run the following command to copy `bp` to `/usr/local/bin`:

```shell
sudo cp bp /usr/local/bin
```

#### Configure Credentials

To call services, you need `AK`, `SK`, and `region`. You can configure them in two ways:

1. Use `bp configure set`:

```shell
bp configure set --profile test --region ap-southeast-1 --access-key ak --secret-key sk --endpoint endpoint
```

Supported Parameters:

```shell
profile: Configuration name. If it exists, it will be updated; otherwise, a new one will be created.
mode: Credential mode. Supported values: ak, sso, console-login
access-key: Your AK
secret-key: Your SK
region: Region, e.g., ap-southeast-1
session-token: Required for role-based access
disable-ssl: Whether to disable SSL, default is false
endpoint: Optional; custom service endpoint. Ignored when endpoint-resolver is standard.
endpoint-resolver: Optional; set to standard (case-insensitive) to use the standard resolver. Otherwise, the CLI uses endpoint when provided.
use-dual-stack: Optional; enable dual-stack endpoints when true. The default value is false.
http-proxy: Optional; HTTP proxy URL used by SDK requests.
https-proxy: Optional; HTTPS proxy URL used by SDK requests.
```

2. Use Environment Variables:

```shell
export BYTEPLUS_ACCESS_KEY=AK
export BYTEPLUS_SECRET_KEY=SK
export BYTEPLUS_REGION=cn-beijing
export BYTEPLUS_DISABLE_SSL=false
export BYTEPLUS_SESSION_TOKEN=sessionToken
export BYTEPLUS_ENDPOINT=endpoint
export BYTEPLUS_ENDPOINT_RESOLVER=standard
export BYTEPLUS_USE_DUALSTACK=true
export BYTEPLUS_PROFILE=default
export BYTEPLUS_CLI_PROFILE=default
```

#### Other `configure` Commands

1. `bp` will prioritize using the credentials in the active profile.
2. After adding or modifying a profile, it becomes the active profile.
3. If multiple profiles exist, switch using:

```shell
bp configure profile --profile [profile_name]
```

##### View a Specific Profile

```shell
bp configure get --profile [profile_name]
```

If omitted, current profile is shown. If profile doesn't exist, default values are returned.

##### List All Profiles

```shell
bp configure list
```

##### Switch Active Profile (>= v1.0.16)

```shell
bp configure profile --profile [profile_name]
```

If profile doesn't exist, the switch fails with an error message.

##### Add/Modify Profile

```shell
bp configure set --profile [name] --region [region] --access-key [AK] --secret-key [SK] --endpoint [endpoint]
```

Additional Fields:

* access-key
* secret-key
* region
* session-token
* disable-ssl
* endpoint
* endpoint-resolver
* use-dual-stack
* http-proxy
* https-proxy
* mode

##### Delete Profile

```shell
bp configure delete --profile [profile_name]
```

If the deleted profile is the current one, `bp` will randomly pick another available profile.

#### SSO Session Management

###### Configure SSO Session (configure sso-session)

```shell
bp configure sso-session --name [session name] --start-url [SSO Start URL] --region [region] --registration-scopes [scope1,scope2]
```

This command creates or updates an SSO session for later use with `configure sso`. Behaviors and parameters:

```shell
name: SSO session name; if omitted, enter interactive selection/creation mode
start-url: SSO Start URL; required; if editing an existing session, Enter keeps the default
region: SSO region; required; default is cn-beijing, existing session values are used as defaults
registration-scopes: SSO scope list (comma-separated); allowed values are cloudidentity:account:access, offline_access
```

Interactive flow notes:
- If no sessions exist, you will be prompted to enter a non-empty session name
- If sessions exist, you get a searchable list with "<Create new session>" to add one
- When editing, Start URL, Region, and Scopes are pre-filled; press Enter to keep defaults
- Scopes are deduplicated and validated; empty input falls back to cloudidentity:account:access, offline_access

###### Configure SSO Profile (configure sso)

```shell
bp configure sso --profile [profile name] --sso-session [session name]
```

This command configures an SSO profile, links an SSO session, runs the device authorization flow, and writes the profile to ~/.byteplus/config.json. Behaviors and parameters:

```shell
profile: profile name; empty uses {sso-role-name}-{sso-account-id} as the default
sso-session: SSO session name; if omitted, enter interactive selection/creation mode
no-browser: Adding the `--no-browser` parameter to the command line disables the browser from opening; omitting it will automatically open the browser by default.
```

Notes:
- If the profile already exists and is not an SSO profile, the command will refuse to overwrite it
- If the SSO session does not exist, you will be guided to create it and enter Start URL, Region, and Scopes

#### SSO Login and Logout

##### SSO Login (sso login)

```shell
bp sso login --profile [profile name]
# or
bp sso login --sso-session [session name]
```

This command performs an explicit SSO login, runs a new authorization flow, obtains an access token, and caches it. Each execution re-authenticates the current SSO session and does not silently exchange an existing refresh_token for an access_token. Parameters:

```shell
profile: the SSO profile to use; must exist, be of sso type, and have sso-session configured
sso-session: the SSO session to use; the session must exist and be valid
no-browser: Adding the `--no-browser` parameter to the command line disables the browser from opening; omitting it will automatically open the browser by default.
```

Login behavior:
- If neither profile nor sso-session is provided: error when no sessions are configured; use the only session if one exists; otherwise enter interactive selection with search
- Only the device code flow is supported; use `--no-browser` to disable auto-opening the browser
- When running business commands, if the cached SSO access token is expired or close to expiry, the CLI attempts to silently refresh the access token with the cached refresh_token before requesting role credentials
- If the refresh_token is missing, the client registration has expired, or refresh fails, business commands do not automatically open a browser for re-authorization and will ask you to run `bp sso login` again

##### SSO Logout (sso logout)

```shell
bp sso logout --sso-session [session name]
```

This command logs out from SSO by revoking the cached token and clearing local credentials. Parameters:

```shell
sso-session: the SSO session name to log out
```

Logout behavior:
- If sso-session is not provided: error when no sessions are configured; logout the only session if one exists; otherwise enter interactive selection that includes "All SSO sessions"
- Batch logout logs out each session and returns aggregated errors on failure

#### SSO Login and Logout

##### SSO Login (sso login)

```shell
bp sso login --profile [profile name]
# or
bp sso login --sso-session [session name]
```

This command logs in via SSO, obtains an access token, and caches it. Parameters:

```shell
profile: the SSO profile to use; must exist, be of sso type, and have sso-session configured
sso-session: the SSO session to use; the session must exist and be valid
no-browser: Adding the `--no-browser` parameter to the command line disables the browser from opening; omitting it will automatically open the browser by default.
```

Login behavior:
- If neither profile nor sso-session is provided: error when no sessions are configured; use the only session if one exists; otherwise enter interactive selection with search
- Only the device code flow is supported; use `--no-browser` to disable auto-opening the browser

##### SSO Logout (sso logout)

```shell
bp sso logout --sso-session [session name]
```

This command logs out from SSO by revoking the cached token and clearing local credentials. Parameters:

```shell
sso-session: the SSO session name to log out
```

Logout behavior:
- If sso-session is not provided: error when no sessions are configured; logout the only session if one exists; otherwise enter interactive selection that includes "All SSO sessions"
- Batch logout logs out each session and returns aggregated errors on failure

#### Console Login (login)

```shell
# Login with default profile; you will be prompted for region if not specified
bp login

# Recommended: specify profile and region; supports -p/-r shortcuts
bp login -p dev -r ap-southeast-1

# Cross-device login for headless/remote servers or containers
bp login -p dev -r ap-southeast-1 --remote
```

This command performs OAuth 2.0 + PKCE login through the Byteplus console and caches temporary STS credentials locally. Subsequent CLI commands can reuse the profile's temporary credentials.

Parameters:
- `--profile, -p`: Profile name, default `default`
- `--region, -r`: Region; if not specified, you will be prompted; press Enter to use `ap-southeast-1`
- `--remote`: Cross-device login; follow the URL shown in the terminal to complete login in a browser, then paste the authorization code back into the terminal
- `--endpoint-url`: Login service URL, default `https://signin.byteplus.com`; usually does not need to be changed

Notes:
- After successful login, the profile will be set to `console-login` mode and record a `login-session`
- After logging in with a non-`default` profile, CLI commands will not automatically switch profiles; run `bp configure profile --profile dev` first
- Use `--remote` when the local browser cannot redirect back to the CLI machine

Full usage example:

```shell
# 1. Login and write to the dev profile
bp login --profile dev --region ap-southeast-1

# 2. If dev is not the current default profile, switch first
bp configure profile --profile dev

# 3. Run CLI commands using the console-login temporary credentials
bp ecs DescribeInstances

# 4. Clean up the local login state when done
bp logout --profile dev
```

#### Console Logout (logout)

```shell
# Logout the default profile
bp logout

# Logout a specific profile
bp logout -p dev

# Logout all Console Login profiles in the current configuration
bp logout --all
```

This command cleans up the local login state created by `bp login`. It only deletes local cached credentials and clears the `login-session` from the profile; it does not delete the profile or make any server-side requests.

Parameters:
- `--profile, -p`: Profile name, default `default`
- `--all`: Clean up all `console-login` profiles; when enabled, `--profile` is ignored

Notes:
- `bp logout` without `--profile` only processes the `default` profile; it does not automatically log out based on `current`
- Only `console-login` profiles can use this command; AK/SSO profiles are not applicable
- After logout, run `bp login` again to re-authenticate

#### Configure Auto-Completion

Run `bp completion --help` to check how to enable shell auto-completion.

##### Bash

1. View bash script: `bp completion bash`
2. Install bash-completion:

```shell
yum install bash-completion
# or
apt-get install bash-completion
```

3. Enable it:

```shell
source /usr/share/bash-completion/bash_completion
```

Add it to `~/.bashrc` for persistence.

4. Configure auto-completion:

```shell
echo 'source <(bp completion bash)' >> ~/.bashrc
bp completion bash > /etc/bash_completion.d/bp
```

Then reload shell or run `source ~/.bashrc`

##### Zsh

1. Enable `compinit`:

```shell
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

2. Configure:

```shell
bp completion zsh > "${fpath[1]}/_bp"
```

Then reload shell or run `source ~/.zshrc`

#### Configure Color Output

Enable color:

```shell
bp enable-color
```

Disable color:

```shell
bp disable-color
```

### Using `bp`

#### List Services and Actions

1. View available services:

```shell
bp [-h|--help]
```

2. View service's supported actions:

```shell
bp ecs [-h|--help]
```

#### Show CLI version

```shell
bp version
```

```shell
bp -v
```

#### Calling APIs

Basic structure:

```shell
bp <service> <action> [--param1 val1 --param2 val2 ...]
```

Fixed flags use three hyphens and only affect the current invocation:

```shell
bp sts GetCallerIdentity ---profile dev ---region ap-southeast-1
bp sts GetCallerIdentity ---region ap-southeast-1 ---endpoint sts.byteplusapi.com
```

Supported fixed flags:

- `---profile`: use a configured profile only for this invocation.
- `---region`: override the region only for this invocation.
- `---endpoint`: override the endpoint only for this invocation.

You can enable CLI debug logs with an environment variable:

```shell
BYTEPLUS_CLI_DEBUG=true bp ecs DescribeInstances
```

`BYTEPLUS_CLI_DEBUG` disables debug when set to `false`, `0`, `off`, `no`, or an empty value. Any other non-empty value enables debug. When enabled, logs are appended to the hourly log file under the configuration directory: `~/.byteplus/logs/YYYYMMDDHH.log`. When debug is disabled, no logs are written and no log file is created.

Example:

```shell
bp ecs DescribeInstances
bp rds_mysql ListDBInstanceIPLists --InstanceId "xxxxxx"
```

Support JSON input:

```shell
bp rds_mysql ModifyDBInstanceIPList --InstanceId "xxxxxx" --GroupName "xxxxxx" --IPList '["10.20.30.40", "50.60.70.80"]'
```

Support `--body` for application/json:

```shell
bp rds_mysql ModifyDBInstanceIPList --body '{"InstanceId":"xxxxxx", "GroupName": "xxxxxx", "IPList": ["10.20.30.40", "50.60.70.80"]}'
```

For application/json APIs, you can also pass flattened parameters instead of
`--body`. Array indexes are 1-based:

```shell
bp example_service ExampleAction --Items.1.Name first --Items.1.Count 3
```

## Security and privacy
This project takes security seriously. 
For vulnerability reporting and supported versions, see [SECURITY.md](SECURITY.md)
