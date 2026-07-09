Getting Started | [Authentication](2-Authentication.md)

---

## Getting Started

This guide explains how to install `bp`, put it on your PATH, and make a minimal API call.

## Requirements

- Go 1.12+ is recommended.
- Use the `bp` command prefix in scripts and shell examples.

## Install

### Install with npm

The CLI is published as the npm package `@byteplus/cli`. If Node.js >= 14 is available, install it globally:

```shell
npm install -g @byteplus/cli
```

The package provides the `bp` command:

```shell
bp version
bp --help
```

To upgrade to the latest version:

```shell
npm update -g @byteplus/cli
```

### Download from Release

1. Open <https://github.com/byteplus-sdk/byteplus-cli/releases>.
2. Download the archive for your OS and architecture.
3. Extract it to get `bp`, or `bp.exe` on Windows.

### Build from Source

The repository provides `build.sh`. It can auto-detect the current OS and architecture, or build for an explicit target.

```shell
# Build for the current machine
sh build.sh

# Specify OS; architecture is still auto-detected
sh build.sh darwin
sh build.sh linux
sh build.sh windows

# Cross-compile with an explicit architecture: amd64 / arm64 / 386 / arm
sh build.sh linux amd64

# Show help
sh build.sh -h
```

The output binary is `bp`, or `bp.exe` on Windows.

## Configure PATH

When installed globally with npm, npm places `bp` in the global bin directory. If the command is unavailable, check whether npm's global bin directory is in PATH:

```shell
npm bin -g
```

When using Release or a source build, make sure the directory containing `bp` is in your PATH. A common setup is:

```shell
sudo cp bp /usr/local/bin
```

Verify the command:

```shell
bp version
bp --help
```

If `/usr/local/bin` is not in `$PATH`, configure PATH for your shell.

## Minimal Configuration

The most direct setup is an AK/SK profile:

```shell
bp configure set --profile default --region ap-southeast-1 --access-key AK --secret-key SK
```

You can also skip the config file and use environment variables:

```shell
export BYTEPLUS_ACCESS_KEY=AK
export BYTEPLUS_SECRET_KEY=SK
export BYTEPLUS_REGION=ap-southeast-1
```

See [Authentication](2-Authentication.md) for more credential modes.

## First API Call

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
bp ecs DescribeRegions --help
```

Call an API:

```shell
bp sts GetCallerIdentity
```

Override region for one invocation:

```shell
bp sts GetCallerIdentity ---region ap-southeast-1
```

`---region` is a CLI fixed flag and does not conflict with API parameters written as `--Param value`. See [Usage](4-Usage.md) for more examples.

---

Getting Started | [Authentication](2-Authentication.md)
