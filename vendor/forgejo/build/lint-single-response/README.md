# lint-single-response

The lint-single-response Go analyzer attempts to prevent a common problem in Forgejo where it is possible for a web handler to provide a response to a request, and then continue code execution unintentionally.  For example:

```go
err := json.Unmarshal(data, &claims)
if err != nil {
    ctx.Error(http.StatusInternalServerError, "Error in unmarshal", err)
    // Oops, I forgot to `return` here...
}
// ... more work occurs ...
ctx.JSON(http.StatusOK, resp)
```

In order to detect these cases, lint-single-response contains a list of functions that deliver a web response.  When any of those functions are used within a function, the control flow must not perform any work after the function is invoked -- it can only return and exit the function.

Methods named `Test...` are omitted from analysis, as this naming scheme suggests a test case where an error would have no user impact, and such methods sometimes invoke web response methods in unusual but safe patterns.

## Limitations

lint-single-response only works within the control-flow of a single function.  If a web handler calls another function that invokes `ctx.Error(...)`, then there is no guarantee that the web handler doesn't go on to do more work.  This could be addressed in the future but would require a multi-pass analysis -- all functions that invoke web responses would need to be identified, then all functions that invoke those functions would need to be identified, recursively, until no new functions are identified.  And then lint-single-response's current behaviour would need to be implemented against that entire set of functions.

## Usage

Direct invocation:

```
go run ./build/lint-single-response/cmd ./...
```

It is also integrated into Forgejo's `Makefile`, and can be run directly as the target `make lint-single-response`, or as part of `make lint-backend` or `make pr-go`.

## Testing

lint-single-response contains internal tests to verify that it works correctly. These tests are included in `make test-backend`, but, Go tends to think that they're cached even if data in `testdata` is changed. For development and testing of lint-single-response, it is recommended to run the tests with `-count 1` to avoid caching:

```
GOTESTFLAGS="-count 1" GO_TEST_PACKAGES=forgejo.org/build/lint-single-response make test-backend
```

Testing is done with the [`analysistest` package](https://pkg.go.dev/golang.org/x/tools@v0.46.0/go/analysis/analysistest#Run). In short, comments `// want ...` indicate that a lint diagnostic must be produced on that line for the test to pass.

An empty implementation of `context.Base`, `context.Context`, and `context.APIContext` are included in the test package so that the exact method signatures being used in Forgejo can be covered in the tests.

