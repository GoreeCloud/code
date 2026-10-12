;;; Copyright 2025 The Forgejo Authors. All rights reserved.
;;; SPDX-License-Identifier: MIT
;;;
;;; Commentary:
;;;
;;; This is a GNU Guix manifest that can be used to create a
;;; development environment to build and test Forgejo.
;;;
;;; The following is a usage example to create a containerized
;;; environment, with HOME shared for the Go cache and the network
;;; made available to fetch required Go and Node dependencies.
;;;
#|

guix shell -CNF --share=$HOME -m manifest.scm
export GOTOOLCHAIN=local     # to use the Go binary from Guix
export CC=gcc CGO_ENABLED=1
# The following is to preserve debug info symbols:
export STRIP=0 CGO_CFLAGS='-O0 -g' EXTRA_GOFLAGS='-gcflags="all=-N -l"'
export TAGS="timetzdata sqlite sqlite_unlock_notify"
make clean
make -j$(nproc)
make test -j$(nproc)         # run unit tests
make test-sqlite -j$(nproc)  # run integration tests
make watch                   # run an instance/rebuild on changes

# For debugging, you can either attach the delve debugger like this:
dlv attach $(pgrep gitea)

# Or start Forgejo directly with it:
dlv exec ./gitea

|#

(use-modules (guix packages)
             (guix utils))

(define go-1.26                         ;minimal version required by Forgejo
  (specification->package "go@1.26"))

(define (package-with-go base go)
  "Return a variant of BASE, a Guix package object built with GO, another
package."
  (package/inherit base
    (arguments (ensure-keyword-arguments (package-arguments base)
                                         (list #:go go)))))

(define packages-for-debugging
  (list (specification->package "procps")   ;pgrep
        (package-with-go (specification->package "delve") go-1.26)     ;debugger
        (package-with-go (specification->package "gomacro") go-1.26))) ;Go REPL

(concatenate-manifests
 (list
  (packages->manifest
   (append (list go-1.26)
           packages-for-debugging))
  (specifications->manifest
   (list "bash-minimal"
         "coreutils"
         "diffutils"
         "findutils"
         "gcc-toolchain"
         "git"                            ;libpcre support is required
         "git-lfs"
         "gnupg"
         "grep"
         "make"
         "node"
         "nss-certs"
         "openssh"
         "sed"))))
