# Flahmingo Golang Helpers

## Sentry gRPC integration

Setup sentry with the following variables, they are optional, but recommended:

Option  | Type | Purpose
------------- | ------------- | -------------
DSN: | string | Sentry key for service. Found in Sentry dashboard for the specific project (i.e backend)
Environment: | string | As defined in the config file will tell Sentry how to segregate logs
AttachStacktrace: | bool | Release: branch and sha passed in from build
Release: | string | The dist to be sent with events
Debug: | bool | Configures whether SDK should generate and attach stacktraces to pure capture message calls.

**_More options found in [sentry documentation](https://docs.sentry.io/platforms/go/configuration/)_**

### Implementation Code:

There are two vital pieces to implement and start the Sentry service:

1. This code is to live in the app start or main file

```go
// These vars to be overriden by the compiler
var (
  sha    = "unknown"
  branch = "unknown"
  // swap for service name
  serviceName = "name_your_service"
)

// Setup sentry
err = sentry.Init(sentry.ClientOptions{
  Dsn:              config.Sentry.DSN,
  Environment:      config.Env,
  AttachStacktrace: true,
  Release:          fmt.Sprintf("%s@%s", branch, sha),
  Debug:            config.Debug,
})

if err != nil {
  // please note requires flog package or in its abscence another package
  flog.Fatalf("sentry.Init: %v", err)
}

// Sets the tag so that the specific service can be traced in the dashboard 
sentry.ConfigureScope(func(scope *sentry.Scope) {
  scope.SetTag("service_name", serviceName)
})

// defer flush on shutdown by two seconds as per documentation (lint no magic number)
sentryFlushTimeout := 2 * time.Second

// We have to wait before quitting so, sentry push all the events.
defer sentry.Flush(sentryFlushTimeout)
```
2. Implementation with Unary or Stream Server attach to gRPC middlewares

```go
// gRPC middlewares
grpcOpts := []grpc.ServerOption{
  grpc.StreamInterceptor(
    middleware.ChainStreamServer(
      grpczap.StreamServerInterceptor(logger, zapOpts...),
      // your interceptors
      sentrygrpc.SentryStreamServerInterceptor(),
    ),
  ),
  grpc.UnaryInterceptor(
    middleware.ChainUnaryServer(
      tags.UnaryServerInterceptor(),
      // your interceptors
      grpczap.UnaryServerInterceptor(logger, zapOpts...),
      sentrygrpc.SentryUnaryServerInterceptor(unarySentryOptions...),
    ),
  ),
}

grpcServer := grpc.NewServer(grpcOpts...)

```