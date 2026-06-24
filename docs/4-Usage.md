[Configuration](3-Configuration.md) | Usage | [Advanced Usage](5-Advanced.md)

---

## Usage

Basic command format:

```shell
bp <service> <action> [--Param value ...] [---profile name] [---region region] [---endpoint endpoint]
```

`--Param value` is an API parameter. `---profile`, `---region`, and `---endpoint` are CLI fixed flags.

## Discover Services and Actions

List supported services:

```shell
bp --help
```

List actions under a service:

```shell
bp ecs --help
```

Show action parameters:

```shell
bp ecs DescribeInstances --help
```

Show version:

```shell
bp version
bp -v
```

## Call APIs

Call without parameters:

```shell
bp sts GetCallerIdentity
```

Call with parameters:

```shell
bp ecs DescribeInstances --InstanceIds.1 i-1234567890abcdef0
```

Multiple parameters:

```shell
bp rds_mysql ListDBInstanceIPLists --InstanceId mysql-xxxxxx --GroupName default
```

Parameter names and values are separated by spaces. The supported syntax is:

```shell
--Param value
---region ap-southeast-1
```

Both `--Param value` and `--Param=value` are supported. Fixed flags also support `---region value` and `---region=value`.

## CLI Fixed Flags

Fixed flags use three hyphens `---` and do not conflict with API parameters:

| Flag | Purpose |
| --- | --- |
| `---profile` | Use a specific profile for this invocation without changing current |
| `---region` | Override region for this invocation |
| `---endpoint` | Override endpoint for this invocation and clear endpoint resolver |

Examples:

```shell
# Use a specific profile
bp ecs DescribeInstances ---profile prod

# Use a specific profile and override region
bp ecs DescribeInstances ---profile prod ---region ap-southeast-1

# Override only region
bp ecs DescribeInstances ---region ap-southeast-1

# Specify endpoint for an STS call
bp sts GetCallerIdentity ---region ap-southeast-1 ---endpoint sts.byteplusapi.com
```

If `---profile` references a profile that does not exist, the command returns an error.

## JSON Parameters

For query/form APIs, if a parameter value is a JSON object or JSON array, the CLI attempts to parse it as JSON:

```shell
bp rds_mysql ModifyDBInstanceIPList \
  --InstanceId mysql-xxxxxx \
  --GroupName default \
  --IPList '["10.20.30.40","50.60.70.80"]'
```

String parameters are kept as strings and are not forcibly parsed just because they look like JSON.

## application/json Requests

For APIs whose `ContentType` is `application/json`, pass a JSON body directly:

```shell
bp rds_mysql ModifyDBInstanceIPList \
  --body '{"InstanceId":"mysql-xxxxxx","GroupName":"default","IPList":["10.20.30.40","50.60.70.80"]}'
```

`--body` must be a JSON object or JSON array. It cannot be mixed with flattened parameters:

```shell
# Wrong: --body cannot be used together with other API parameters
bp rds_mysql ModifyDBInstanceIPList --body '{"InstanceId":"mysql-xxxxxx"}' --GroupName default
```

application/json APIs also support dotted keys. The CLI expands them into nested JSON using metadata:

```shell
bp some_service SomeJsonAction \
  --Name demo \
  --Ports.1 80 \
  --Ports.2 443 \
  --Tags.1.Key env \
  --Tags.1.Value prod
```

Array indices are 1-based and must be contiguous. `0`, negative indices, and skipped indices are errors.

## Arrays and Nested Parameters

Common array syntax:

```shell
bp ecs DescribeInstances --InstanceIds.1 i-123 --InstanceIds.2 i-456
```

Array of objects:

```shell
bp some_service SomeAction \
  --Filters.1.Key InstanceType \
  --Filters.1.Values.1 ecs.g1.large \
  --Filters.1.Values.2 ecs.g2.large
```

For application/json APIs, dotted keys are restored to nested objects and arrays. For non-JSON APIs, dotted keys are preserved and handled by the service/API layer.

## Unknown Parameters

The CLI allows unknown API parameters to pass through to the service/API layer. Unless the parameter path itself is invalid, the CLI does not reject a parameter only because it is absent from metadata.

Example:

```shell
bp ecs DescribeInstances --NewServerSideParam value
```

This is useful when the service has added a parameter but local metadata has not been updated yet.

## Common Scenarios

Use current profile:

```shell
bp ecs DescribeInstances
```

Use a non-current profile:

```shell
bp ecs DescribeInstances ---profile prod
```

Use environment-based default credential chain:

```shell
export BYTEPLUS_ACCESS_KEY=AK
export BYTEPLUS_SECRET_KEY=SK
export BYTEPLUS_REGION=ap-southeast-1
bp ecs DescribeInstances
```

Use an OIDC profile:

```shell
bp configure set --profile ci-oidc --mode oidc --region ap-southeast-1 \
  --oidc-token-file /var/run/secrets/oidc-token \
  --role-trn trn:iam::2000000000:role/CIRole

bp ecs DescribeInstances ---profile ci-oidc
```

Use an ECS instance role profile:

```shell
bp configure set --profile ecs-role --mode ecsrole --region ap-southeast-1 --role-name MyRole
bp ecs DescribeInstances ---profile ecs-role
```

## Common Errors

Missing credentials:

```text
credentials not configured, please run 'bp login' or 'bp configure set', or set BYTEPLUS_ACCESS_KEY and BYTEPLUS_SECRET_KEY environment variables
```

Missing region:

```text
region not set, please set it via profile, ---region flag, or BYTEPLUS_REGION environment variable
```

Unsupported fixed flag:

```text
---debug is not supported, supported fixed flags: ---profile, ---region, ---endpoint
```

The only supported fixed flags are `---profile`, `---region`, and `---endpoint`.

---

[Configuration](3-Configuration.md) | Usage | [Advanced Usage](5-Advanced.md)
