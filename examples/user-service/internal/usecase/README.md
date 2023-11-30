# Use case

This package contains all `Use cases` grouped by `Use case subject` as package

Create one package for each `Use case subject` and then create a file of same name if there are not many functions. If
the functions explode then divide the functions based on the domain entity that is being mainly mutated or fetched.

> Example: Create `user.go` and `role.go` under `managing` package if `managing.go` becomes bloated
