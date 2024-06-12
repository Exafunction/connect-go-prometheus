package connect_go_prometheus

import (
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/easyCZ/connect-go-prometheus/gen/greet/greetconnect"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestServerMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	sm := NewServerMetrics(
		WithHistogram(true),
		WithByteMetrics(true),
		WithInflightMetrics(true),
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
		WithHistogramBuckets([]float64{0.5, 1, 1.5}),
	)
	require.NoError(t, reg.Register(sm))

	started := sm.requestStarted.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	started.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(started))

	handled := sm.requestHandled.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
	handled.Inc()
	require.EqualValues(t, 1, testutil.ToFloat64(handled))

	msgSent := sm.streamMsgSent.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	msgSent.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(msgSent))

	msgReceived := sm.streamMsgReceived.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	msgReceived.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(msgReceived))

	require.NotNil(t, sm.bytesSent)
	bytesSent := sm.bytesSent.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	bytesSent.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(bytesSent))

	require.NotNil(t, sm.bytesReceived)
	bytesReceived := sm.bytesReceived.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	bytesReceived.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(bytesReceived))

	require.NotNil(t, sm.inflightRequests)
	inflight := sm.inflightRequests.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	inflight.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(inflight))

	require.NotNil(t, sm.requestHandledSeconds)
	handledHist := sm.requestHandledSeconds.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
	handledHist.Observe(1)
	err := testutil.CollectAndCompare(sm.requestHandledSeconds, strings.NewReader(`
			# HELP namespace_subsystem_connect_server_handled_seconds Histogram of RPCs handled server-side
			# TYPE namespace_subsystem_connect_server_handled_seconds histogram
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="0.5"} 0
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1"} 1
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1.5"} 1
			namespace_subsystem_connect_server_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="Inf"} 1
			namespace_subsystem_connect_server_handled_seconds_sum{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
			namespace_subsystem_connect_server_handled_seconds_count{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
		`))
	require.NoError(t, err)
}

func TestClientMetrics(t *testing.T) {
	reg := prom.NewRegistry()
	cm := NewClientMetrics(
		WithHistogram(true),
		WithByteMetrics(true),
		WithInflightMetrics(true),
		WithNamespace("namespace"),
		WithSubsystem("subsystem"),
		WithConstLabels(prom.Labels{"component": "foo"}),
		WithHistogramBuckets([]float64{0.5, 1, 1.5}),
	)
	require.NoError(t, reg.Register(cm))

	started := cm.requestStarted.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	started.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(started))

	msgSent := cm.streamMsgSent.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	msgSent.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(msgSent))

	msgreceived := cm.streamMsgReceived.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	msgreceived.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(msgreceived))

	handled := cm.requestHandled.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
	handled.Inc()
	require.EqualValues(t, 1, testutil.ToFloat64(handled))

	require.NotNil(t, cm.bytesSent)
	bytesSent := cm.bytesSent.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	bytesSent.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(bytesSent))

	require.NotNil(t, cm.bytesReceived)
	bytesReceived := cm.bytesReceived.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	bytesReceived.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(bytesReceived))

	require.NotNil(t, cm.inflightRequests)
	inflight := cm.inflightRequests.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet")
	inflight.Inc()
	require.EqualValues(t, float64(1), testutil.ToFloat64(inflight))

	require.NotNil(t, cm.requestHandledSeconds)
	handledHist := cm.requestHandledSeconds.WithLabelValues("unary", greetconnect.GreetServiceName, "Greet", connect.CodeAborted.String())
	handledHist.Observe(1)
	err := testutil.CollectAndCompare(cm.requestHandledSeconds, strings.NewReader(`
		# HELP namespace_subsystem_connect_client_handled_seconds Histogram of RPCs handled client-side
		# TYPE namespace_subsystem_connect_client_handled_seconds histogram
		namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="0.5"} 0
		namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1"} 1
		namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="1.5"} 1
		namespace_subsystem_connect_client_handled_seconds_bucket{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary",le="Inf"} 1
		namespace_subsystem_connect_client_handled_seconds_sum{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
		namespace_subsystem_connect_client_handled_seconds_count{code="aborted",component="foo",method="Greet",service="greet.v1.GreetService",type="unary"} 1
	`))
	require.NoError(t, err)
}
