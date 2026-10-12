// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"testing"

	"forgejo.org/modules/json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Real-world Forgejo Actions .well-known/openid-configuration from Forgejo v15
const forgejoOIDC = `
{
    "issuer": "https://example.org/api/actions",
    "jwks_uri": "https://example.org/api/actions/.well-known/keys",
    "subject_types_supported": [
        "public"
    ],
    "response_types_supported": [
        "id_token"
    ],
    "claims_supported": [
        "sub",
        "aud",
        "exp",
        "iat",
        "iss",
        "nbf",
        "actor",
        "base_ref",
        "event_name",
        "head_ref",
        "ref",
        "ref_protected",
        "ref_type",
        "repository",
        "repository_owner",
        "run_attempt",
        "run_id",
        "run_number",
        "sha",
        "workflow",
        "workflow_ref"
    ],
    "id_token_signing_alg_values_supported": [
        "RS256"
    ],
    "scopes_supported": [
        "openid"
    ]
}
`

// Real-world Forgejo Actions JWKS from Forgejo v15
const forgejoJWKS = `
{
    "keys": [
        {
            "alg": "RS256",
            "e": "AQAB",
            "kid": "SNNttXGzw6l53JC158lXddjSjQ5bJ9bdTTqTi12gaLY",
            "kty": "RSA",
            "n": "7RL963BVzemasfImhlR3KUX97YdA7g3SBnq_ZLzcdxLXPGDhsnSoxMX7gY30b1qpQlML8yiAyz_gxUydiVlqpEEPypR9lfKtZXv4JTM-X2rccegcreUyfFJnFuzVUoY7SVEzAulLUwqP8MH8kxDI7JZRQ8_JIjm9IxEuWCSc3XnVxNCTS2XEdHsug_Kt6SQdcH8xL9U2w0EHAUna9KkLAl6_PzBg1JxIQDHQtfp_CN7YNyoyilH88XAGEeQm0fLz6GH7hhyw6y1b9NprYIxNrdD4Pb1b66j4K--bCJy530UmEAlfLbCiDCh4k78TPUnU_YwwT4ujC0t28zoHNB3Y0w",
            "use": "sig"
        }
    ]
}
`

// Real-world GitHub Actions /.well-known/openid-configuration
const githubOIDC = `
{
    "issuer": "https://token.actions.githubusercontent.com",
    "jwks_uri": "https://token.actions.githubusercontent.com/.well-known/jwks",
    "subject_types_supported": [
        "public",
        "pairwise"
    ],
    "response_types_supported": [
        "id_token"
    ],
    "claims_supported": [
        "sub",
        "aud",
        "exp",
        "iat",
        "iss",
        "jti",
        "nbf",
        "ref",
        "sha",
        "repository",
        "repository_id",
        "repository_owner",
        "repository_owner_id",
        "enterprise",
        "enterprise_id",
        "run_id",
        "run_number",
        "run_attempt",
        "actor",
        "actor_id",
        "workflow",
        "workflow_ref",
        "workflow_sha",
        "head_ref",
        "base_ref",
        "event_name",
        "ref_type",
        "ref_protected",
        "environment",
        "environment_node_id",
        "job_workflow_ref",
        "job_workflow_sha",
        "repository_visibility",
        "runner_environment",
        "issuer_scope",
        "check_run_id"
    ],
    "id_token_signing_alg_values_supported": [
        "RS256"
    ],
    "scopes_supported": [
        "openid"
    ]
}
`

