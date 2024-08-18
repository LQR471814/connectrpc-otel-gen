package main

import (
	"fmt"
	"go/ast"
	"strings"
)

const importsTemplate = `import (
	"context"

	connect "connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
%s)`

const newTracerAndProviderFunc = `func newTracerAndProvider(serviceName string, exporter trace.SpanExporter, attrs []attribute.KeyValue) (*trace.TracerProvider, oteltrace.Tracer) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			append([]attribute.KeyValue{semconv.ServiceName(serviceName)}, attrs...)...,
		),
	)
	if err != nil {
		panic(err)
	}
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(r),
	)
	return traceProvider, traceProvider.Tracer(serviceName)
}`

const initTracerProvidersSignature = `// Initializes separate tracer providers for each service.
// Call ShutdownTracerProviders to call shutdown on all of them.
func InitTraceProviders(exporter trace.SpanExporter, attrs ...attribute.KeyValue) {`

const shutdownTracerProvidersPrefix = `// Shuts down all tracer providers initialized, it is
// a no-op if they have not been initialized.
func ShutdownTraceProviders(ctx context.Context) {
	if !providersInitialized {
		return
	}`
const shutdownTracerProvidersSuffix = `	providersInitialized = false
}`

type generateTarget struct {
	target                 *target
	tracerName             string
	instrumentedClientName string
}

func generate(file *ast.File, targets []*target) string {
	generateTargets := make([]generateTarget, len(targets))
	for i, t := range targets {
		generateTargets[i] = generateTarget{
			target:                 t,
			tracerName:             fmt.Sprintf("%sTracer", strings.ToLower(string(t.serviceName[0]))+t.serviceName[1:]),
			instrumentedClientName: fmt.Sprintf("Instrumented%s", t.clientIntfName),
		}
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("package %s\n\n", file.Name))

	var additionalImports strings.Builder
	for _, t := range targets {
		additionalImports.WriteString(fmt.Sprintf("\t%s %s\n", t.importAlias, t.importPath))
	}
	builder.WriteString(fmt.Sprintf(
		importsTemplate,
		additionalImports.String(),
	))

	builder.WriteString("\n\nvar (\n")
	for _, t := range generateTargets {
		builder.WriteString(fmt.Sprintf(
			"\t%s = otel.Tracer(\"%s\")\n",
			t.tracerName,
			t.target.fullServiceName,
		))
	}
	builder.WriteString(")\n\n")

	builder.WriteString("var (\n")
	for _, t := range generateTargets {
		builder.WriteString(fmt.Sprintf(
			"\t%sProvider *trace.TracerProvider\n",
			t.tracerName,
		))
	}
	builder.WriteString(")\n\n")

	builder.WriteString("var providersInitialized = false\n\n")

	builder.WriteString(newTracerAndProviderFunc + "\n\n")
	builder.WriteString(initTracerProvidersSignature + "\n")
	for _, t := range generateTargets {
		builder.WriteString(fmt.Sprintf(
			"\t%[1]sProvider, %[1]s = newTracerAndProvider(\"%[2]s\", exporter, attrs)\n",
			t.tracerName,
			t.target.fullServiceName,
		))
	}
	builder.WriteString("}\n\n")

	builder.WriteString(shutdownTracerProvidersPrefix + "\n")
	for _, t := range generateTargets {
		builder.WriteString(fmt.Sprintf(
			"\t%sProvider.Shutdown(ctx)\n",
			t.tracerName,
		))
	}
	builder.WriteString(shutdownTracerProvidersSuffix + "\n\n")

	for _, t := range generateTargets {
		t.write(&builder)
	}

	return builder.String()
}

const structTemplate = `// A wrapper around a value that implements "%[1]s"
// that adds telemetry.
type %[1]s struct {
	inner %[2]s
}`

// 1: Instrumented<Service>Client
// 2: <Service>Client
// 3: tracer initialization
// 4: full service name
const constructorTemplate = `// Creates a wrapper instance for telemetry around a value that implements
// "%[1]s"
func New%[1]s(inner %[2]s) %[1]s {
	return %[1]s{inner: inner}
}`

const methodTemplate = `func (c %[1]s) %[3]s(ctx context.Context, req *connect.Request[%[4]s]) (*connect.Response[%[5]s], error) {
	ctx, span := %[2]s.Start(ctx, "%[3]s")
	defer span.End()

	if span.IsRecording() {
		input, err := protojson.Marshal(req.Msg)
		if err == nil {
			span.SetAttributes(attribute.String("input", string(input)))
		} else {
			span.SetAttributes(attribute.String("input", "ERROR: FAILED TO SERIALIZE"))
			span.RecordError(err)
		}
	}

	res, err := c.inner.%[3]s(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if span.IsRecording() {
		output, err := protojson.Marshal(res.Msg)
		if err == nil {
			span.SetAttributes(attribute.String("output", string(output)))
		} else {
			span.SetAttributes(attribute.String("output", "ERROR: FAILED TO SERIALIZE"))
			span.RecordError(err)
		}
	}

	return res, nil
}`

func (gen generateTarget) write(out *strings.Builder) {
	out.WriteString(fmt.Sprintf(
		structTemplate,
		gen.instrumentedClientName,
		gen.target.clientIntfName,
	) + "\n\n")

	out.WriteString(fmt.Sprintf(
		constructorTemplate,
		gen.instrumentedClientName,
		gen.target.clientIntfName,
		"",
		gen.target.fullServiceName,
	) + "\n\n")

	for _, method := range gen.target.methods {
		out.WriteString(fmt.Sprintf(
			methodTemplate,
			gen.instrumentedClientName,
			gen.tracerName,
			method.name,
			method.requestType,
			method.responseType,
		) + "\n\n")
	}
}
