package connect_go_prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/easyCZ/connect-go-prometheus/gen/greet"
	"github.com/easyCZ/connect-go-prometheus/gen/greet/greetconnect"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

var (
	testMetricOptions = []MetricsOption{
		WithHistogram(true),
		WithByteMetrics(true),
		WithInflightMetrics(true),
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
		WithHistogramBuckets([]float64{1, 5}),
		WithByteMetrics(true),
	}
)

func createClientAndRequest(t *testing.T, srv *httptest.Server, interceptor *Interceptor) {
	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(interceptor))
	_, err := client.Greet(context.Background(), connect.NewRequest(&greet.GreetRequest{
		Name: "eliza",
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeOf(err), connect.CodeUnimplemented)
}

// TODO(prem): Verify metric counts after calling this.
func createClientAndStreamRequest(t *testing.T, srv *httptest.Server, interceptor *Interceptor) {
	client := greetconnect.NewGreetServiceClient(http.DefaultClient, srv.URL, connect.WithInterceptors(interceptor))
	stream, err := client.ServerStreamGreet(context.Background(), connect.NewRequest(&greet.GreetRequest{
		Name: "eliza",
	}))
	require.NoError(t, err)
	defer stream.Close()
	require.False(t, stream.Receive())
	err = stream.Err()
	require.Error(t, err)
	require.Equal(t, connect.CodeOf(err), connect.CodeUnimplemented)
}

func TestInterceptor_WithClient_WithServer_Histogram(t *testing.T) {
	reg := prom.NewRegistry()

	clientMetrics := NewClientMetrics(testMetricOptions...)
	serverMetrics := NewServerMetrics(testMetricOptions...)

	reg.MustRegister(clientMetrics, serverMetrics)

	interceptor := NewInterceptor(WithClientMetrics(clientMetrics), WithServerMetrics(serverMetrics))

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	createClientAndRequest(t, srv, interceptor)

	expectedMetrics := []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",
		"namespace_subsystem_connect_client_bytes_sent_total",
		"namespace_subsystem_connect_client_inflight_requests",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
		"namespace_subsystem_connect_server_bytes_received_total",
		"namespace_subsystem_connect_server_inflight_requests",
	}
	count, err := testutil.GatherAndCount(reg, expectedMetrics...)
	require.NoError(t, err)
	require.Equal(t, len(expectedMetrics), count)

	clientMetrics.Reset()
	serverMetrics.Reset()

	createClientAndStreamRequest(t, srv, interceptor)

	expectedMetrics = []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",
		"namespace_subsystem_connect_client_msg_sent_total",
		"namespace_subsystem_connect_client_bytes_sent_total",
		"namespace_subsystem_connect_client_inflight_requests",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
		"namespace_subsystem_connect_server_msg_received_total",
		"namespace_subsystem_connect_server_bytes_received_total",
		"namespace_subsystem_connect_server_inflight_requests",
	}
	count, err = testutil.GatherAndCount(reg, expectedMetrics...)
	require.NoError(t, err)
	require.Equal(t, len(expectedMetrics), count)
}

func TestInterceptor_Default(t *testing.T) {
	interceptor := NewInterceptor()

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	createClientAndRequest(t, srv, interceptor)

	expectedMetrics := []string{
		"connect_client_handled_total",
		"connect_client_started_total",

		"connect_server_handled_total",
		"connect_server_started_total",
	}
	count, err := testutil.GatherAndCount(prom.DefaultGatherer, expectedMetrics...)
	require.NoError(t, err)
	require.Equal(t, len(expectedMetrics), count)

	createClientAndStreamRequest(t, srv, interceptor)
}

func TestInterceptor_WithClientMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	clientMetrics := NewClientMetrics(testMetricOptions...)
	require.NoError(t, reg.Register(clientMetrics))

	interceptor := NewInterceptor(WithClientMetrics(clientMetrics), WithServerMetrics(nil))

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	createClientAndRequest(t, srv, interceptor)

	possibleMetrics := []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",
		"namespace_subsystem_connect_client_bytes_sent_total",
		"namespace_subsystem_connect_client_inflight_requests",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
		"namespace_subsystem_connect_server_bytes_received_total",
		"namespace_subsystem_connect_server_inflight_requests",
	}
	count, err := testutil.GatherAndCount(reg, possibleMetrics...)
	require.NoError(t, err)
	require.Equal(t, 5, count, "must report only client-side metrics, as server-side is disabled")

	clientMetrics.Reset()

	createClientAndStreamRequest(t, srv, interceptor)

	possibleMetrics = []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",
		"namespace_subsystem_connect_client_msg_sent_total",
		"namespace_subsystem_connect_client_bytes_sent_total",
		"namespace_subsystem_connect_client_inflight_requests",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
		"namespace_subsystem_connect_server_msg_received_total",
		"namespace_subsystem_connect_server_bytes_received_total",
		"namespace_subsystem_connect_server_inflight_requests",
	}
	count, err = testutil.GatherAndCount(reg, possibleMetrics...)
	require.NoError(t, err)
	require.Equal(t, 6, count, "must report only client-side metrics, as server-side is disabled")
}

func TestInterceptor_WithServerMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	serverMetrics := NewServerMetrics(testMetricOptions...)
	require.NoError(t, reg.Register(serverMetrics))

	interceptor := NewInterceptor(WithServerMetrics(serverMetrics), WithClientMetrics(nil))

	_, handler := greetconnect.NewGreetServiceHandler(greetconnect.UnimplementedGreetServiceHandler{}, connect.WithInterceptors(interceptor))
	srv := httptest.NewServer(handler)

	createClientAndRequest(t, srv, interceptor)

	possibleMetrics := []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",
		"namespace_subsystem_connect_client_bytes_sent_total",
		"namespace_subsystem_connect_client_inflight_requests",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
		"namespace_subsystem_connect_server_bytes_received_total",
		"namespace_subsystem_connect_server_inflight_requests",
	}
	count, err := testutil.GatherAndCount(reg, possibleMetrics...)
	require.NoError(t, err)
	require.Equal(t, 5, count, "must report only server-side metrics, client-side is disabled")

	serverMetrics.Reset()

	createClientAndStreamRequest(t, srv, interceptor)

	possibleMetrics = []string{
		"namespace_subsystem_connect_client_handled_seconds",
		"namespace_subsystem_connect_client_handled_total",
		"namespace_subsystem_connect_client_started_total",
		"namespace_subsystem_connect_client_msg_sent_total",
		"namespace_subsystem_connect_client_bytes_sent_total",
		"namespace_subsystem_connect_client_inflight_requests",

		"namespace_subsystem_connect_server_handled_seconds",
		"namespace_subsystem_connect_server_handled_total",
		"namespace_subsystem_connect_server_started_total",
		"namespace_subsystem_connect_server_msg_received_total",
		"namespace_subsystem_connect_server_bytes_received_total",
		"namespace_subsystem_connect_server_inflight_requests",
	}
	count, err = testutil.GatherAndCount(reg, possibleMetrics...)
	require.NoError(t, err)
	require.Equal(t, 6, count, "must report only server-side metrics, client-side is disabled")
}
