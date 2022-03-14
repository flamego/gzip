// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flamego/flamego"
)

func TestGzip(t *testing.T) {
	calledBefore := false

	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(Gzip(Options{-10}))
	f.Use(func(r http.ResponseWriter) {
		r.(flamego.ResponseWriter).Before(func(rw flamego.ResponseWriter) {
			calledBefore = true
		})
	})
	f.Get("/", func() string { return "hello world!" })

	// Not accepting gzip
	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)

	f.ServeHTTP(resp, req)

	ce := resp.Header().Get(headerContentEncoding)
	assert.NotEqual(t, "gzip", ce)

	// Accepting gzip
	resp = httptest.NewRecorder()
	req.Header.Set(headerAcceptEncoding, "gzip")
	f.ServeHTTP(resp, req)

	ce = resp.Header().Get(headerContentEncoding)
	assert.Equal(t, "gzip", ce)

	r, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)

	body, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "hello world!", string(body))

	assert.True(t, calledBefore)
}

type hijackableResponse struct {
	Hijacked bool
	header   http.Header
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{header: make(http.Header)}
}

func (h *hijackableResponse) Header() http.Header       { return h.header }
func (h *hijackableResponse) Write([]byte) (int, error) { return 0, nil }
func (h *hijackableResponse) WriteHeader(int)           {}
func (h *hijackableResponse) Flush()                    {}
func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func TestResponseWriterHijack(t *testing.T) {
	hijackable := newHijackableResponse()

	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(Gzip())
	f.Use(func(rw http.ResponseWriter) {
		hj, ok := rw.(http.Hijacker)
		require.True(t, ok)

		_, _, err := hj.Hijack()
		assert.Nil(t, err)
	})

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)

	r.Header.Set(headerAcceptEncoding, "gzip")
	f.ServeHTTP(hijackable, r)

	assert.True(t, hijackable.Hijacked)
}
