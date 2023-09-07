package tracing

// refer https://github.com/sunmi-OS/gocore/blob/master/istio/trace.go

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

const (
	X_REQUEST_ID      = "x-request-id"
	X_B3_TRACEID      = "x-b3-traceid"
	X_B3_SPANID       = "x-b3-spanid"
	X_B3_PARENTSPANID = "x-b3-parentspanid"
	X_B3_SAMPLED      = "x-b3-sampled"
	X_B3_FLAGS        = "x-b3-flags"
	X_OT_SPAN_CONTEXT = "x-ot-span-context"
)

type TraceHeader struct {
	HttpHeader   http.Header
	GrpcMetadata metadata.MD
}

func NewGrpcContextFromGin(ctx *gin.Context) context.Context {
	traceHeader := LoadFromHttpHeader(ctx.Request.Header)
	grpcCtx := metadata.NewOutgoingContext(ctx, traceHeader.GrpcMetadata)
	return grpcCtx
}

func NewGrpcContextFromGrpc(ctx context.Context) context.Context {
	grpcMetadata, _ := metadata.FromIncomingContext(ctx)
	traceHeader := LoadFromGrpcMetadata(grpcMetadata)
	grpcCtx := metadata.NewOutgoingContext(ctx, traceHeader.GrpcMetadata)
	return grpcCtx
}

func LoadFromHttpHeader(headers http.Header) TraceHeader {

	md := map[string]string{}
	httpHeaders := http.Header{}

	fillFromHttpHeader(headers, X_REQUEST_ID, md, httpHeaders)
	fillFromHttpHeader(headers, X_B3_TRACEID, md, httpHeaders)
	fillFromHttpHeader(headers, X_B3_SPANID, md, httpHeaders)
	fillFromHttpHeader(headers, X_B3_PARENTSPANID, md, httpHeaders)
	fillFromHttpHeader(headers, X_B3_SAMPLED, md, httpHeaders)
	fillFromHttpHeader(headers, X_B3_FLAGS, md, httpHeaders)
	fillFromHttpHeader(headers, X_OT_SPAN_CONTEXT, md, httpHeaders)

	return TraceHeader{
		HttpHeader:   httpHeaders,
		GrpcMetadata: metadata.New(md),
	}
}

func LoadFromGrpcMetadata(rawMetadata metadata.MD) TraceHeader {

	md := map[string]string{}
	httpHeaders := http.Header{}

	fillFromGrpcMetadata(rawMetadata, X_REQUEST_ID, md, httpHeaders)
	fillFromGrpcMetadata(rawMetadata, X_B3_TRACEID, md, httpHeaders)
	fillFromGrpcMetadata(rawMetadata, X_B3_SPANID, md, httpHeaders)
	fillFromGrpcMetadata(rawMetadata, X_B3_PARENTSPANID, md, httpHeaders)
	fillFromGrpcMetadata(rawMetadata, X_B3_SAMPLED, md, httpHeaders)
	fillFromGrpcMetadata(rawMetadata, X_B3_FLAGS, md, httpHeaders)
	fillFromGrpcMetadata(rawMetadata, X_OT_SPAN_CONTEXT, md, httpHeaders)

	return TraceHeader{
		HttpHeader:   httpHeaders,
		GrpcMetadata: metadata.New(md),
	}
}

func fillFromHttpHeader(headers http.Header, key string, md map[string]string, httpHeaders http.Header) {
	value := headers.Get(key)
	if value != "" {
		md[key] = value
		httpHeaders.Add(key, value)
	}
}

func fillFromGrpcMetadata(rawMetadata metadata.MD, key string, md map[string]string, httpHeaders http.Header) {
	value := rawMetadata.Get(key)
	if len(value) > 0 {
		md[key] = value[0]
		httpHeaders.Add(key, value[0])
	}
}
