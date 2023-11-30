# How to write your use case

1. First create your domain struct if it doesn't exist in `pkg/domain` directory. Package name is the entity name and
   the struct is named as Entity following go conventions.
2. Create your api structs under `pkg/api` but remember to group it under a `Use case subject` same as your use case.
   Refer `internal/usecase/README.md` for understanding this.
3. Create your use case under `internal/usecase` by following `internal/usecase/README.md`
4. If there is any need for specific error codes then keep them in `code.go`
5. Call `mux.Handle` in `main.go` and pass your `handler` (generic) along with your other parameters.  
