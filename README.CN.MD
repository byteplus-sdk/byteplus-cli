[English](README.MD) | 中文

## Byteplus 命令行工具

### 概览

1. Byteplus CLI 工具
2. 最低 Go 版本要求：1.5+，推荐：1.12+

### 安装 `bp`

#### 通过 Release 获取客户端

1. 从以下地址下载最新版本：[https://github.com/byteplus-sdk/byteplus-cli/releases](https://github.com/byteplus-sdk/byteplus-cli/releases)
2. 选择适合当前操作系统的版本，解压后即可使用

#### 自行构建客户端

1. 使用 `build.sh` 编译客户端

```shell
# macOS
sh build.sh darwin
# Windows
sh build.sh windows
# Linux
sh build.sh linux
```

### 配置 `bp`

将 `bp` 添加到系统 `PATH`

1. 检查 `$PATH` 是否包含 `/usr/local/bin`。如果不包含，请按需添加。
2. 执行以下命令将 `bp` 复制到 `/usr/local/bin`：

```shell
sudo cp bp /usr/local/bin
```

#### 配置凭证

`bp` 支持多种凭证模式，并按以下顺序解析凭证：

1. 当前激活或运行时指定的 profile：如果存在当前 profile，或通过 `---profile` 在运行时选择了 profile，`bp` 会使用该 profile 中的凭证。
2. SDK 默认凭证链：如果没有可用 profile，`bp` 会依次尝试环境变量、OIDC 环境变量、CLI profile 和 ECS Role 凭证。
3. 如果所有方式都失败，`bp` 会返回错误。

Region 可以配置在 profile 中，也可以通过 `BYTEPLUS_REGION` 设置，或在运行时通过 `---region` 覆盖。

1. 使用 `bp configure set`：

```shell
bp configure set --profile prod --region ap-southeast-1 --access-key ak --secret-key sk
bp configure set --profile oidc-prod --mode oidc --region ap-southeast-1 --oidc-token-file /var/run/token --role-trn trn:iam::123456789012:role/example
bp configure set --profile ecs-prod --mode ecsrole --region ap-southeast-1 --role-name example-role
```

支持的参数：

```shell
profile: 配置名称。如果已存在则更新，否则新建。
mode: 凭证模式。支持 ak、sso、console-login、ramrolearn、oidc、ecsrole。默认值为 ak。
access-key: 你的 AK
secret-key: 你的 SK
region: 可选 Region，例如 ap-southeast-1
session-token: 可选临时 session token，用于 STS 凭证
account-id: 账号 ID，ramrolearn 模式必填
role-name: 角色名称，ramrolearn 和 ecsrole 模式必填
oidc-token-file: OIDC token 文件路径，oidc 模式必填
role-trn: 角色 TRN，oidc 模式必填
disable-ssl: 是否禁用 SSL，默认值为 false
endpoint: 可选自定义 endpoint。如果省略，SDK 会自动解析 endpoint。设置为 auto-addressing 可使用标准 endpoint 解析器。
endpoint-resolver: 可选；设置为 standard（大小写不敏感）可使用标准解析器。该参数优先于 endpoint。
use-dual-stack: 可选；为 true 时启用双栈 endpoint。默认值为 false。
```

支持的 profile 模式：

| 模式 | 说明 | 必填 profile 字段 |
|------|------|-------------------|
| `ak` | 静态 AK/SK 凭证（默认） | `access-key`、`secret-key` |
| `sso` | SSO 登录 | 通过 `bp configure sso` 配置；需要作为当前 profile 使用时，通过 `bp configure profile` 激活 |
| `console-login` | Console Login 临时 STS 凭证 | 通过 `bp login` 配置，该命令会写入 `login-session`；请勿手动配置 |
| `ramrolearn` | 通过 STS AssumeRole | `access-key`、`secret-key`、`role-name`、`account-id` |
| `oidc` | OIDC token 交换 | `oidc-token-file`、`role-trn` |
| `ecsrole` | 通过 IMDS 获取 ECS 实例角色 | `role-name` |

使用 `bp configure set` 配置 profile 时，该模式所需的所有字段都必须写入 profile 自身。环境变量仅在没有当前激活或运行时指定的 profile 时，由默认凭证链使用。

凭证模式示例：

```shell
# AK/SK 静态凭证。
bp configure set --profile ak-prod --mode ak --region ap-southeast-1 --access-key AK --secret-key SK

# 使用 AK/SK Assume Role。
bp configure set --profile role-prod --mode ramrolearn --region ap-southeast-1 --access-key AK --secret-key SK --account-id 123456789012 --role-name example-role

# OIDC token 文件。
bp configure set --profile oidc-prod --mode oidc --region ap-southeast-1 --oidc-token-file /var/run/token --role-trn trn:iam::123456789012:role/example

# ECS role 元数据凭证。
bp configure set --profile ecs-prod --mode ecsrole --region ap-southeast-1 --role-name example-role

# SSO 和 console-login 由各自的专用流程创建。
bp configure sso --profile sso-prod --sso-session corp
bp login --profile console-prod --region ap-southeast-1
```

2. 使用环境变量（默认凭证链）：

如果没有可用的当前 profile，`bp` 会使用 SDK 默认凭证链。环境变量是该链中的第一个来源：

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

当没有可用的当前 profile 时，`bp` 会使用 SDK 默认凭证链：

```shell
环境变量 -> OIDC 环境变量 -> CLI profile -> ECS role
```

OIDC 凭证也可以完全通过环境变量提供：

```shell
export BYTEPLUS_OIDC_TOKEN_FILE=/var/run/token
export BYTEPLUS_OIDC_ROLE_TRN=trn:iam::123456789012:role/example
export BYTEPLUS_REGION=ap-southeast-1
```

ECS role 凭证会在 ECS 环境中通过 IMDS 解析。设置 `BYTEPLUS_ECS_METADATA` 可请求指定角色名称，或设置 `BYTEPLUS_ECS_METADATA_DISABLED=true` 禁用 ECS role 查找。

设置 `BYTEPLUS_DISABLE_DEFAULT_CREDENTIALS=true` 可禁用该兜底逻辑，并要求显式 profile。

Endpoint 行为：

```shell
# 使用 SDK 标准 endpoint 解析。
bp configure set --profile prod --endpoint auto-addressing
bp configure set --profile prod --endpoint-resolver standard

# 使用固定 endpoint。
bp configure set --profile prod --endpoint ecs.ap-southeast-1.byteplusapi.com
```

#### 其他 `configure` 命令

1. `bp` 会优先使用当前激活 profile 中的凭证。
2. `bp configure set` 在新增或修改目标 profile 后，会将其设为当前激活 profile。
3. `bp configure sso` 会创建或更新 SSO profile，但不会自动切换当前激活 profile。
4. 如果存在多个 profile，使用以下命令切换：

```shell
bp configure profile --profile [profile_name]
```

##### 查看指定 Profile

```shell
bp configure get --profile [profile_name]
```

如果省略参数，则展示当前 profile。如果 profile 不存在，则返回默认值。

##### 列出所有 Profile

```shell
bp configure list
```

##### 切换当前激活 Profile（>= v1.0.16）

```shell
bp configure profile --profile [profile_name]
```

如果 profile 不存在，切换会失败并返回错误信息。

##### 新增/修改 Profile

```shell
bp configure set --profile [name] --region [region] --access-key [AK] --secret-key [SK] --endpoint [endpoint]
bp configure set --profile [name] --mode ramrolearn --access-key [AK] --secret-key [SK] --account-id [account_id] --role-name [role_name]
bp configure set --profile [name] --mode oidc --oidc-token-file [token_file] --role-trn [role_trn]
bp configure set --profile [name] --mode ecsrole --role-name [role_name]
```

使用 `bp configure set` 创建或修改 profile 后，当前激活 profile 会切换到目标 profile。

> 注意：`configure set` 时 `region` 不是必填项，但调用 API 时必须能通过 profile 配置、`---region` 或 `BYTEPLUS_REGION` 获得 region。

附加字段：

* access-key
* secret-key
* region
* session-token
* mode
* account-id
* role-name
* oidc-token-file
* role-trn
* login-session（console-login 字段，由 `bp login` 写入）
* disable-ssl
* endpoint
* endpoint-resolver
* use-dual-stack

##### 删除 Profile

```shell
bp configure delete --profile [profile_name]
```

如果删除的是当前 profile，`bp` 会随机选择另一个可用 profile。

#### SSO 快速开始

SSO 配置分为两层：`sso-session` 保存企业 SSO 入口（Start URL、Region、Scopes），SSO profile 保存 API 调用所选的账号和角色。首次配置时，按以下顺序执行：

```shell
# 1. 创建 SSO session。registration-scopes 可省略以使用默认值。
bp configure sso-session --name my-sso --start-url https://{custom-sso-domain}/userportal --region ap-southeast-1

# 2. 创建 SSO profile，完成设备授权，并选择账号和角色。
bp configure sso --profile my-dev --sso-session my-sso

# 3. 如果希望业务命令默认使用该 profile，请切换当前激活 profile。
bp configure profile --profile my-dev

# 4. 运行业务命令。CLI 会按需刷新临时 STS 凭证。
bp [service] [action] [params]
```

`bp configure sso` 会写入 SSO profile，但不会自动切换当前激活 profile。如果跳过第 3 步，业务命令仍会使用之前的 current profile。

#### SSO 命令关系

| 命令 | 适用场景 | 主要效果 | 是否切换当前激活 profile |
| --- | --- | --- | --- |
| `bp configure sso-session` | 通常每个 SSO 入口配置一次 | 保存 Start URL、Region 和 Scopes，供 SSO profile 复用 | 否 |
| `bp configure sso` | 每组账号和角色组合配置一次 | 关联 SSO session，执行首次授权，选择账号和角色，并写入 SSO profile | 否 |
| `bp configure profile --profile [profile name]` | 希望业务命令默认使用某个 profile 时 | 切换当前激活 profile | 是 |
| `bp sso login` | 提示需要重新登录时，或显式刷新 SSO 登录状态时 | 重新执行设备授权并缓存新的 access token | 否 |
| `bp sso logout` | 退出一个或全部 SSO session 时 | 撤销缓存 token，删除 token 缓存，并清理临时 STS 凭证 | 否 |

#### SSO Session 管理

###### 配置 SSO Session（configure sso-session）

```shell
bp configure sso-session --name [session name] --start-url [SSO Start URL] --region [region] --registration-scopes [scope1,scope2]
```

该命令会创建或更新一个 SSO session，供后续 `configure sso` 使用。行为和参数：

```shell
name: SSO session 名称；如果省略，会进入交互式选择/创建模式
start-url: SSO Start URL；必填；编辑现有 session 时按 Enter 保留默认值
region: SSO region；必填；默认值为 ap-southeast-1，已有 session 的值会作为默认值
registration-scopes: SSO scope 列表（逗号分隔）；允许值为 cloudidentity:account:access、offline_access
```

交互流程说明：
- 如果没有任何 session，会提示输入非空 session 名称
- 如果已有 session，会展示可搜索列表，并包含 "<Create new session>" 用于新增
- 编辑时，Start URL、Region 和 Scopes 会预填；按 Enter 保留默认值
- Scopes 会去重并校验；空输入会回退到 cloudidentity:account:access, offline_access

###### 配置 SSO Profile（configure sso）

```shell
bp configure sso --profile [profile name] --sso-session [session name]
```

该命令会配置 SSO profile，关联 SSO session，执行设备授权流程，并将 profile 写入 ~/.byteplus/config.json。行为和参数：

```shell
profile: profile 名称；为空时使用 {sso-role-name}-{sso-account-id} 作为默认值
sso-session: SSO session 名称；如果省略，会进入交互式选择/创建模式
no-browser: 在命令行中添加 `--no-browser` 参数会禁止自动打开浏览器；省略时默认自动打开浏览器。
```

说明：
- 如果 profile 已存在且不是 SSO profile，该命令会拒绝覆盖
- 如果 SSO session 不存在，会引导你创建它并输入 Start URL、Region 和 Scopes
- `bp configure sso` 不会自动切换当前激活 profile；如果希望将其作为默认 profile 使用，请执行 `bp configure profile --profile [profile name]`
- 重新配置已有 SSO profile 时，会先清理旧的 STS 临时凭证，再写入更新后的 profile，避免复用过期角色凭证

#### SSO 登录和退出

##### SSO 登录（sso login）

```shell
bp sso login --profile [profile name]
# 或
bp sso login --sso-session [session name]
```

该命令会执行一次显式 SSO 登录，运行新的授权流程，获取 access token 并缓存。每次执行都会重新认证当前 SSO session，不会静默地用已有 refresh_token 交换 access_token。参数：

```shell
profile: 要使用的 SSO profile；必须存在，类型必须为 sso，并且已配置 sso-session
sso-session: 要使用的 SSO session；该 session 必须存在且有效
no-browser: 在命令行中添加 `--no-browser` 参数会禁止自动打开浏览器；省略时默认自动打开浏览器。
```

登录行为：
- 如果既未提供 profile，也未提供 sso-session：无 session 时返回错误；只有一个 session 时直接使用；否则进入带搜索能力的交互式选择
- 仅支持 device code flow；使用 `--no-browser` 可禁止自动打开浏览器
- 运行业务命令时，如果缓存的 SSO access token 已过期或接近过期，CLI 会先尝试使用缓存的 refresh_token 静默刷新 access token，再请求角色凭证
- 如果 refresh_token 缺失、client registration 已过期或刷新失败，业务命令不会自动打开浏览器重新授权，而会提示你再次运行 `bp sso login`

##### SSO 退出（sso logout）

```shell
bp sso logout --sso-session [session name]
```

该命令会通过撤销缓存 token 并清理本地凭证来退出 SSO。参数：

```shell
sso-session: 要退出的 SSO session 名称
```

退出行为：
- 如果未提供 sso-session：无 session 时返回错误；只有一个 session 时退出该 session；否则进入交互式选择，并包含 "All SSO sessions"
- 批量退出会逐个退出 session，并在失败时返回聚合错误
- 退出会删除缓存 token，并清理 SSO profile 中的临时凭证字段（`access-key`、`secret-key`、`session-token`、`sts-expiration`），但不会删除 SSO profile、SSO session 配置、`account-id` 或 `role-name`

#### SSO FAQ

- `bp configure sso` 之后，为什么业务命令仍然使用旧账号？
  执行 `bp configure profile --profile [profile name]` 切换当前激活 profile。`configure sso` 只写入 profile，不会切换 `current`。

- 什么时候需要运行 `bp sso login`？
  首次执行 `bp configure sso` 时已经完成授权。日常业务命令会复用或静默刷新凭证。只有在 CLI 提示需要重新登录，或你想显式刷新 SSO 登录状态时，才需要运行 `bp sso login`。

- 为什么 `bp sso login` 每次都会打开授权流程？
  它是显式登录命令。每次运行都会重新开始设备授权，不会静默使用已有 refresh token。

- 如何切换账号或角色？
  重新运行 `bp configure sso --profile [profile name] --sso-session [session name]`，并选择不同账号或角色。

- 在没有浏览器的机器上如何登录？
  在 `bp configure sso` 或 `bp sso login` 中添加 `--no-browser`，然后在浏览器中打开终端输出的授权 URL。

- Scopes 应该怎么填？
  通常省略即可。默认值为 `cloudidentity:account:access,offline_access`。如果手动设置，只允许 `cloudidentity:account:access` 和 `offline_access`。

#### Console Login（login）

```shell
# 使用默认 profile 登录；未指定 region 时会提示输入
bp login

# 推荐：指定 profile 和 region；支持 -p/-r 简写
bp login -p dev -r ap-southeast-1

# 面向无头/远程服务器或容器的跨设备登录
bp login -p dev -r ap-southeast-1 --remote
```

该命令会通过 Byteplus 控制台执行 OAuth 2.0 + PKCE 登录，并在本地缓存临时 STS 凭证。后续 CLI 命令可以复用该 profile 的临时凭证。

参数：
- `--profile, -p`：Profile 名称，默认值为 `default`
- `--region, -r`：Region；如果未指定，会提示输入；按 Enter 使用 `ap-southeast-1`
- `--remote`：跨设备登录；按终端展示的 URL 在浏览器中完成登录，然后将 authorization code 粘贴回终端
- `--endpoint-url`：登录服务 URL，默认值为 `https://signin.byteplus.com`；通常无需修改

说明：
- 登录成功后，profile 会被设置为 `console-login` 模式，并记录 `login-session`
- 使用非 `default` profile 登录后，CLI 命令不会自动切换 profile；请先运行 `bp configure profile --profile dev`
- 当本地浏览器无法重定向回 CLI 所在机器时，使用 `--remote`

完整使用示例：

```shell
# 1. 登录并写入 dev profile
bp login --profile dev --region ap-southeast-1

# 2. 如果 dev 不是当前默认 profile，请先切换
bp configure profile --profile dev

# 3. 使用 console-login 临时凭证运行 CLI 命令
bp ecs DescribeInstances

# 4. 使用完成后清理本地登录状态
bp logout --profile dev
```

#### Console Logout（logout）

```shell
# 退出 default profile
bp logout

# 退出指定 profile
bp logout -p dev

# 退出当前配置中的所有 Console Login profile
bp logout --all
```

该命令会清理由 `bp login` 创建的本地登录状态。它只删除本地缓存凭证并清理 profile 中的 `login-session`，不会删除 profile，也不会发起任何服务端请求。

参数：
- `--profile, -p`：Profile 名称，默认值为 `default`
- `--all`：清理所有 `console-login` profile；启用后会忽略 `--profile`

说明：
- 未指定 `--profile` 时，`bp logout` 只处理 `default` profile；不会自动按 `current` 退出
- 只有 `console-login` profile 可以使用该命令；AK/SSO profile 不适用
- 退出后，重新运行 `bp login` 即可再次认证

#### 配置自动补全

运行 `bp completion --help` 查看如何启用 shell 自动补全。

##### Bash

1. 查看 bash 脚本：`bp completion bash`
2. 安装 bash-completion：

```shell
yum install bash-completion
# 或
apt-get install bash-completion
```

3. 启用：

```shell
source /usr/share/bash-completion/bash_completion
```

将其添加到 `~/.bashrc` 以持久化。

4. 配置自动补全：

```shell
echo 'source <(bp completion bash)' >> ~/.bashrc
bp completion bash > /etc/bash_completion.d/bp
```

然后重新加载 shell，或执行 `source ~/.bashrc`

##### Zsh

1. 启用 `compinit`：

```shell
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

2. 配置：

```shell
bp completion zsh > "${fpath[1]}/_bp"
```

然后重新加载 shell，或执行 `source ~/.zshrc`

#### 配置彩色输出

启用颜色：

```shell
bp enable-color
```

禁用颜色：

```shell
bp disable-color
```

### 使用 `bp`

#### 列出服务和 Action

1. 查看可用服务：

```shell
bp [-h|--help]
```

2. 查看服务支持的 action：

```shell
bp ecs [-h|--help]
```

#### 查看 CLI 版本

```shell
bp version
```

```shell
bp -v
```

#### 调用 API

基本结构：

```shell
bp <service> <action> [--param1 val1 --param2 val2 ...]
```

##### 运行时覆盖参数

调用 API 时，使用 `---` 前缀传递 CLI 内部覆盖参数。这些参数不会和 API 参数冲突：

- `---profile`：在运行时使用指定 profile，且不修改配置文件。后面必须跟一个已存在的 profile 名称。
- `---region`：在运行时覆盖 region。

```shell
# 使用指定 profile 调用 API。
bp ecs DescribeInstances ---profile prod

# 使用指定 profile 并覆盖 region 调用 API。
bp ecs DescribeInstances ---profile prod ---region ap-southeast-1

# 仅为本次 API 调用覆盖 region。
bp ecs DescribeInstances ---region ap-southeast-1 --InstanceId "i-xxx"

# `--` 参数仍会发送给 API，而 `---` 参数由 CLI 消费。
bp ecs DescribeInstances ---profile prod ---region ap-southeast-1 --InstanceIds '["i-xxx"]'
```

示例：

```shell
bp ecs DescribeInstances
bp rds_mysql ListDBInstanceIPLists --InstanceId "xxxxxx"
```

支持 JSON 输入：

```shell
bp rds_mysql ModifyDBInstanceIPList --InstanceId "xxxxxx" --GroupName "xxxxxx" --IPList '["10.20.30.40", "50.60.70.80"]'
```

支持通过 `--body` 传入 application/json：

```shell
bp rds_mysql ModifyDBInstanceIPList --body '{"InstanceId":"xxxxxx", "GroupName": "xxxxxx", "IPList": ["10.20.30.40", "50.60.70.80"]}'
```

使用不同凭证模式：

```shell
# 通过 ecsrole profile 使用 ECS 实例角色凭证。
bp configure set --profile ecs-role --mode ecsrole --role-name MyRole --region ap-southeast-1
bp ecs DescribeInstances ---profile ecs-role

# 通过 oidc profile 使用 OIDC 凭证。
bp configure set --profile ci-oidc --mode oidc --region ap-southeast-1 \
  --oidc-token-file /var/run/secrets/oidc-token --role-trn trn:iam::123456789012:role/CIRole
bp ecs DescribeInstances ---profile ci-oidc

# 不使用当前激活 profile，改用默认凭证链。
export BYTEPLUS_ACCESS_KEY=AK
export BYTEPLUS_SECRET_KEY=SK
export BYTEPLUS_REGION=ap-southeast-1
bp ecs DescribeInstances
```

## 安全和隐私

本项目非常重视安全。
有关漏洞报告和支持版本，请参阅 [SECURITY.md](SECURITY.md)
