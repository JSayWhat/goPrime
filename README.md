# Project Name

## Overview

This project is a Go application that demonstrates the use of channels, goroutines, and context for concurrent programming. It includes functions for channel fan-in, repeating functions with context, and more.

## Features

- **Channel Fan-In**: Combines multiple input channels into a single output channel.
- **Function Repeater**: Repeats a function execution until a done signal is received.
- **Logging**: Logs the status of goroutines at regular intervals.

## Requirements

- Go 1.18 or later

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/JSayWhat/goPrime.git
   cd yourproject
   ```

2. Build the project:
   ```sh
   go build -o main main.go
   ```

## Usage

1. Run the application:

   ```sh
   ./main
   ```

2. The application will start and demonstrate the concurrent features implemented.

## Code Overview

### `chanIn` Function

Combines multiple input channels into a single output channel and logs the status of the goroutines.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "math"
    "math/rand"
    "net/http"
    "runtime"
    "sync"
    "time"
)

func chanIn[T any](done <-chan int, channels ...<-chan T) <-chan T {
    var wg sync.WaitGroup
    fannedInStream := make(chan T)
    transfer := func(c <-chan T) {
        defer wg.Done()
        ticker := time.NewTicker(3 * time.Second) // Log every 3 seconds
        for i := range c {
            select {
            case <-done:
                log.Println("chanIn Go routine has finished.")
                ticker.Stop()
                return
            case <-ticker.C:
                log.Println("chanIn Go routine is still running...")
            case fannedInStream <- i:
            }
        }
    }
    for _, c := range channels {
        wg.Add(1)
        go transfer(c)
    }
    go func() {
        wg.Wait()
        close(fannedInStream)
    }()
    return fannedInStream
}
```

### `repeatFunc` Function

Repeats a function execution until a done signal is received.

```go
func repeatFunc[T any, K any](done <-chan K, fn func() T, ctx context.Context) <-chan T {
    stream := make(chan T)
    go func() {
        defer close(stream)
        for {
            select {
            case <-done:
                return
            case <-ctx.Done():
                return
            case stream <- fn():
            }
        }
    }()
    return stream
}


## Contributing

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-branch`).
5. Create a new Pull Request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contact

For any questions or suggestions, please open an issue or contact the repository owner.
```
