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
go install github.com/pijng/prep/cmd/prep@latest
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
  result := prep.Comptime(fibonacci(300))

  fmt.Println("Result:", result)
}

func fibonacci(n int) int {
  fmt.Printf("calculating fibonacci for %d\n", n)

  if n <= 1 {
    return n
  }
  a, b := 0, 1
  for i := 2; i <= n; i++ {
    a, b = b, a+b
  }
  return b
}
```

### Building with `prep`

After wrapping your functions with `prep.Comptime`, you need to use `prep` during the build process. This is done by using the `-toolexec` flag:

```bash
go build -toolexec="prep" main.go
```

This command will evaluate all functions wrapped with `prep.Comptime` and replace them with their computed results during the build.

### Run the final binary:

```bash
./main
```

Note the speed and absence of `calculating fibonacci for %d` message.

## Limitations

**Basic Literal Arguments**

At it's core, `prep.Comptime` allows you to provide basic literals (e.g., strings, integers, floats) as arguments. Which means you can either:

  1. Pass a basic literal directly like this:
  ```go
  func job() {
    prep.Comptime(myFunc(1))
  }
  ```

  2. Use a variable with the value of basic literal from the same scope as wrapped function:

  ```go
  func job() {
    x := 1
    y := 2
    prep.Comptime(myFunc(x, y))
  }
  ```

These literals are evaluated directly at compile-time

**Compile-Time Evaluation (Experimental)**

Additionally, `prep.Comptime` also supports evaluation during compilation when you can use values from IO operations, for example:

  ```go
  func main() {
    config := prep.Comptime(readFile("./test.txt"))
    fmt.Println(config)
  }

  func readFile(path string) string {
    content, _ := os.ReadFile(path)

    return string(content)
  }
  ```

This way the `prep` will call `readFile` function as part of compilation step and include it's result to the final binary. This means you can remove the `test.txt` file after compilation, and still be able to execute the binary as if it exists.

Note that this feature is somewhat experimental and prone to errors.

## Motivation

Because reasons.
