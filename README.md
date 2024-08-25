# prep

`prep` is a small Go tool that enables compile-time function evaluation. By using `prep.Comptime`, you can evaluate functions at build time, replacing them with their computed results. Just like `comptime` from Zig. Except it's not.

## Features

- **Compile-Time Evaluation**: Replace function calls with their computed results at build time.
- **Simple Integration**: Use `prep` as both a Go library and a standalone executable.
- **Tooling Support**: Easily integrate `prep` with your Go build process using `-toolexec`.

## Installation

### 1. Install the `prep` executable

To use `prep` in your build process, install it as a binary executable:

```bash
go install github.com/pijng/prep@latest
```

### 2. Add `prep` as a dependency

To use the `prep` library in your Go project, add it via Go modules:

```bash
go get github.com/pijng/prep
go mod tidy
```

## Usage

### Wrapping Functions with `prep.Comptime`

To mark a function for compile-time evaluation, wrap it with `prep.Comptime`:

```go
package main

import (
  "fmt"
  "github.com/pijng/prep"
)

func main() {
  // This will be evaluated at compile-time
  result := prep.Comptime(fibonacci(9999999))

  fmt.Println("Result:", result)
}

func fibonacci(n int) int {
  if n <= 1 {
    return n
  }
  return fibonacci(n-1) + fibonacci(n-2)
}
```

### Building with `prep`

After wrapping your functions with `prep.Comptime`, you need to use `prep` during the build process. This is done by using the `-toolexec` flag:

```bash
go build -a -toolexec="prep <absolute/path/to/project>" main.go
```

This command will evaluate all functions wrapped with `prep.Comptime` and replace them with their computed results during the build.

**Important:**
  * `-a` flag is required to recompile all your project, otherwise go compiler might do nothing and use cached build
  * `<absolute/path/to/project>` is and absolute path to the root of your project. If you run `go build` from the root of the project – simply specify `$PWD` as an argument.

### Run the final binary:

```bash
./main
```

## Limitations

* Basic Literals Only: Currently, `prep.Comptime` only supports basic literals as arguments. This means you cannot pass variables to the functions that are wrapped with `prep.Comptime`.
* Compile-Time Evaluation: Only functions that can be fully resolved with the provided literal arguments will be evaluated at compile-time.

## Motivation

Because reasons.