// Real-world GitHub Actions JWKS
const githubJWKS = `
{
    "keys": [
        {
            "kty": "RSA",
            "alg": "RS256",
            "use": "sig",
            "kid": "cc413527-173f-5a05-976e-9c52b1d7b431",
            "n": "w4M936N3ZxNaEblcUoBm-xu0-V9JxNx5S7TmF0M3SBK-2bmDyAeDdeIOTcIVZHG-ZX9N9W0u1yWafgWewHrsz66BkxXq3bscvQUTAw7W3s6TEeYY7o9shPkFfOiU3x_KYgOo06SpiFdymwJflRs9cnbaU88i5fZJmUepUHVllP2tpPWTi-7UA3AdP3cdcCs5bnFfTRKzH2W0xqKsY_jIG95aQJRBDpbiesefjuyxcQnOv88j9tCKWzHpJzRKYjAUM6OPgN4HYnaSWrPJj1v41eEkFM1kORuj-GSH2qMVD02VklcqaerhQHIqM-RjeHsN7G05YtwYzomE5G-fZuwgvQ",
            "e": "AQAB"
        },
        {
            "kty": "RSA",
            "alg": "RS256",
            "use": "sig",
            "kid": "38826b17-6a30-5f9b-b169-8beb8202f723",
            "n": "5Manmy-zwsk3wEftXNdKFZec4rSWENW4jTGevlvAcU9z3bgLBogQVvqYLtu9baVm2B3rfe5onadobq8po5UakJ0YsTiiEfXWdST7YI2Sdkvv-hOYMcZKYZ4dFvuSO1vQ2DgEkw_OZNiYI1S518MWEcNxnPU5u67zkawAGsLlmXNbOylgVfBRJrG8gj6scr-sBs4LaCa3kg5IuaCHe1pB-nSYHovGV_z0egE83C098FfwO1dNZBWeo4Obhb5Z-ZYFLJcZfngMY0zJnCVNmpHQWOgxfGikh3cwi4MYrFrbB4NTlxbrQ3bL-rGKR5X318veyDlo8Dyz2KWMobT4wB9U1Q",
            "e": "AQAB",
            "x5c": [
                "MIIDKzCCAhOgAwIBAgIUDnwm6eRIqGFA3o/P1oBrChvx/nowDQYJKoZIhvcNAQELBQAwJTEjMCEGA1UEAwwaYWN0aW9ucy5zZWxmLXNpZ25lZC5naXRodWIwHhcNMjQwMTIzMTUyNTM2WhcNMzQwMTIwMTUyNTM2WjAlMSMwIQYDVQQDDBphY3Rpb25zLnNlbGYtc2lnbmVkLmdpdGh1YjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOTGp5svs8LJN8BH7VzXShWXnOK0lhDVuI0xnr5bwHFPc924CwaIEFb6mC7bvW2lZtgd633uaJ2naG6vKaOVGpCdGLE4ohH11nUk+2CNknZL7/oTmDHGSmGeHRb7kjtb0Ng4BJMPzmTYmCNUudfDFhHDcZz1Obuu85GsABrC5ZlzWzspYFXwUSaxvII+rHK/rAbOC2gmt5IOSLmgh3taQfp0mB6Lxlf89HoBPNwtPfBX8DtXTWQVnqODm4W+WfmWBSyXGX54DGNMyZwlTZqR0FjoMXxopId3MIuDGKxa2weDU5cW60N2y/qxikeV99fL3sg5aPA8s9iljKG0+MAfVNUCAwEAAaNTMFEwHQYDVR0OBBYEFIPALo5VanJ6E1B9eLQgGO+uGV65MB8GA1UdIwQYMBaAFIPALo5VanJ6E1B9eLQgGO+uGV65MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAGS0hZE+DqKIRi49Z2KDOMOaSZnAYgqq6ws9HJHT09MXWlMHB8E/apvy2ZuFrcSu14ZLweJid+PrrooXEXEO6azEakzCjeUb9G1QwlzP4CkTcMGCw1Snh3jWZIuKaw21f7mp2rQ+YNltgHVDKY2s8AD273E8musEsWxJl80/MNvMie8Hfh4n4/Xl2r6t1YPmUJMoXAXdTBb0hkPy1fUu3r2T+1oi7Rw6kuVDfAZjaHupNHzJeDOg2KxUoK/GF2/M2qpVrd19Pv/JXNkQXRE4DFbErMmA7tXpp1tkXJRPhFui/Pv5H9cPgObEf9x6W4KnCXzT3ReeeRDKF8SqGTPELsc="
            ],
            "x5t": "ykNaY4qM_ta4k2TgZOCEYLkcYlA"
        },
        {
            "kty": "RSA",
            "alg": "RS256",
            "use": "sig",
            "kid": "38E9B30B3A023A1B72309921A69A42FCC496C42C",
            "n": "tEq2Fp9HcdT5MwMsB_UTm8j_woJJLi3sA-y0RX2tioTm581seyfvOH6lJ5JmHVtS-_fb8B2tRT1pznHQSNq14PsJdu9bp5egbWmIz-5RvhqoM-oKem_MJENCNFuqXijRLT47FRdfH3inqde1vJlA_JJHCqYMKIpHH7kqNFYcCpwr0vk80Hc2rTyL0uBXI7NqBZbtUgNoyucWO5O7QQrPNOmlr-GI8aFckFRfobCaCOiH9qW02FtkV74fwBGVCNhNf3a1CK81-O8xEGimvVydI_pQA5B8QqVuQjY_ntOu555HdirA0hKkY6fsE9eZCMFmWDHZ2kSWLjhabxWxIzSzXQ",
            "e": "AQAB",
            "x5c": [
                "MIIDrDCCApSgAwIBAgIQbuIOJTcGQ4GOQs29F1/uLzANBgkqhkiG9w0BAQsFADA2MTQwMgYDVQQDEyt2c3RzLXZzdHNnaHJ0LWdoLXZzby1vYXV0aC52aXN1YWxzdHVkaW8uY29tMB4XDTI1MDkxMTE5MzcxOFoXDTI3MDkxMTE5NDcxOFowNjE0MDIGA1UEAxMrdnN0cy12c3RzZ2hydC1naC12c28tb2F1dGgudmlzdWFsc3R1ZGlvLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALRKthafR3HU+TMDLAf1E5vI/8KCSS4t7APstEV9rYqE5ufNbHsn7zh+pSeSZh1bUvv32/AdrUU9ac5x0EjateD7CXbvW6eXoG1piM/uUb4aqDPqCnpvzCRDQjRbql4o0S0+OxUXXx94p6nXtbyZQPySRwqmDCiKRx+5KjRWHAqcK9L5PNB3Nq08i9LgVyOzagWW7VIDaMrnFjuTu0EKzzTppa/hiPGhXJBUX6Gwmgjoh/altNhbZFe+H8ARlQjYTX92tQivNfjvMRBopr1cnSP6UAOQfEKlbkI2P57TrueeR3YqwNISpGOn7BPXmQjBZlgx2dpEli44Wm8VsSM0s10CAwEAAaOBtTCBsjAOBgNVHQ8BAf8EBAMCBaAwCQYDVR0TBAIwADAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwNgYDVR0RBC8wLYIrdnN0cy12c3RzZ2hydC1naC12c28tb2F1dGgudmlzdWFsc3R1ZGlvLmNvbTAfBgNVHSMEGDAWgBQLxdObdnfWzaBcxau87tdSUtEuQjAdBgNVHQ4EFgQUC8XTm3Z31s2gXMWrvO7XUlLRLkIwDQYJKoZIhvcNAQELBQADggEBAD2Eo703wXQgB2vJn/RwTTcHGkeMkYXm0mWCxOSh4iCKVvqypBJrmLzRkMMJN0/10qIGciYWUl6EkL7yj48tpXXH01Ep0ONDdo9UYmKGp81Z4j3u3FBJTVQSdj2tnPOPZlYWaBkerIkcIeyWBRKvne1UBaobbk84epfBmUfAMFmyJEk+x+q7cqmsbjDtdrmhiWaInqCijpS2dW2MitJ5F7tBBS26SMTqLQteA2IOwIW1BMlYIPuSO3dKn/rYVS8RjL+x+MxP98vla5sichoEZwVWnXiXgFZ4n/asGqc+Da9q6ILLtInvgI5bi7kjJJ2ARTRC5/a+J/v3EL+t8SdnOO8="
            ],
            "x5t": "OOmzCzoCOhtyMJkhpppC_MSWxCw"
        },
        {
            "kty": "RSA",
            "alg": "RS256",
            "use": "sig",
            "kid": "4F3E9AD8C9A6F5EB3173006F4FA630E28F43DCE9",
            "n": "tGevqhkBGn8NB0dKxs8Ddxhn-xZPm55svcSlkJZEOwDOXDLl_0-iVOVKNJfcHHLHvMqa6zh2DDcpAWZi2FpeBAJupsrymqwzllxOODWKWoVIoaIjOO7h1JLiF9Knwuq-o6BPtKdwOT-bOrXRzChMtQsc5C1Auex-D0Z6loObBuK1Lkm0RK9ISQsLqBEwq8g0OOupI_shU1r2rT2G0nkZ0CvxVlQeUGShFi8Mdys2s5LPqBwjC4LKwjk8moWQV32KEccbTPKxnG_539DxRglHJgHPHisSVGsfZIUXi2chtXdQHZPdVve8ZRmknCykZtkJ6K87llSUXi7oyzhCIZdiUQ",
            "e": "AQAB",
            "x5c": [
                "MIIDrDCCApSgAwIBAgIQPQS35v3ITW6fNLO8GX5QBjANBgkqhkiG9w0BAQsFADA2MTQwMgYDVQQDEyt2c3RzLXZzdHNnaHJ0LWdoLXZzby1vYXV0aC52aXN1YWxzdHVkaW8uY29tMB4XDTI1MDgwNjE0MTEzMloXDTI3MDgwNjE0MjEzMlowNjE0MDIGA1UEAxMrdnN0cy12c3RzZ2hydC1naC12c28tb2F1dGgudmlzdWFsc3R1ZGlvLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALRnr6oZARp/DQdHSsbPA3cYZ/sWT5uebL3EpZCWRDsAzlwy5f9PolTlSjSX3Bxyx7zKmus4dgw3KQFmYthaXgQCbqbK8pqsM5ZcTjg1ilqFSKGiIzju4dSS4hfSp8LqvqOgT7SncDk/mzq10cwoTLULHOQtQLnsfg9GepaDmwbitS5JtESvSEkLC6gRMKvINDjrqSP7IVNa9q09htJ5GdAr8VZUHlBkoRYvDHcrNrOSz6gcIwuCysI5PJqFkFd9ihHHG0zysZxv+d/Q8UYJRyYBzx4rElRrH2SFF4tnIbV3UB2T3Vb3vGUZpJwspGbZCeivO5ZUlF4u6Ms4QiGXYlECAwEAAaOBtTCBsjAOBgNVHQ8BAf8EBAMCBaAwCQYDVR0TBAIwADAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwNgYDVR0RBC8wLYIrdnN0cy12c3RzZ2hydC1naC12c28tb2F1dGgudmlzdWFsc3R1ZGlvLmNvbTAfBgNVHSMEGDAWgBRKoYOga736JYE15vT7b4gWjC1hwTAdBgNVHQ4EFgQUSqGDoGu9+iWBNeb0+2+IFowtYcEwDQYJKoZIhvcNAQELBQADggEBAJVZIPtoZUlvqgu+Pl0nj8WopA8iuy1m7JRg5fg+bOIGFhXFR8+mH8prpeodjUQ40q2Hq6IwnVir+G56zVwAPf2HHksqdp8be9qjkTjD0mJorPCt/lumrKoNGOVmYffYuIyr73hwsl8fN6sGjAyXLFBkozE4s5ssbeodFxiYE1A61SXnzldC00M7qWleMWjTUBixiZ+R/eroddkLNBGDv9ewDrTQv1ipNec89+Wi7Wb6SAXNxBADiC5kVlFylBgHo3oZNg3KFzZS01REyc4zdH7v1wfZzilLluI6ygTyYRYpJCsKrX5D9JW196f2PCzcs+VXMfneRDnvyfjep7Y1Pi8="
            ],
            "x5t": "Tz6a2Mmm9esxcwBvT6Yw4o9D3Ok"
        }
    ]
}
`

