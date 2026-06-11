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
# For macOS
sh build.sh darwin
# For Windows
sh build.sh windows
# For Linux
sh build.sh linux
```

### Configuring `bp`

Add `bp` to your system `PATH`

1. Check if `$PATH` includes `/usr/local/bin`. If not, add it accordingly.
2. Run the following command to copy `bp` to `/usr/local/bin`:

```shell
sudo cp bp /usr/local/bin
```

#### Configure Credentials

`bp` supports multiple credential modes and resolves credentials in the following order:

1. Active or runtime-selected profile: if a current profile exists, or `---profile` selects one at runtime, `bp` uses the credentials in that profile.
2. SDK default credential chain: if no profile is available, `bp` tries environment variables, OIDC environment variables, CLI profile, and ECS role credentials in order.
3. If all methods fail, `bp` returns an error.

Region can be configured in a profile, set through `BYTEPLUS_REGION`, or overridden at runtime with `---region`.

1. Use `bp configure set`:

```shell
bp configure set --profile prod --region ap-southeast-1 --access-key ak --secret-key sk
bp configure set --profile oidc-prod --mode oidc --region ap-southeast-1 --oidc-token-file /var/run/token --role-trn trn:iam::123456789012:role/example
bp configure set --profile ecs-prod --mode ecsrole --region ap-southeast-1 --role-name example-role
```

Supported Parameters:

```shell
profile: Configuration name. If it exists, it will be updated; otherwise, a new one will be created.
mode: Credential mode. Supported values are ak, sso, console-login, ramrolearn, oidc, ecsrole. Default is ak.
access-key: Your AK
secret-key: Your SK
region: Optional region, e.g., ap-southeast-1
session-token: Optional temporary session token, used with STS credentials
account-id: Account ID, required for ramrolearn mode
role-name: Role name, required for ramrolearn and ecsrole modes
oidc-token-file: OIDC token file path, required for oidc mode
role-trn: Role TRN, required for oidc mode
disable-ssl: Whether to disable SSL, default is false
endpoint: Optional custom endpoint. If omitted, the SDK resolves the endpoint automatically. Set to auto-addressing to use the standard endpoint resolver.
endpoint-resolver: Optional; set to standard (case-insensitive) to use the standard resolver. This takes precedence over endpoint.
use-dual-stack: Optional; enable dual-stack endpoints when true. The default value is false.
```

Supported profile modes:

| Mode | Description | Required profile fields |
|------|-------------|-------------------------|
| `ak` | Static AK/SK credentials (default) | `access-key`, `secret-key` |
| `sso` | SSO sign-in | Configured through `bp configure sso`; use `bp configure profile` to make it active when needed |
| `console-login` | Console Login with temporary STS credentials | Configured through `bp login`, which writes `login-session`; do not configure it manually |
| `ramrolearn` | AssumeRole via STS | `access-key`, `secret-key`, `role-name`, `account-id` |
| `oidc` | OIDC token exchange | `oidc-token-file`, `role-trn` |
| `ecsrole` | ECS instance role through IMDS | `role-name` |

When you configure a profile with `bp configure set`, all required fields for that mode must be written into the profile itself. Environment variables are used by the default credential chain when no active or runtime-selected profile is available.

Credential mode examples:

```shell
# AK/SK static credentials.
bp configure set --profile ak-prod --mode ak --region ap-southeast-1 --access-key AK --secret-key SK

# Assume role with AK/SK.
bp configure set --profile role-prod --mode ramrolearn --region ap-southeast-1 --access-key AK --secret-key SK --account-id 123456789012 --role-name example-role

# OIDC token file.
bp configure set --profile oidc-prod --mode oidc --region ap-southeast-1 --oidc-token-file /var/run/token --role-trn trn:iam::123456789012:role/example

# ECS role metadata credentials.
bp configure set --profile ecs-prod --mode ecsrole --region ap-southeast-1 --role-name example-role

