# Tripswitch

[![Go Reference](https://pkg.go.dev/badge/github.com/mgiaccone/tripswitch.svg)](https://pkg.go.dev/github.com/mgiaccone/tripswitch)

[Tripswitch][repo_url] implements a set of resiliency patterns for Go 1.18+ with no external dependencies.

***WARNING: This project is in its infancy and it is not ready for production use.***

## Installation

```
go get github.com/mgiaccone/tripswitch
```

## Circuit Breaker

The complete documentation can be found [here](docs/circuitbreaker.md).

### Default configuration

| Property          | Value | Description                                                           |
|-------------------|-------|-----------------------------------------------------------------------|
| Failure threshold | 3     | Number of subsequent failures required to open the circuit            |
| Success threshold | 3     | Number of subsequent successes required to close the circuit          |
| Wait interval     | 30s   | Amount of time an open circuit will wait before attempting a recovery |

### Basic usage

**Use a named circuit breaker with default configuration**

```go
// lazily create and use a "sample" circuit breaker using the default configuration
res, err := breaker.Do[int]("sample", func() (int, error) {
    // body of the protected function
    return 1, nil
})
// handle error
```

**Use a named circuit breaker with custom configuration**
```go
// set the configuration for the "sample" circuit with a failure threshold of 5
breaker.MustConfigure[int]("sample", breaker.WithFailThreshold(5))

// use the previously configured "sample" circuit breaker
res, err := breaker.Do[int]("sample", func() (int, error) {
    // body of the protected function
    return 1, nil
})
// handle error
```

**Override default options**

```go
// set the default wait interval to 10 seconds
breaker.DefaultOptions(breaker.WithWaitInterval(10 * time.Second))

// lazily create and use a "sample" circuit breaker using the overridden default configuration
res, err := breaker.Do[int]("sample", func() (int, error) {
    // body of the protected function
    return 1, nil
})
// handle error
```

## Retrier

The complete documentation can be found [here](docs/retrier.md).

### Quick start

```go
TODO: add quick start code sample
```

## Contributing

**Contributions will not be accepted until after the first release**

Please make sure you read and follow the guidelines [here](docs/contributing.md).

## License

The MIT License (MIT)

See [LICENSE](LICENSE) for details.

[repo_url]: https://github.com/mgiaccone/tripswitch
