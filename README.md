# go-kit - A collection of tools and utilities for Go<!-- omit from toc -->

<!-- markdownlint-disable MD033 -->
<p align="center">
    <a href="https://pkg.go.dev/github.com/lvlcn-t/go-kit"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/lvlcn-t/go-kit.svg"></a>
    <a href="/../../commits/" title="Last Commit"><img alt="Last Commit" src="https://img.shields.io/github/last-commit/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img alt="Open Issues" src="https://img.shields.io/github/issues/lvlcn-t/go-kit?style=flat"></a>
    <a href="/../../pulls" title="Open Pull Requests"><img alt="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lvlcn-t/go-kit?style=flat"></a>
</p>
<!-- markdownlint-enable MD033 -->

- [Introduction](#introduction)
- [Installation](#installation)
- [Usage](#usage)
  - [Configuration](#configuration)
  - [Lists](#lists)
  - [Executors](#executors)
  - [Metrics](#metrics)
- [Code of Conduct](#code-of-conduct)
- [Working Language](#working-language)
- [Support and Feedback](#support-and-feedback)
- [How to Contribute](#how-to-contribute)
- [Licensing](#licensing)

This library is a collection of my personal tools and utilities for the Go programming language. It is designed to be used in a wide range of applications, from simple scripts to complex web applications.

It is designed to be simple and easy to use, with a focus on performance and reliability. It is also designed to be flexible and extensible, so that you can easily add new features and functionality as needed.

## Introduction

The library is divided into several packages, each of which provides a different set of tools and utilities. The following is a brief overview of each package:

- [config](/config/loader.go): A wrapper around [spf13/viper](https://github.com/spf13/viper) to load configuration files with into a struct without the need to write boilerplate code.
- [lists](/lists/lists.go): A collection of functions to work with lists, such as filtering, mapping, and reducing.
- [executors](/executors/retry.go): Some useful executors to handle common scenarios like retries with exponential backoff.
- [metrics](/metrics/metrics.go): A simple wrapper around [otel](https://opentelemetry.io/docs/languages/go/getting-started/) to get a trace provider based on the provided configuration.

## Installation

To install, run:

```bash
go get github.com/lvlcn-t/go-kit
```

And then import the wanted package in your code:

```go
import "github.com/lvlcn-t/go-kit/<package>"
```

## Usage

The following is a brief overview of how you can use the packages provided by this library:

### Configuration

You can use the `config` package to load configuration files into a struct. The package provides a simple API to load configuration files and bind them to a struct.

Here is an example of how you can use the `config` package to load a configuration file into a struct:

```go
package main

import (
  "fmt"

  "github.com/lvlcn-t/go-kit/config"
)

type Config struct {
  Host string `mapstructure:"host"`
  Port int    `mapstructure:"port"`
}

func (c Config) IsEmpty() bool {
  return c == (Config{})
}

func main() {
  // Load the configuration file
  cfg, err := config.Load[Config]("config.yaml")
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(cfg.Host) // Output: localhost
  fmt.Println(cfg.Port) // Output: 8080
}
```

### Lists

You can use the `lists` package to work with lists in Go. The package provides a collection of functions to work with lists, such as filtering, mapping, and reducing.

Here is an example of how you can use the `lists` package to filter a list of integers:

```go
package main

import (
  "fmt"

  "github.com/lvlcn-t/go-kit/lists"
)

func main() {
  numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

  // Filter the list to only include even numbers
  evens := lists.Filter(numbers, func(n int) bool {
    return n%2 == 0
  })

  fmt.Println(evens) // Output: [2 4 6 8 10]
}
```

### Executors

The `executors` package provides some useful executors to handle common scenarios like retries with exponential backoff or a simple executor to run a function with a timeout.

Here is an example of how you can use the `executors` package to create an executor with multiple policies that are applied to the task in the order of the method calls:

```go
package main

import (
  "context"
  "fmt"
  "time"

  "github.com/lvlcn-t/go-kit/executors"
)

func main() {
  // Create a task that may fail and needs to be retried
  task := executors.Effector(func(ctx context.Context) error {
    // Do something that may fail like an HTTP request
    return nil
  })

  // Apply multiple policies to the task
  task = task.WithRetry(executors.DefaultRetrier).
    WithTimeout(1*time.Second).
    WithRateLimit(executors.RateLimit(1)).
    WithCircuitBreaker(3, 1*time.Second)

  // Run the task with the applied policies with a context
  err := task.Do(context.Background())
  if err != nil {
    // Handle the error after all retries
    fmt.Println(err)
  }
}
```

### Metrics

_tbd_

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
