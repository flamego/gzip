// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/flamego/flamego"
	"github.com/klauspost/compress/gzip"
)

const (
	_HEADER_ACCEPT_ENCODING  = "Accept-Encoding"
	_HEADER_CONTENT_ENCODING = "Content-Encoding"
	_HEADER_CONTENT_TYPE     = "Content-Type"
	_HEADER_VARY             = "Vary"
)

// Options represents a struct for specifying configuration options for the GZip middleware.
type Options struct {
	// Compression level. Can be DefaultCompression(-1), ConstantCompression(-2)
	// or any integer value between BestSpeed(1) and BestCompression(9) inclusive.
	CompressionLevel int
}

func isCompressionLevelValid(level int) bool {
	return level == gzip.DefaultCompression ||
		level == gzip.ConstantCompression ||
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
		if !strings.Contains(ctx.Request().Header.Get(_HEADER_ACCEPT_ENCODING), "gzip") {
			return
		}

		headers := ctx.ResponseWriter().Header()
		headers.Set(_HEADER_CONTENT_ENCODING, "gzip")
		headers.Set(_HEADER_VARY, _HEADER_ACCEPT_ENCODING)

		// We've made sure compression level is valid in prepareGzipOptions,
		// no need to check same error again.
		gz, err := gzip.NewWriterLevel(ctx.ResponseWriter(), opt.CompressionLevel)
		if err != nil {
			panic(err.Error())
		}
		defer gz.Close()

		gzw := gzipResponseWriter{gz, ctx.ResponseWriter()}
		ctx.MapTo(gzw, (*http.ResponseWriter)(nil))

		ctx.Next()

		// delete content length after we know we have been written to
		gzw.Header().Del("Content-Length")
	})
}

type gzipResponseWriter struct {
	w *gzip.Writer
	flamego.ResponseWriter
}

func (grw gzipResponseWriter) Write(p []byte) (int, error) {
	if len(grw.Header().Get(_HEADER_CONTENT_TYPE)) == 0 {
		grw.Header().Set(_HEADER_CONTENT_TYPE, http.DetectContentType(p))
	}
	return grw.w.Write(p)
}

func (grw gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := grw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}
