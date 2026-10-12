// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestCSRFProtection(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// Creates a cross origin HTTP request. Indicate it's cross origin via the
	// `Sec-Fetch-Site` header.
	crossOriginSecFetchRequest := func(t *testing.T, method, urlStr string) *RequestWrapper {
		t.Helper()

		req := NewRequest(t, method, urlStr)
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		return req
	}

	// Creates a cross origin HTTP request. Indicate it's cross origin via the
	// `Origin` and `Host` header.
	crossOriginHostRequest := func(t *testing.T, method, urlStr string) *RequestWrapper {
		t.Helper()

		req := NewRequest(t, method, urlStr)
		req.Header.Set("Origin", "https://evil.com")
		// Cannot set Host as header, `req.Host` is used (which is normally parsed)
		// from the Host header, but in testing we can't do this.
		req.Host = "forgejo.org"
		return req
	}

	// Creates a same origin HTTP request. Indicate it's same origin via the
	// `Sec-Fetch-Site` header.
	sameOriginSecFetchRequest := func(t *testing.T, method, urlStr string) *RequestWrapper {
		t.Helper()

		req := NewRequest(t, method, urlStr)
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		return req
	}

	// Creates a same origin HTTP request. Indicate it's same origin via the
	// `Origin` and `Host` header.
	sameOriginHostRequest := func(t *testing.T, method, urlStr string) *RequestWrapper {
		t.Helper()

		req := NewRequest(t, method, urlStr)
		req.Header.Set("Origin", "https://forgejo.org")
		// Cannot set Host as header, `req.Host` is used (which is normally parsed)
		// from the Host header, but in testing we can't do this.
		req.Host = "forgejo.org"
		return req
	}

	success := func(t *testing.T, session *TestSession, statusCode int, req *RequestWrapper) {
		t.Helper()
		session.MakeRequest(t, req, statusCode)
	}

	failure := func(t *testing.T, session *TestSession, req *RequestWrapper) {
		t.Helper()
		resp := session.MakeRequest(t, req, http.StatusForbidden)
		if req.Header.Get("Sec-Fetch-Site") != "" {
			assert.Equal(t, "cross-origin request detected from Sec-Fetch-Site header\n", resp.Body.String())
		} else {
			assert.Equal(t, "cross-origin request detected, and/or browser is out of date: Sec-Fetch-Site is missing, and Origin does not match Host\n", resp.Body.String())
		}
	}

	user2 := loginUser(t, "user2")

	// CSRF is not checked for safe methods: GET, HEAD, OPTIONS.
	t.Run("Safe methods", func(t *testing.T) {
		// There's no route that does not ignore CSRF and allows OPTIONS.
		t.Run("Cross-origin", func(t *testing.T) {
			t.Run("Get", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				success(t, user2, http.StatusOK, crossOriginSecFetchRequest(t, http.MethodGet, "/user/settings"))
				success(t, user2, http.StatusOK, crossOriginHostRequest(t, http.MethodGet, "/user/settings"))
			})

			t.Run("Head", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				success(t, user2, http.StatusOK, crossOriginSecFetchRequest(t, http.MethodHead, "/user/settings"))
				success(t, user2, http.StatusOK, crossOriginHostRequest(t, http.MethodHead, "/user/settings"))
			})
		})

		t.Run("Same-origin", func(t *testing.T) {
			t.Run("Get", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				success(t, user2, http.StatusOK, sameOriginSecFetchRequest(t, http.MethodGet, "/user/settings"))
				success(t, user2, http.StatusOK, sameOriginHostRequest(t, http.MethodGet, "/user/settings"))
			})

			t.Run("Head", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				success(t, user2, http.StatusOK, sameOriginSecFetchRequest(t, http.MethodHead, "/user/settings"))
				success(t, user2, http.StatusOK, sameOriginHostRequest(t, http.MethodHead, "/user/settings"))
			})
		})
	})

	// Request is blocked for non-safe methods on routes that check for CSRF.
	t.Run("Non-safe methods", func(t *testing.T) {
		t.Run("Cross-origin", func(t *testing.T) {
			t.Run("Post", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				failure(t, user2, crossOriginSecFetchRequest(t, http.MethodPost, "/user/settings"))
				failure(t, user2, crossOriginHostRequest(t, http.MethodPost, "/user/settings"))
			})

			t.Run("Put", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				failure(t, user2, crossOriginSecFetchRequest(t, http.MethodPut, "/user2/repo1/projects/1/2"))
				failure(t, user2, crossOriginHostRequest(t, http.MethodPut, "/user2/repo1/projects/1/2"))
			})

			t.Run("Delete", func(t *testing.T) {
				failure(t, user2, crossOriginSecFetchRequest(t, http.MethodDelete, "/user2/repo1/projects/1/2"))
				failure(t, user2, crossOriginHostRequest(t, http.MethodDelete, "/user2/repo1/projects/1/2"))
			})
		})

		t.Run("Same-origin", func(t *testing.T) {
			t.Run("Post", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				success(t, user2, http.StatusSeeOther, sameOriginSecFetchRequest(t, http.MethodPost, "/user/settings"))
				success(t, user2, http.StatusSeeOther, sameOriginHostRequest(t, http.MethodPost, "/user/settings"))
			})

			t.Run("Put", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				success(t, user2, http.StatusOK, sameOriginSecFetchRequest(t, http.MethodPut, "/user2/repo1/projects/1/2"))
				success(t, user2, http.StatusOK, sameOriginHostRequest(t, http.MethodPut, "/user2/repo1/projects/1/2"))
			})

			t.Run("Delete", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				success(t, user2, http.StatusOK, sameOriginSecFetchRequest(t, http.MethodDelete, "/user2/repo1/projects/1/2"))
				success(t, user2, http.StatusOK, sameOriginHostRequest(t, http.MethodDelete, "/user2/repo1/projects/1/3"))
			})
		})

		// Check routes where CSRF protection is disabled.
		t.Run("Cross-origin ignored", func(t *testing.T) {
			t.Run("Post", func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				req := crossOriginSecFetchRequest(t, http.MethodPost, "/login/oauth/introspect")
				req.SetBasicAuth("da7da3ba-9a13-4167-856f-3899de0b0138", "4MK8Na6R55smdCY0WuCCumZ6hjRPnGY5saWVRHHjJiA=")
				success(t, user2, http.StatusOK, req)

				req = crossOriginHostRequest(t, http.MethodPost, "/login/oauth/introspect")
				req.SetBasicAuth("da7da3ba-9a13-4167-856f-3899de0b0138", "4MK8Na6R55smdCY0WuCCumZ6hjRPnGY5saWVRHHjJiA=")
				success(t, user2, http.StatusOK, req)
			})
		})
	})
}
