
# Contributing

go-snaps is a pet open-source project. That means there are no any sorts of guarantees but
I am happy to review any contribution to help the project improve.

## Running Locally

Running the project locally is straight forward. You will need `go>=1.16` version installed.

In the project there is an `examples` folder where you can experiment and test how `go-snaps` works. You can run the tests with 

```go
go test ./examples/... -v -race -count=1
```

## Creating a pr

Before making a pull request make sure your changes are linted and formatted else GH actions will fail.

In the root of the project there is a `Makefile`. You can run `make help` for the list of commands. 
