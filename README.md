# logger
Opinionated [logxi](https://github.com/mgutz/logxi) based logging with the option of passing errors on to Sentry 

----

## Getting started
### Install

To install `logger` pull it with `go get` like this

`go get github.com/unacast/logger`

or if you use `glide` just do a 
```bash
glide get github.com/unacast/logger
glide install
```

### Usage expamles

_Without Sentry_

```go
package main
import (
    "github.com/unacast/logger"
    "errors"
)

var log = logger.NewLogger("main")

func main() {

    // Info takes a message and a var args list of key-value pairs that are 
    log.Info(
        "Log message",
        "jobID", "1312313", "count", 1000,
    )
    
    // Error takes an error as the second argument and it's added to 
    // the list of key-value pairs with "error" as the key 
    err := errors.New("This is an error!")
    log.Info(
        "Error log message",
        err,
        "jobID", "1312313", "count", 1000,
    )
}
```

_With Sentry_

```go
package main
import (
    "github.com/unacast/logger"
    "github.com/getsentry/raven-go"
    "errors"
)

// Notice the call to NewSentryLogger here
var log = logger.NewSentryLogger("main")

func main() {

    // If you're going to use the Sentry enabled logger
    // remember to set these  
    raven.SetDSN("[SENTRY DSN GOES HERE]")
    raven.SetRelease("[GIT SHA extracted programatically]")
    raven.SetEnvironment("[ENVIRONMENT e.g from an environment variable]")
    
    // Error takes an error as the second argument and it's added to 
    // the list of key-value pairs with "error" as the key 
    err := errors.New("This is an error!")
    log.Info(
        "Error log message",
        err,
        "jobID", "1312313", "count", 1000,
    )
}
```

### Loggers

Logxi defaults to using a json formatter in produection and the a typical log line looks like this
`løkasd`

## Maintainers 
 - @torbjornvatn 