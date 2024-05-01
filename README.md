# gitty

<!--TODO: badges (release, github actions, codecov, go report, license) -->
[![codecov](https://codecov.io/gh/worlpaker/gitty/graph/badge.svg?token=6T5V3C1IXE)](https://codecov.io/gh/worlpaker/gitty)

![gitty](_examples/gitty.gif)

## Overview

Gitty is a CLI tool that helps you download GitHub files and directories directly! It is fast, simple, and works concurrently. Gitty is easy to use, making it perfect for downloading specific files and folders from large repositories.

## Install

### Go

> Requires **Go 1.22+**

```sh
go install github.com/worlpaker/gitty@latest
```

## Usage

```sh
gitty github-url
```

### Examples

- Download GitHub Directory

```sh
gitty https://github.com/worlpaker/go-syntax/tree/master/examples
```

- Download GitHub File

```sh
gitty https://github.com/worlpaker/go-syntax/blob/master/test/semantic_tokens.go
```

- Gitty also works without the https prefix

```sh
gitty github.com/worlpaker/go-syntax/tree/master/examples
```

## Authorization

GitHub has **hourly** [rate limit](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api):

- For **unauthenticated** users: **60**

- For **authenticated** users: **5000**

- For **enterprise cloud** users: **15000**

Gitty retrieves your GitHub token from your os environment variable with `GH_TOKEN` key. Get your token from [GitHub Personal Tokens](https://github.com/settings/tokens). More details about tokens can be found on [Managing your personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens).

You can set your token manually in your os environment variable with `GH_TOKEN` key, or you can use **gitty**!

- Set token

```sh
gitty -s=your_github_token
```

- Unset token

```sh
gitty -u
```

- Authenticated user info

```sh
gitty -a
```

- Check client auth status and remaninig rate limit

```sh
gitty -c
```

> **NOTE:** Gitty doesn't store your token. It gets, saves, and deletes the token from your os environment variable.

## How it works

Gitty uses [go-github](https://github.com/google/go-github) to interact with GitHub and [cobra](https://github.com/spf13/cobra) for CLI.

## Test

Test coverage is **100%** for Windows, Linux and macOS (see: [gitty test commands](Makefile)).

## Contributing

Yes, please! Feel free to contribute.

## Credits

Inspired by [download-directory](https://github.com/download-directory/download-directory.github.io).

## License

<!--TODO: add license -->
