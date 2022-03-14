// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"bufio"
	"compress/gzip"
	"net"
	"net/http"
	"strings"

	"github.com/flamego/flamego"
	"github.com/pkg/errors"
)

const (
	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerVary            = "Vary"
)

// Options represents a struct for specifying configuration options for the gzip
// middleware.
type Options struct {
	// CompressionLevel indicates the compression level. Default is 4.
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

// Gzip returns a Handler that adds gzip compression to all requests. Make sure
// to include the gzip middleware above other middleware that alter the response
// body (like the render middleware).
func Gzip(options ...Options) flamego.Handler {
	opt := prepareOptions(options)

	return flamego.ContextInvoker(func(ctx flamego.Context) {
		if !strings.Contains(ctx.Request().Header.Get(headerAcceptEncoding), "gzip") {
			return
		}

		headers := ctx.ResponseWriter().Header()
		headers.Set(headerContentEncoding, "gzip")
		headers.Set(headerVary, headerAcceptEncoding)

		// We've made sure compression level is valid in prepareOptions,
		// no need to check same error again.
		gz, err := gzip.NewWriterLevel(ctx.ResponseWriter(), opt.CompressionLevel)
		if err != nil {
			panic("gzip: " + err.Error())
		}
		defer func() { _ = gz.Close() }()

		w := &responseWriter{
			writer:         gz,
			ResponseWriter: ctx.ResponseWriter(),
		}
		ctx.MapTo(w, (*http.ResponseWriter)(nil))
		ctx.Next()

		// Delete content length after we know we have been written to
		ctx.ResponseWriter().Header().Del("Content-Length")
	})
}

type responseWriter struct {
	writer *gzip.Writer
	flamego.ResponseWriter
}

func (w *responseWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

var _ http.Hijacker = (*responseWriter)(nil)

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}
