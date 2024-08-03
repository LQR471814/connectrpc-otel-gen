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
	"google.golang.org/protobuf/encoding/protojson"
%s)`

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

	for _, t := range generateTargets {
		t.write(&builder)
	}

	return builder.String()
}

const structTemplate = `type %s struct {
	inner %s
}`

const constructorTemplate = `func New%[1]s(inner %[2]s) %[1]s {
	return %[1]s{inner: inner}
}`

const methodTemplate = `func (c %[1]s) %[3]s(ctx context.Context, req *connect.Request[%[4]s]) (*connect.Response[%[5]s], error) {
	ctx, span := %[2]s.Start(ctx, "%[3]s")
	defer span.End()

	if span.IsRecording() {
		span.SetAttributes(attribute.String("procedure", req.Spec().Procedure))

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
