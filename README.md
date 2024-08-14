# go-kit - A collection of tools and utilities for Go<!-- omit from toc -->

<!-- markdownlint-disable MD033 -->
<p align="center">
    <a href="https://pkg.go.dev/github.com/lvlcn-t/go-kit"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/lvlcn-t/go-kit.svg"></a>
    <a href="/../../commits/" title="Last Commit"><img alt="Last Commit" src="https://img.shields.io/github/last-commit/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img alt="Open Issues" src="https://img.shields.io/github/issues/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../pulls" title="Open Pull Requests"><img alt="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lvlcn-t/go-kit?style=flat"></a>
</p>
<!-- markdownlint-enable MD033 -->

- [Overview](#overview)
  - [Documentations](#documentations)
  - [Examples](#examples)
- [Code of Conduct](#code-of-conduct)
- [Working Language](#working-language)
- [Support and Feedback](#support-and-feedback)
- [How to Contribute](#how-to-contribute)
- [Licensing](#licensing)

This repository is a collection of my personal tools and utilities for the Go programming language. It is designed to be used in a wide range of applications, from simple scripts to complex web applications.

It is designed to be simple and easy to use, with a focus on performance and reliability. It is also designed to be flexible and extensible, so that you can easily add new features and functionality as needed.

## Overview

The repository is divided into several modules, each of which provides a different set of tools and utilities. The following is a brief overview of each module:

- [apimanager](/apimanager/README.md): A simple way to manage an API server with the ability to add routes, route-groups, and middlewares. It is built on top of the [gofiber/fiber](https://github.com/gofiber/fiber) framework.
- [config](/config/README.md): A wrapper around [spf13/viper](https://github.com/spf13/viper) to load configuration files with into a struct and validate them.
- [dependency](/dependency/README.md): A simple dependency injection container that allows you to register and resolve dependencies.
- [executors](/executors/README.md): Useful executors and policies to handle common scenarios like retries with exponential backoff and more.
- [lists](/lists/lists.go): A collection of functions to complement the [`slices`](https://pkg.go.dev/slices) package in the standard library.
- [metrics](/metrics/README.md): A set of tools to collect and expose metrics for a Go application.

### Documentations

You can find the documentation for each module in their respective directories. Each module has its own `README.md` file that provides an overview of the module and how to use it.

### Examples

You can find examples for each module in the [`example`](/example) directory. Each example demonstrates how to use the module in a specific scenario.

## Code of Conduct

This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of
conduct. Please see the details in our [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). All contributors must abide by the code
of conduct.

## Working Language

We decided to apply _English_ as the primary project language.

Consequently, all content will be made available primarily in English.
We also ask all interested people to use English as the preferred language to create issues,
in their code (comments, documentation, etc.) and when you send requests to us.
The application itself and all end-user facing content will be made available in other languages as needed.

## Support and Feedback

The following channels are available for discussions, feedback, and support requests:

| Type       | Channel                                                                                                                  |
| ---------- | ------------------------------------------------------------------------------------------------------------------------ |
| **Issues** | [![General Discussion](https://img.shields.io/github/issues/lvlcn-t/go-kit?style=flat-square)](/../../issues/new/choose) |

## How to Contribute

Contribution and feedback is encouraged and always welcome. For more information about how to contribute, the project
structure, as well as additional contribution information, see our [Contribution Guidelines](./CONTRIBUTING.md). By
participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright (c) 2024 lvlcn-t.

Licensed under the **MIT** (the "License"); you may not use this file except in compliance with
the License.

You may obtain a copy of the License at <https://www.mit.edu/~amini/LICENSE.md>.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "
AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the [LICENSE](./LICENSE) for
the specific language governing permissions and limitations under the License.
