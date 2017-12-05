# Style

## Go

### Formatting

All Go code should be formatted with `gofmt`. This is enforced by the `pre-commit` hook.

Code can be easily formatted by running `make format`.

### Vetting

All code should pass `go tool vet`. This is not enforced by any hooks right now but may be in the future.

Code can be vetted by running `make vet`.

### Linting

Code should generally pass `golint`.

Not every exported function, type, constant, and variable need to be commented, but in general they should be. Other than that in general code should pass the linter.

You can run the linter by running `make lint`.

You can run the linter without warnings about exported values not being commented by running `make lint-no-export-comments`.

## Terraform

### Formatting

All Go code should be formatted with `terraform fmt`. This is enforced by the `pre-commit` hook.

Code can be easily formatted by running `make format`.
