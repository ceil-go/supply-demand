# supply-demand

A small Go library for building demand-driven supplier workflows.

## What is it?

This package lets you organize logic as a set of "suppliers." Each supplier can respond to "demand" for a certain type of result. You can chain suppliers, change the available set of suppliers per request, and handle asynchronous needs with Go routines and channels.

## Example

```go
import (
    "github.com/ceil-go/supply-demand"
    "fmt"
    "time"
)

func main() {
    suppliers := map[string]supply_demand.Supplier{
        "first": func(data any, scope supply_demand.Scope) chan any {
            result := make(chan any)
            go func() {
                time.Sleep(1 * time.Second)
                result <- "Value from first"
                close(result)
            }()
            return result
        },
    }

    root := func(data any, scope supply_demand.Scope) chan any {
        result := make(chan any)
        go func() {
            res := <-scope.Demand(supply_demand.ScopedDemandProps{Type: "first"})
            fmt.Println(res)
            result <- nil
            close(result)
        }()
        return result
    }

    <-supply_demand.SupplyDemand(root, suppliers)
}
```

## Features

- Write supplier functions that produce values on demand (possibly asynchronously)
- Each demand can request a specific type and data payload
- Suppliers can invoke further demands and even modify what suppliers are available to sub-calls
- Useful for trees of requests, plugin systems, dependency graphs, and more

## Status

- Not heavily documented
- For adventurous developers and prototyping
- API may change

## Installation

```sh
go get github.com/ceil-go/supply-demand
```

## License

MIT License.  
See [LICENSE](LICENSE) file for details.
