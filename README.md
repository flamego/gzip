# gzip

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/flamego/gzip/Go?logo=github&style=for-the-badge)](https://github.com/flamego/gzip/actions?query=workflow%3AGo)
[![Codecov](https://img.shields.io/codecov/c/gh/flamego/gzip?logo=codecov&style=for-the-badge)](https://app.codecov.io/gh/flamego/gzip)
[![GoDoc](https://img.shields.io/badge/GoDoc-Reference-blue?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/flamego/gzip?tab=doc)
[![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?style=for-the-badge&logo=sourcegraph)](https://sourcegraph.com/github.com/flamego/gzip)

Package gzip is a middleware that provides gzip compression to responses for [Flamego](https://github.com/flamego/flamego).

## Installation

The minimum requirement of Go is **1.16**.

    go get github.com/flamego/gzip


## Getting started

```go
package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/gzip"
)

func main() {
	f := flamego.Classic()
	f.Use(gzip.Gzip())
	f.Get("/", func() string {
		return "ok"
	})
	f.Run()
}
```

## Getting help

- Read [documentation and examples](https://flamego.dev/middleware/gzip.html).
- Please [file an issue](https://github.com/flamego/flamego/issues) or [start a discussion](https://github.com/flamego/flamego/discussions) on the [flamego/flamego](https://github.com/flamego/flamego) repository.

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
