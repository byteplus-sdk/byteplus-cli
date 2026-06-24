[Authentication](2-Authentication.md) | Configuration | [Usage](4-Usage.md)

---

## Configuration

CLI profiles and SSO sessions are stored in `~/.byteplus/config.json` by default. The config file is written with `0600` permissions, and the config directory is created with `0700` permissions.

This document covers profile inspection, switching, updates, and deletion. Credential modes are covered in [Authentication](2-Authentication.md).

## Config File Structure

The config file contains:

- `current`: current default profile name.
- `profiles`: profile map.
- `sso-session`: SSO session map.
- `enableColor`: whether colored JSON output is enabled. See [Advanced Usage](5-Advanced.md).

Example:

```json
{
    "current": "prod",
    "profiles": {
        "prod": {
            "name": "prod",
            "mode": "ak",
            "access-key": "AK",
            "secret-key": "SK",
            "region": "ap-southeast-1"
        }
    },
    "enableColor": false,
    "sso-session": {}
}
```

Avoid manually editing sensitive fields. Prefer CLI commands.

## Show Current Profile

```shell
bp configure get
```

Without `--profile`, the command shows the current profile:

```shell
no profile name specified, show current profile: [prod]
```

## Show a Specific Profile

```shell
bp configure get --profile prod
```

If the profile does not exist, the command prints an empty profile object and does not create it.

## List All Profiles

```shell
bp configure list
```

The output starts with current:

```shell
*** current profile: prod ***
```

Then each profile in the config file is printed.

## Switch Current Profile

```shell
bp configure profile --profile prod
```

`--profile` is required. If the profile does not exist, current is not changed and an error is returned.

Switching current affects later service commands that do not specify `---profile`. For a single invocation, use:

```shell
bp ecs DescribeInstances ---profile prod
```

## Create or Update a Profile

```shell
bp configure set --profile prod --region ap-southeast-1 --access-key AK --secret-key SK
```

Behavior:

- `--profile` is required.
- If the profile does not exist, it is created with default mode `ak`.
- If the profile exists, only non-empty fields provided in this command are updated; omitted fields keep their previous values.
- `--disable-ssl` and `--use-dual-stack` are written only when explicitly provided.
- Successful create or update switches current to that profile.
- `region` is not mandatory during `configure set`, but API calls must be able to resolve a region.

Update region:

```shell
bp configure set --profile prod --region ap-southeast-1
```

Update endpoint:

```shell
bp configure set --profile prod --endpoint ecs.ap-southeast-1.byteplusapi.com
```

Use the standard endpoint resolver:

```shell
bp configure set --profile prod --endpoint-resolver standard
```

Configure proxy:

```shell
bp configure set --profile prod --https-proxy http://127.0.0.1:7890
```

Enable dual-stack:

```shell
bp configure set --profile prod --use-dual-stack
```

Disable SSL:

```shell
bp configure set --profile prod --disable-ssl
```

## Delete a Profile

```shell
bp configure delete --profile prod
```

`--profile` is required. If the deleted profile is current, the CLI selects one remaining profile as the new current. If no profiles remain, current becomes empty.

Deleting a profile does not delete SSO sessions or the global Console Login cache directory. Console Login cache cleanup is covered in [Authentication](2-Authentication.md#console-logout).

## Selection Examples

### Switch Between Environments

```shell
bp configure set --profile dev --region ap-southeast-1 --access-key DEV_AK --secret-key DEV_SK
bp configure set --profile prod --region ap-southeast-1 --access-key PROD_AK --secret-key PROD_SK

bp configure profile --profile dev
bp ecs DescribeInstances

bp configure profile --profile prod
bp ecs DescribeInstances
```

### Override Profile for One Call

```shell
bp configure profile --profile dev
bp ecs DescribeInstances ---profile prod
```

This call uses `prod` only for this invocation and does not modify `current`.

### Override Region and Endpoint for One Call

```shell
bp ecs DescribeInstances ---region ap-southeast-1
bp sts GetCallerIdentity ---region ap-southeast-1 ---endpoint sts.byteplusapi.com
```

---

[Authentication](2-Authentication.md) | Configuration | [Usage](4-Usage.md)
