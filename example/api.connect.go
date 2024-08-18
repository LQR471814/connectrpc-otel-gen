package authv1connect

import (
	"context"

	v1 "vcassist-backend/proto/vcassist/services/auth/v1"

	connect "connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	authServiceTracer = otel.Tracer("vcassist.services.auth.v1.AuthService")
)

func newTracerAndProvider(serviceName string, exporter trace.SpanExporter, attrs []attribute.KeyValue) oteltrace.Tracer {
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
	return traceProvider.Tracer(serviceName)
}

func InitTraceProvider(exporter trace.SpanExporter, attrs ...attribute.KeyValue) {
	authServiceTracer = newTracerAndProvider("vcassist.services.auth.v1.AuthService", exporter, attrs)
}

type InstrumentedAuthServiceClient struct {
	inner AuthServiceClient
}

// Creates a wrapper instance around a value that implements
// InstrumentedAuthServiceClient
func NewInstrumentedAuthServiceClient(inner AuthServiceClient) InstrumentedAuthServiceClient {
	return InstrumentedAuthServiceClient{inner: inner}
}

func (c InstrumentedAuthServiceClient) StartLogin(ctx context.Context, req *connect.Request[v1.StartLoginRequest]) (*connect.Response[v1.StartLoginResponse], error) {
	ctx, span := authServiceTracer.Start(ctx, "StartLogin")
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

	res, err := c.inner.StartLogin(ctx, req)
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
}

func (c InstrumentedAuthServiceClient) ConsumeVerificationCode(ctx context.Context, req *connect.Request[v1.ConsumeVerificationCodeRequest]) (*connect.Response[v1.ConsumeVerificationCodeResponse], error) {
	ctx, span := authServiceTracer.Start(ctx, "ConsumeVerificationCode")
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

	res, err := c.inner.ConsumeVerificationCode(ctx, req)
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
}

func (c InstrumentedAuthServiceClient) VerifyToken(ctx context.Context, req *connect.Request[v1.VerifyTokenRequest]) (*connect.Response[v1.VerifyTokenResponse], error) {
	ctx, span := authServiceTracer.Start(ctx, "VerifyToken")
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

	res, err := c.inner.VerifyToken(ctx, req)
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
}
