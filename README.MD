## Byteplus Command Line Tools

### Overview

1. Byteplus CLI Tool
2. Minimum Go version required: 1.5+, recommended: 1.12+

### Installing `bp`

#### Get the Client via Release

1. Download the latest version from: [https://github.com/byteplus-sdk/byteplus-cli/releases](https://github.com/byteplus/byteplus-cli/releases)
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

To call services, you need `AK`, `SK`, and `region`. You can configure them in two ways:

1. Use `bp configure set`:

```shell
bp configure set --profile test --region cn-beijing --access-key ak --secret-key sk --endpoint endpoint
```

Supported Parameters:

```shell
profile: Configuration name. If it exists, it will be updated; otherwise, a new one will be created.
access-key: Your AK
secret-key: Your SK
region: Region, e.g., cn-beijing
session-token: Required for role-based access
disable-ssl: Whether to disable SSL, default is false
```

2. Use Environment Variables:

```shell
export BYTEPLUS_ACCESS_KEY=AK
export BYTEPLUS_SECRET_KEY=SK
export BYTEPLUS_REGION=cn-beijing
export BYTEPLUS_DISABLE_SSL=false
export BYTEPLUS_SESSION_TOKEN=sessionToken
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

##### Delete Profile

```shell
bp configure delete --profile [profile_name]
```

If the deleted profile is the current one, `bp` will randomly pick another available profile.

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

#### Calling APIs

Basic structure:

```shell
bp <service> <action> [--param1 val1 --param2 val2 ...]
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
