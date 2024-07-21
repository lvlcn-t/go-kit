# go-kit/executors - Executors Module<!-- omit from toc -->

<!-- markdownlint-disable MD033 -->
<p align="center">
    <a href="https://pkg.go.dev/github.com/lvlcn-t/go-kit"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/lvlcn-t/go-kit/executors.svg"></a>
    <a href="/../../commits/" title="Last Commit"><img alt="Last Commit" src="https://img.shields.io/github/last-commit/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img alt="Open Issues" src="https://img.shields.io/github/issues/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../pulls" title="Open Pull Requests"><img alt="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lvlcn-t/go-kit?style=flat"></a>
</p>
<!-- markdownlint-enable MD033 -->

- [About this module](#about-this-module)
- [Installation](#installation)
- [Usage](#usage)
- [Code of Conduct](#code-of-conduct)
- [Working Language](#working-language)
- [Support and Feedback](#support-and-feedback)
- [How to Contribute](#how-to-contribute)
- [Licensing](#licensing)

The `executors` module provides some useful executors to handle common scenarios in like retrying with exponential backoff, rate limiting, and more.

## About this module

This module contains a set of executors and policies to handle common scenarios in Go applications. To see a list of all available executors, check the [pkg.go.dev](https://pkg.go.dev/github.com/lvlcn-t/go-kit/executors) documentation.

## Installation

To install, run:

```bash
go get github.com/lvlcn-t/go-kit/executors
```

And import the package in your code:

```go
import "github.com/lvlcn-t/go-kit/executors"
```

## Usage

The documentation for this module can be found on [pkg.go.dev](https://pkg.go.dev/github.com/lvlcn-t/go-kit/executors).

To see how to use this module, you can check the [examples](../example/executors) directory of the [repository](https://github.com/lvlcn-t/go-kit).

## Code of Conduct

This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of conduct. Please see the details in our [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md). All contributors must abide by the code of conduct.

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

Contribution and feedback is encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](../CONTRIBUTING.md). By participating in this project, you agree to abide by its [Code of Conduct](../CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright (c) 2024 lvlcn-t.

Licensed under the **MIT** (the "License"); you may not use this file except in compliance with
the License.

You may obtain a copy of the License at <https://www.mit.edu/~amini/LICENSE.md>.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "
AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the [LICENSE](../LICENSE) for
the specific language governing permissions and limitations under the License.
