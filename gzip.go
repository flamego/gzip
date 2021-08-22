// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"compress/gzip"
	"strings"

	"github.com/flamego/flamego"
)

const (
	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
)

// Options represents a struct for specifying configuration options for the GZip middleware.
type Options struct {
	// Compression level. Can be DefaultCompression(-1), NoCompression(0)
	// or any integer value between BestSpeed(1) and BestCompression(9) inclusive.
	CompressionLevel int
}

func isCompressionLevelValid(level int) bool {
	return level == gzip.DefaultCompression ||
		level == gzip.NoCompression ||
		(level >= gzip.BestSpeed && level <= gzip.BestCompression)
}

func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}

	if !isCompressionLevelValid(opt.CompressionLevel) {
		// For web content, level 4 seems to be a sweet spot.
		opt.CompressionLevel = 4
	}
	return opt
}

// Gzip returns a Handler that adds gzip compression to all requests.
// Make sure to include the Gzip middleware above other middleware
// that alter the response body (like the render middleware).
func Gzip(options ...Options) flamego.Handler {
	opt := prepareOptions(options)

	return flamego.ContextInvoker(func(ctx flamego.Context) {
		if !strings.Contains(ctx.Request().Header.Get(headerAcceptEncoding), "gzip") {
			return
		}

		headers := ctx.ResponseWriter().Header()
		headers.Set(headerContentEncoding, "gzip")
		headers.Set(headerVary, headerAcceptEncoding)

		// We've made sure compression level is valid in prepareGzipOptions,
		// no need to check same error again.
		gz, err := gzip.NewWriterLevel(ctx.ResponseWriter(), opt.CompressionLevel)
		if err != nil {
			panic(err.Error())
		}
		defer gz.Close()

		ctx.Next()

		// delete content length after we know we have been written to
		ctx.ResponseWriter().Header().Del("Content-Length")
	})
}
