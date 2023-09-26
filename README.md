[![CI](https://github.com/tf2d2/tf2d2/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/tf2d2/tf2d2/actions/workflows/ci.yml)
[![codeql](https://github.com/tf2d2/tf2d2/actions/workflows/codeql.yml/badge.svg?branch=main)](https://github.com/tf2d2/tf2d2/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/tf2d2/tf2d2/graph/badge.svg?token=ZUBADJX1MU)](https://codecov.io/gh/tf2d2/tf2d2)
[![Go Reference](https://pkg.go.dev/badge/github.com/tf2d2/tf2d2.svg)](https://pkg.go.dev/github.com/tf2d2/tf2d2)
[![Go Report Card](https://goreportcard.com/badge/github.com/tf2d2/tf2d2)](https://goreportcard.com/report/github.com/tf2d2/tf2d2)

# `tf2d2`

Generate [d2](https://terrastruct.com/) diagrams from [Terraform](https://www.terraform.io/).

## Cloud Providers

Supported Terraform cloud provider(s) are:

- [AWS](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

Note that only specific provider resources are included in the generated `d2` diagram in order to better represent the infrastructure deployed via the supported Terraform cloud provider(s).

## Installation

### MacOS

```console
brew install tf2d2/tap/tf2d2
```

### Binary

> Note: Windows binaries are available in `.zip` format.

```console
curl -Lo ./tf2d2.tar.gz https://github.com/tf2d2/tf2d2/releases/download/v0.1.0/tf2d2-v0.1.0-$(uname)-amd64.tar.gz
tar -xzf tf2d2.tar.gz
chmod +x tf2d2
mv tf2d2 /usr/local/tf2d2
```

Alternatively, download the latest stable binary for your platform from the [Releases](https://github.com/tf2d2/tf2d2/releases) page  and copy to the desired location.

### Source

To compile and build `tf2d2` from source, run the following:

```console
git clone https://github.com/tf2d2/tf2d2
cd tf2d2
go mod tidy
go build -o tf2d2 .
./tf2d2 --version
```

## Usage

`tf2d2` can generate a `d2` diagram from the following sources.

### Local Terraform State

By default, `tf2d2` uses local Terraform state in the current directory:

```console
tf2d2
```

### Remote Terraform State

```console
tf2d2 --organization foo --workspace bar --token abc123
```

## Configuration

`tf2d2` can be configured with a YAML file. The default name is `.tf2d2.yml`.

The path in which the configuration file is located can be:

1. Current directory
2. `$HOME/.tf2d2.yml`

```yaml
hostname: app.terraform.io
# required if using remote Terraform state
organization: ""
workspace: ""
token: "" # it's recommended to use TF_TOKEN env variable

state-file: "terraform.tfstate"

output-file: out.svg # output can be .svg or .png

verbose: false
dry-run: false
```

Alternatively, `tf2d2` can be configured with ENV variables prefixed with `TF_` that match the uppercased name of available CLI flags. For example:

> Note: ENV variables must be delimited with `_` instead of `-`.

```console
export TF_TOKEN=my-token
export TF_STATE_FILE=my-state-file
```

## Compatibility

This project follows the [Go support policy](https://go.dev/doc/devel/release#policy). Only two latest major releases of Go are supported by the project.

Currently, that means **Go 1.20** or later must be used when developing or testing code.

## Credits

This project is inspired by [`cdk2d2`](https://github.com/megaproaktiv/cdk2d2).

## License

[Apache v2.0 License](./LICENSE)