// Real-world .well-known/openid-configuration extracted from a AWS Federated Web Identity endpoint.
const awsOIDC = `
{
    "claims_supported": [
        "sub",
        "iss",
        "aud",
        "exp",
        "iat",
        "jti",
        "https://sts.amazonaws.com/"
    ],
    "id_token_signing_alg_values_supported": [
        "RS256",
        "ES384"
    ],
    "issuer": "https://a103a2cc-b461-473d-84fe-6c4f6d45af88.tokens.sts.global.api.aws",
    "jwks_uri": "https://a103a2cc-b461-473d-84fe-6c4f6d45af88.tokens.sts.global.api.aws/.well-known/jwks.json",
    "subject_types_supported": [
        "public"
    ]
}
`

// Real-world JWKS extracted from a AWS Federated Web Identity endpoint.
const awsJWKS = `
{
    "keys": [
        {
            "e": "AQAB",
            "kid": "RSA_0",
            "kty": "RSA",
            "n": "3AvB0UECoYssZEgSMTa4SYvfqstJxkhbBBSKAFRUW6f_McJ9CAXkTi6YkG0NGm77ZIRW12_gOLKZJUHWp9CMAbmk0O4sMIx8K6Ap7-6qjkt7FYvl4mkQVJd-pU-yE3SJn0S5xEbCYXulgrrGN8POysTblqN0BfrdDAYTVhWQ47rbm--3QrRcVN9XCjlMBVXYauaN6KlszKL6NTe7GWilauYBsVHw7d4ekliuEGGA6zJNGz595KD7yofRc1euFs86KgiFj0mpudCqG39jIlBJ4vZSJPw1Rsvhg8THqlxhmurVYr9TuckLJa5fpEL78xGs3Ar4GIM6w0sxLDbdY-KdCQ",
            "use": "sig"
        },
        {
            "alg": "ES384",
            "crv": "P-384",
            "kid": "EC384_0",
            "kty": "EC",
            "use": "sig",
            "x": "ad_olFw0n3XBA114sefjlirPf2gX6bKqT-kD2lQzfQzkWW1TetKIUWah3md-UgV9",
            "y": "w6GzW2Oen4G7Ei1bFaDkBpPSulvkSznb6YtG79NWK9UjgDqfN6am9lUs-bF8VN7v"
        }
    ]
}
`

