# connectrpc-otel-gen

> A code generator that adds opentelemetry instrumentation to Golang `connectrpc` services.

## Usage

```sh
go install github.com/LQR471814/connectrpc-otel-gen@latest

# calling `connectrpc-otel-gen` without arguments will make it accept
# the contents of api.connect.go in STDIN and print the generated output into STDOUT
cat some/go/code/here.go | connectrpc-otel-gen > output.go

# calling `connectrpc-otel-gen` with the paths of directories will cause it
# to recursively find `api.connect.go` files and generated `api.telemetry.go`
# files in the same directory
connectrpc-otel-gen . other_directory/

# input:
# - api.connect.go
# - other_directory/
#   - api.connect.go

# output:
# - api.connect.go
# - api.telemetry.go
# - other_directory/
#   - api.connect.go
#   - api.telemetry.go
```

Usage of the generated code is as follows.

```go
// some value that implements the servicev1connect.ServiceClient interface
service := &Service{}

// (optional) initialize a separate trace provider instead of the default (and resource)
// for each service contained within this package
// this can be useful if you need each service to have a different resource
servicev1connect.InitTraceProvider(
   traceExporter,
   attribute.String("additional_resource_attr", "..."),
)

wrapped := servicev1connect.NewInstrumentedServiceClient(service)
```

You can see a sample of the generated code [here](./example/sample_output.go), the original connectrpc code [here](./example/api.connect.go), and its corresponding proto definition [here](./example/api.proto).

## Why?

You may be wondering why this exists when there is an official solution for opentelemetry with connectrpc in Go the form [otelconnect](https://github.com/connectrpc/otelconnect-go).

The reason is because `otelconnect-go` only works when there is a layer of HTTP between the client and the server, meaning if you want to compose multiple gRPC services in a single server (like a monolith setup) you have to deal with HTTP overhead, both in performance and in boilerplate.

If you use `connectrpc-otel-gen` on the other hand, it will generate some wrapper code that you can wrap your gRPC server implementation with so even when calling the methods directly, you can still have instrumentation.