# SSO and console-login are created by their dedicated flows.
bp configure sso --profile sso-prod --sso-session corp
bp login --profile console-prod --region ap-southeast-1
```

2. Use Environment Variables (default credential chain):

If no active profile is available, `bp` uses the SDK default credential chain. Environment variables are the first source in that chain:

```shell
export BYTEPLUS_ACCESS_KEY=AK
export BYTEPLUS_SECRET_KEY=SK
export BYTEPLUS_REGION=ap-southeast-1
export BYTEPLUS_DISABLE_SSL=false
export BYTEPLUS_SESSION_TOKEN=sessionToken
export BYTEPLUS_ENDPOINT=auto-addressing
export BYTEPLUS_ENDPOINT_RESOLVER=standard
export BYTEPLUS_USE_DUALSTACK=true
export BYTEPLUS_PROFILE=profileName
export BYTEPLUS_CLI_PROFILE=profileName
export BYTEPLUS_CLI_CONFIG_FILE=/path/to/config.json
export BYTEPLUS_OIDC_TOKEN_FILE=/var/run/token
export BYTEPLUS_OIDC_ROLE_TRN=trn:iam::123456789012:role/example
export BYTEPLUS_ECS_METADATA=roleName
export BYTEPLUS_ECS_METADATA_DISABLED=false
export BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS=false
```

When no active profile is available, `bp` uses the SDK default credential chain:

```shell
Environment variables -> OIDC environment variables -> CLI profile -> ECS role
```

OIDC credentials can also be supplied entirely through environment variables:

```shell
export BYTEPLUS_OIDC_TOKEN_FILE=/var/run/token
export BYTEPLUS_OIDC_ROLE_TRN=trn:iam::123456789012:role/example
export BYTEPLUS_REGION=ap-southeast-1
```

ECS role credentials are resolved from IMDS in ECS environments. Set `BYTEPLUS_ECS_METADATA` to request a specific role name, or set `BYTEPLUS_ECS_METADATA_DISABLED=true` to disable ECS role lookup.

Set `BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS=true` to disable this fallback and require an explicit profile.

Endpoint behavior:

```shell
# Use SDK standard endpoint resolution.
bp configure set --profile prod --endpoint auto-addressing
bp configure set --profile prod --endpoint-resolver standard

# Use a fixed endpoint.
bp configure set --profile prod --endpoint ecs.ap-southeast-1.byteplusapi.com
```

#### Other `configure` Commands

1. `bp` will prioritize using the credentials in the active profile.
2. `bp configure set` makes the target profile active after adding or modifying it.
3. `bp configure sso` creates or updates an SSO profile but does not switch the active profile automatically.
4. If multiple profiles exist, switch using:

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
bp configure set --profile [name] --mode ramrolearn --access-key [AK] --secret-key [SK] --account-id [account_id] --role-name [role_name]
bp configure set --profile [name] --mode oidc --oidc-token-file [token_file] --role-trn [role_trn]
bp configure set --profile [name] --mode ecsrole --role-name [role_name]
```

After creating or modifying a profile with `bp configure set`, the active profile switches to the target profile.

> Note: `region` is not mandatory during `configure set`, but it must be available at API call time through profile configuration, `---region`, or `BYTEPLUS_REGION`.

Additional Fields:

* access-key
* secret-key
* region
* session-token
* mode
* account-id
* role-name
* oidc-token-file
* role-trn
* login-session (console-login field, written by `bp login`)
* disable-ssl
* endpoint
* endpoint-resolver
* use-dual-stack

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
region: SSO region; required; default is ap-southeast-1, existing session values are used as defaults
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
- `bp configure sso` does not switch the active profile automatically; run `bp configure profile --profile [profile name]` if you want to use it as the default profile
- Reconfiguring an existing SSO profile clears old STS temporary credentials before writing the updated profile, so stale role credentials will not be reused

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

##### Runtime Override Parameters

When calling APIs, use the `---` prefix to pass CLI internal override parameters. These parameters do not conflict with API parameters:

- `---profile`: Use a specific profile at runtime without modifying the configuration file. It must be followed by an existing profile name.
- `---region`: Override region at runtime.

```shell
# Call an API with a specific profile.
bp ecs DescribeInstances ---profile prod

# Call an API with a specific profile and region override.
bp ecs DescribeInstances ---profile prod ---region ap-southeast-1

# `--` parameters are still sent to the API, while `---` parameters are consumed by the CLI.
bp ecs DescribeInstances ---profile prod ---region ap-southeast-1 --InstanceIds '["i-xxx"]'
```

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

## Security and privacy
This project takes security seriously. 
For vulnerability reporting and supported versions, see [SECURITY.md](SECURITY.md)