func TestParseOpenIDConfiguration(t *testing.T) {
	t.Run("Forgejo", func(t *testing.T) {
		var retval AuthorizedIntegrationOpenIDConfiguration
		data := []byte(forgejoOIDC)
		require.NoError(t, json.Unmarshal(data, &retval))
		assert.Equal(t, "https://example.org/api/actions/.well-known/keys", retval.JwksURI)
	})
	t.Run("GitHub", func(t *testing.T) {
		var retval AuthorizedIntegrationOpenIDConfiguration
		data := []byte(githubOIDC)
		require.NoError(t, json.Unmarshal(data, &retval))
		assert.Equal(t, "https://token.actions.githubusercontent.com/.well-known/jwks", retval.JwksURI)
	})
	t.Run("AWS", func(t *testing.T) {
		var retval AuthorizedIntegrationOpenIDConfiguration
		data := []byte(awsOIDC)
		require.NoError(t, json.Unmarshal(data, &retval))
		assert.Equal(t, "https://a103a2cc-b461-473d-84fe-6c4f6d45af88.tokens.sts.global.api.aws/.well-known/jwks.json", retval.JwksURI)
	})
}

func TestParseJSONWebKeySet(t *testing.T) {
	t.Run("Forgejo", func(t *testing.T) {
		var retval AuthorizedIntegrationOpenIDKeys
		data := []byte(forgejoJWKS)
		require.NoError(t, json.Unmarshal(data, &retval))
		assert.Len(t, retval.Keys, 1)
	})
	t.Run("GitHub", func(t *testing.T) {
		var retval AuthorizedIntegrationOpenIDKeys
		data := []byte(githubJWKS)
		require.NoError(t, json.Unmarshal(data, &retval))
		assert.Len(t, retval.Keys, 4)
	})
	t.Run("AWS", func(t *testing.T) {
		var retval AuthorizedIntegrationOpenIDKeys
		data := []byte(awsJWKS)
		require.NoError(t, json.Unmarshal(data, &retval))
		assert.Len(t, retval.Keys, 2)
	})
}
