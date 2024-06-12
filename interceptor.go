package connect_go_prometheus

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/cockroachdb/errors"
	prom "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
)

const (
	CodeOk = "ok"
)

func NewInterceptor(opts ...InterceptorOption) *Interceptor {
	options := evaluteInterceptorOptions(&interceptorOptions{
		client: DefaultClientMetrics,
		server: DefaultServerMetrics,
	}, opts...)

	return &Interceptor{
		client: options.client,
		server: options.server,
	}
}

var _ connect.Interceptor = (*Interceptor)(nil)

type Interceptor struct {
	client *Metrics
	server *Metrics
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Short-circuit, not configured to report for either client or server.
		if i.client == nil && i.server == nil {
			return next(ctx, req)
		}

		now := time.Now()
		callType := streamTypeString(req.Spec().StreamType)
		callPackage, callMethod := procedureToPackageAndMethod(req.Spec().Procedure)

		var reporter *Metrics
		if req.Spec().IsClient {
			reporter = i.client
		} else {
			reporter = i.server
		}

		var code string
		if reporter != nil {
			var bytes *prom.CounterVec
			if reporter.isClient {
				bytes = reporter.bytesSent
			} else {
				bytes = reporter.bytesReceived
			}
			if bytes != nil {
				bytes.WithLabelValues(callType, callPackage, callMethod).Add(float64(proto.Size(req.Any().(proto.Message))))
			}
			reporter.ReportStarted(callType, callPackage, callMethod)
			defer func() {
				reporter.ReportHandled(callType, callPackage, callMethod, code)
				reporter.ReportHandledSeconds(callType, callPackage, callMethod, code, time.Since(now).Seconds())
			}()
		}

		resp, err := next(ctx, req)
		code = codeOf(err)
		if err == nil && reporter != nil {
			var bytes *prom.CounterVec
			if reporter.isClient {
				bytes = reporter.bytesReceived
			} else {
				bytes = reporter.bytesSent
			}
			if bytes != nil {
				bytes.WithLabelValues(callType, callPackage, callMethod).Add(float64(proto.Size(resp.Any().(proto.Message))))
			}
		}

		return resp, err
	})
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// Short-circuit, not configured to report for client.
		if i.client == nil {
			return next(ctx, spec)
		}

		now := time.Now()
		callType := streamTypeString(spec.StreamType)
		callPackage, callMethod := procedureToPackageAndMethod(spec.Procedure)

		i.client.ReportStarted(callType, callPackage, callMethod)
		onClose := func(err error) {
			code := codeOf(err)
			i.client.ReportHandled(callType, callPackage, callMethod, code)
			i.client.ReportHandledSeconds(callType, callPackage, callMethod, code, time.Since(now).Seconds())
		}

		conn := next(ctx, spec)
		return newStreamingClientConn(conn, i, onClose)
	})
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		// Short-circuit, not configured to report for server.
		if i.server == nil {
			return next(ctx, shc)
		}

		now := time.Now()
		callType := streamTypeString(shc.Spec().StreamType)
		callPackage, callMethod := procedureToPackageAndMethod(shc.Spec().Procedure)

		var code string
		i.server.ReportStarted(callType, callPackage, callMethod)
		defer func() {
			i.server.ReportHandled(callType, callPackage, callMethod, code)
			i.server.ReportHandledSeconds(callType, callPackage, callMethod, code, time.Since(now).Seconds())
		}()

		shc = newStreamingHandlerConn(shc, i)
		err := next(ctx, shc)
		code = codeOf(err)
		return err
	})
}

func procedureToPackageAndMethod(procedure string) (string, string) {
	procedure = strings.TrimPrefix(procedure, "/") // remove leading slash
	if i := strings.Index(procedure, "/"); i >= 0 {
		return procedure[:i], procedure[i+1:]
	}

	return "unknown", "unknown"
}

func streamTypeString(st connect.StreamType) string {
	switch st {
	case connect.StreamTypeUnary:
		return "unary"
	case connect.StreamTypeClient:
		return "client_stream"
	case connect.StreamTypeServer:
		return "server_stream"
	case connect.StreamTypeBidi:
		return "bidi"
	default:
		return "unknown"
	}
}

func codeOf(err error) string {
	if err == nil {
		return CodeOk
	}
	code := connect.CodeOf(err)
	if code == connect.CodeUnknown {
		if errors.Is(err, context.Canceled) {
			code = connect.CodeCanceled
		} else if errors.Is(err, context.DeadlineExceeded) {
			code = connect.CodeDeadlineExceeded
		}
	}
	return code.String()
}

type interceptorOptions struct {
	client *Metrics
	server *Metrics
}

type InterceptorOption func(*interceptorOptions)

func WithClientMetrics(m *Metrics) InterceptorOption {
	return func(io *interceptorOptions) {
		io.client = m
	}
}

func WithServerMetrics(m *Metrics) InterceptorOption {
	return func(io *interceptorOptions) {
		io.server = m
	}
}

func evaluteInterceptorOptions(defaults *interceptorOptions, opts ...InterceptorOption) *interceptorOptions {
	for _, opt := range opts {
		opt(defaults)
	}
	return defaults
}
