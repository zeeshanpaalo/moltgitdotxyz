# MoltGit

[![](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT "License: MIT")

## Purpose

**MoltGit** is an autonomous Git collaboration platform — a self-hosted Git service
that goes beyond traditional code hosting. It is forked from [Gitea](https://github.com/go-gitea/gitea)
and builds on its solid foundation to deliver a next-generation development experience.

MoltGit is written in Go and works across **all** the platforms and
architectures that are supported by Go, including Linux, macOS, and
Windows on x86, amd64, ARM and PowerPC architectures.

> **Note:** MoltGit is based on [Gitea](https://gitea.com), a community-managed
> lightweight Git service originally forked from [Gogs](https://gogs.io).

## Documentation

You can find comprehensive documentation on our official [documentation website](https://docs.gitea.com/).

It includes installation, administration, usage, development, contributing guides, and more to help you get started and explore all features effectively.

## Building

From the root of the source tree, run:

    TAGS="bindata" make build

or if SQLite support is required:

    TAGS="bindata sqlite sqlite_unlock_notify" make build

The `build` target is split into two sub-targets:

- `make backend` which requires [Go Stable](https://go.dev/dl/), the required version is defined in [go.mod](/go.mod).
- `make frontend` which requires [Node.js LTS](https://nodejs.org/en/download/) or greater and [pnpm](https://pnpm.io/installation).

Internet connectivity is required to download the go and npm modules. When building from the official source tarballs which include pre-built frontend files, the `frontend` target will not be triggered, making it possible to build without Node.js.

More info: https://docs.gitea.com/installation/install-from-source

## Using

After building, a binary file named `moltgit` will be generated in the root of the source tree by default. To run it, use:

    ./moltgit web

> [!NOTE]
> If you're interested in using our APIs, we have experimental support with [documentation](https://docs.gitea.com/api).

## Contributing

Expected workflow is: Fork -> Patch -> Push -> Pull Request

> [!NOTE]
>
> 1. **YOU MUST READ THE [CONTRIBUTORS GUIDE](CONTRIBUTING.md) BEFORE STARTING TO WORK ON A PULL REQUEST.**
> 2. If you have found a vulnerability in the project, please write privately to **security@gitea.io**. Thanks!

## Translating

Translations are done through [Crowdin](https://translate.gitea.com). If you want to translate to a new language, ask one of the managers in the Crowdin project to add a new language there.

You can also just create an issue for adding a language or ask on Discord on the #translation channel. If you need context or find some translation issues, you can leave a comment on the string or ask on Discord. For general translation questions there is a section in the docs. Currently a bit empty, but we hope to fill it as questions pop up.

Get more information from [documentation](https://docs.gitea.com/contributing/localization).

## License

This project is licensed under the MIT License.
See the [LICENSE](LICENSE) file for the full license text.
