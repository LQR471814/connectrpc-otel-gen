package authv1connect

import (
	"context"

	v1 "services/auth/v1"

	connect "connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	authServiceTracer = otel.Tracer("services.auth.v1.AuthService")
)

type InstrumentedAuthServiceClient struct {
	inner AuthServiceClient
}

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
