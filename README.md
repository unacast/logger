# unalogger

![logger](logger.gif)

Opinionated [logxi](https://github.com/mgutz/logxi) based logging with the option of passing errors on to Sentry 

----

## Getting started
### Install

To install `logger` pull it with `go get` like this

`go get github.com/unacast/logger`

or if you use `dep` just do a
```bash
dep ensure -add github.com/unacast/logger
```
oooor if you use `glide` just do a 

```bash
glide get github.com/unacast/logger
glide install
```

### Usage expamles

#### Regular stdout logging

```go
package main
import (
    "github.com/unacast/logger"
    "errors"
)

var log = logger.New("main")

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

#### Logging to custom logfile, e.g. for apps running on Appengine flex
It might be a good idea to create a local logger package to do the setup with the desired filename etc.

```go
package logger
import (
    "github.com/unacast/logger"
)

const logFileEnv = "UC_[APP NAME HERE]_LOG_FILE"

var (
	logFileName = os.Getenv(logFileEnv)
)

func New(name string) unalogger.UnaLogger {
	return unalogger.NewLogger(
        unalogger.Config{
            Name: name, 
            FileName: logFileName
        }
     )
}
```

#### Use the Stackdriver Error Reporting functionality
`unalogger` has an integration to optionally handle panics and report them as errors to *Stackdriver Error Reporting*.
Initialise the `errorClient` in the start of your program, typically in `main.go`.
```go
func main() {

	consts := constants.Consts

	// Enable Stackdriver Errorreporting
	errorClient, recoverPanics := logger.SetUpErrorReporting(context.Background(), consts.Project, "shipit", consts.GitSha)
    // Defering this function handles panics in this scope/goroutine
	defer recoverPanics()
    // Strictly not necessary, 
	defer errorClient.Close()
    ...
}
```
If you have several sub-goroutines in your app, you have to pass the `recoverPanics` into them and `defer recoverPanics` locally to handle panics there.

### Loggers

Logxi defaults to using a json formatter in production and a typical log line looks like this
```bash
{"_t":"2017-08-25T15:04:44+0200", "_p":"20698", "severity":"INFO", "_n":"main", "message":"The service has successfully launched"}
```
Some of the keys has been renamed to confrom with the Google Cloud Logging format

Locally logxi uses a formatter that outputs lines like this
```bash
15:07:59.502228 INFO goa mount ctrl: K8sStatuses action: Liveness
   route: GET /liveness
```

## Maintainers 
 - @torbjornvatn 
