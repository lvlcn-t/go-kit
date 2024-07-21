# go-kit/config - Config Module<!-- omit from toc -->

<!-- markdownlint-disable MD033 -->
<p align="center">
    <a href="https://pkg.go.dev/github.com/lvlcn-t/go-kit"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/lvlcn-t/go-kit/config.svg"></a>
    <a href="/../../commits/" title="Last Commit"><img alt="Last Commit" src="https://img.shields.io/github/last-commit/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img alt="Open Issues" src="https://img.shields.io/github/issues/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../pulls" title="Open Pull Requests"><img alt="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lvlcn-t/go-kit?style=flat"></a>
</p>
<!-- markdownlint-enable MD033 -->

- [About this module](#about-this-module)
- [Installation](#installation)
- [Usage](#usage)
  - [Validation](#validation)
    - [Available tags](#available-tags)
- [Code of Conduct](#code-of-conduct)
- [Working Language](#working-language)
- [Support and Feedback](#support-and-feedback)
- [How to Contribute](#how-to-contribute)
- [Licensing](#licensing)

The `config` module provides a set of tools to manage configurations in Go applications.

## About this module

This module can be used to load configurations from different sources like environment variables and configuration files and bind them to a struct using the [spf13/viper](https://github.com/spf13/viper) library. It also provides a way to validate the configuration.

## Installation

To install, run:

```bash
go get github.com/lvlcn-t/go-kit/config
```

And import the package in your code:

```go
import "github.com/lvlcn-t/go-kit/config"
```

## Usage

The documentation for this module can be found on [pkg.go.dev](https://pkg.go.dev/github.com/lvlcn-t/go-kit/config).

To see how to use this module, you can check the [examples](../example/config) directory of the [repository](https://github.com/lvlcn-t/go-kit).

### Validation

To validate a configuration, you can use the `config.Validate` function. You can either implement the `config.Validator` interface on the passed type or use the `validate` tag on the struct fields.

```go
type Config struct {
    Host string `validate:"required"`
    Port int    `validate:"required,min=1024,max=65535"`
}
```

#### Available tags

The following tags are available for validation:

| Tag        | Description                                                    | Example               | Available Types                                                                                                                                                                            |
| ---------- | -------------------------------------------------------------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `required` | The field must be set                                          | `validate:"required"` | `any`                                                                                                                                                                                      |
| `min`      | The field must be greater than or equal to the specified value | `validate:"min=10"`   | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered), [slices](https://go.dev/tour/moretypes/7), [arrays](https://go.dev/tour/moretypes/6) and [maps](https://go.dev/tour/moretypes/19)         |
| `max`      | The field must be less than or equal to the specified value    | `validate:"max=10"`   | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered), [slices](https://go.dev/tour/moretypes/7), [arrays](https://go.dev/tour/moretypes/6) and [maps](https://go.dev/tour/moretypes/19)         |
| `len`      | The field must have the specified length                       | `validate:"len=10"`   | `string`, [slices](https://go.dev/tour/moretypes/7), [arrays](https://go.dev/tour/moretypes/6), [maps](https://go.dev/tour/moretypes/19) and [channels](https://go.dev/tour/concurrency/2) |
| `eq`       | The field must be equal to the specified value                 | `validate:"eq=10"`    | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered)                                                                                                                                            |
| `ne`       | The field must not be equal to the specified value             | `validate:"ne=10"`    | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered)                                                                                                                                            |
| `gt`       | The field must be greater than the specified value             | `validate:"gt=10"`    | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered)                                                                                                                                            |
| `lt`       | The field must be less than the specified value                | `validate:"lt=10"`    | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered)                                                                                                                                            |
| `gte`      | The field must be greater than or equal to the specified value | `validate:"gte=10"`   | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered)                                                                                                                                            |
| `lte`      | The field must be less than or equal to the specified value    | `validate:"lte=10"`   | [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered)                                                                                                                                            |

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
