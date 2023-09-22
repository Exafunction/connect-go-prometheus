package connect_go_prometheus

import (
	"google.golang.org/protobuf/proto"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	DefaultClientMetrics = NewClientMetrics()
	DefaultServerMetrics = NewServerMetrics()
)

func init() {
	// Register default metrics against default prometheus metrics registry.
	prom.MustRegister(DefaultServerMetrics)
	prom.MustRegister(DefaultClientMetrics)
}

// NewServerMetrics creates new Connect metrics for server-side handling.
func NewServerMetrics(opts ...MetricsOption) *Metrics {
	config := evaluateMetricsOptions(&metricsOptions{
		histogramBuckets:          prom.DefBuckets,
		requestStartedName:        "connect_server_started_total",
		requestHandledName:        "connect_server_handled_total",
		requestHandledSecondsName: "connect_server_handled_seconds",
		streamMsgSentName:         "connect_server_msg_sent_total",
		streamMsgReceivedName:     "connect_server_msg_received_total",
		bytesSentName:             "connect_server_bytes_sent_total",
		bytesReceivedName:         "connect_server_bytes_received_total",
	}, opts...)

	m := &Metrics{
		isClient: false,
		requestStarted: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestStartedName,
			Help:        "Total number of RPCs started handling server-side",
		}, []string{"type", "service", "method"}),
		requestHandled: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledName,
			Help:        "Total number of RPCs handled server-side",
		}, []string{"type", "service", "method", "code"}),
		streamMsgSent: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.streamMsgSentName,
			Help:        "Total number of stream messages sent by server-side",
		}, []string{"type", "service", "method"}),
		streamMsgReceived: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.streamMsgReceivedName,
			Help:        "Total number of stream messages received by server-side",
		}, []string{"type", "service", "method"}),
	}

	if config.withHistogram {
		m.requestHandledSeconds = prom.NewHistogramVec(prom.HistogramOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledSecondsName,
			Help:        "Histogram of RPCs handled server-side",
			Buckets:     config.histogramBuckets,
		}, []string{"type", "service", "method", "code"})
	}

	if config.withByteMetrics {
		m.bytesSent = prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.bytesSentName,
			Help:        "Total number of bytes sent by server-side",
		}, []string{"type", "service", "method"})
		m.bytesReceived = prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.bytesReceivedName,
			Help:        "Total number of bytes received by server-side",
		}, []string{"type", "service", "method"})
	}

	return m
}

func NewClientMetrics(opts ...MetricsOption) *Metrics {
	config := evaluateMetricsOptions(&metricsOptions{
		histogramBuckets:          prom.DefBuckets,
		requestStartedName:        "connect_client_started_total",
		requestHandledName:        "connect_client_handled_total",
		requestHandledSecondsName: "connect_client_handled_seconds",
		streamMsgSentName:         "connect_client_msg_sent_total",
		streamMsgReceivedName:     "connect_client_msg_recieved_total",
		bytesSentName:             "connect_client_bytes_sent_total",
		bytesReceivedName:         "connect_client_bytes_received_total",
	}, opts...)

	m := &Metrics{
		isClient: true,
		requestStarted: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestStartedName,
			Help:        "Total number of RPCs started handling client-side",
		}, []string{"type", "service", "method"}),
		requestHandled: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledName,
			Help:        "Total number of RPCs handled client-side",
		}, []string{"type", "service", "method", "code"}),
		streamMsgSent: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.streamMsgSentName,
			Help:        "Total number of stream messages sent by client-side",
		}, []string{"type", "service", "method"}),
		streamMsgReceived: prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.streamMsgReceivedName,
			Help:        "Total number of stream messages received by client-side",
		}, []string{"type", "service", "method"}),
	}

	if config.withHistogram {
		m.requestHandledSeconds = prom.NewHistogramVec(prom.HistogramOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.requestHandledSecondsName,
			Help:        "Histogram of RPCs handled client-side",
			Buckets:     config.histogramBuckets,
		}, []string{"type", "service", "method", "code"})
	}

	if config.withByteMetrics {
		m.bytesSent = prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.bytesSentName,
			Help:        "Total number of bytes sent by client-side",
		}, []string{"type", "service", "method"})
		m.bytesReceived = prom.NewCounterVec(prom.CounterOpts{
			Namespace:   config.namespace,
			Subsystem:   config.subsystem,
			ConstLabels: config.constLabels,
			Name:        config.bytesReceivedName,
			Help:        "Total number of bytes received by client-side",
		}, []string{"type", "service", "method"})
	}

	return m
}

var _ prom.Collector = (*Metrics)(nil)

type Metrics struct {
	isClient              bool
	requestStarted        *prom.CounterVec
	requestHandled        *prom.CounterVec
	requestHandledSeconds *prom.HistogramVec
	streamMsgSent         *prom.CounterVec
	streamMsgReceived     *prom.CounterVec
	bytesSent             *prom.CounterVec
	bytesReceived         *prom.CounterVec
}

// Describe implements Describe as required by prom.Collector
func (m *Metrics) Describe(c chan<- *prom.Desc) {
	m.requestStarted.Describe(c)
	m.requestHandled.Describe(c)
	if m.requestHandledSeconds != nil {
		m.requestHandledSeconds.Describe(c)
	}
	m.streamMsgSent.Describe(c)
	m.streamMsgReceived.Describe(c)
}

// Collect implements collect as required by prom.Collector
func (m *Metrics) Collect(c chan<- prom.Metric) {
	m.requestStarted.Collect(c)
	m.requestHandled.Collect(c)
	if m.requestHandledSeconds != nil {
		m.requestHandledSeconds.Collect(c)
	}
	m.streamMsgSent.Collect(c)
	m.streamMsgReceived.Collect(c)
}

func (m *Metrics) ReportStarted(callType, service, method string, message any) {
	m.requestStarted.WithLabelValues(callType, service, method).Inc()
	var streamMsg *prom.CounterVec
	var bytes *prom.CounterVec
	if m.isClient {
		streamMsg = m.streamMsgSent
		bytes = m.bytesSent
	} else {
		streamMsg = m.streamMsgReceived
		bytes = m.bytesReceived
	}
	streamMsg.WithLabelValues(callType, service, method).Inc()
	if bytes != nil {
		bytes.WithLabelValues(callType, service, method).Add(float64(proto.Size(message.(proto.Message))))
	}
}

func (m *Metrics) ReportHandled(callType, service, method, code string, message any) {
	m.requestHandled.WithLabelValues(callType, service, method, code).Inc()
	if code == CodeOk {
		var streamMsg *prom.CounterVec
		var bytes *prom.CounterVec
		if m.isClient {
			streamMsg = m.streamMsgReceived
			bytes = m.bytesReceived
		} else {
			streamMsg = m.streamMsgSent
			bytes = m.bytesSent
		}
		streamMsg.WithLabelValues(callType, service, method).Inc()
		if bytes != nil {
			bytes.WithLabelValues(callType, service, method).Add(float64(proto.Size(message.(proto.Message))))
		}
	}
}

func (m *Metrics) ReportHandledSeconds(callType, service, method, code string, val float64) {
	if m.requestHandledSeconds != nil {
		m.requestHandledSeconds.WithLabelValues(callType, service, method, code).Observe(val)
	}
}

type metricsOptions struct {
	withHistogram    bool
	histogramBuckets []float64

	namespace string
	subsystem string

	requestStartedName        string
	requestHandledName        string
	requestHandledSecondsName string
	streamMsgSentName         string
	streamMsgReceivedName     string
	bytesSentName             string
	bytesReceivedName         string

	constLabels prom.Labels

	withByteMetrics bool
}

type MetricsOption func(opts *metricsOptions)

func WithHistogram(enabled bool) MetricsOption {
	return func(opts *metricsOptions) {
		opts.withHistogram = enabled
	}
}

func WithHistogramBuckets(buckets []float64) MetricsOption {
	return func(opts *metricsOptions) {
		opts.histogramBuckets = buckets
	}
}

func WithNamespace(namespace string) MetricsOption {
	return func(opts *metricsOptions) {
		opts.namespace = namespace
	}
}

func WithSubsystem(subsystem string) MetricsOption {
	return func(opts *metricsOptions) {
		opts.subsystem = subsystem
	}
}

func WithConstLabels(labels prom.Labels) MetricsOption {
	return func(opts *metricsOptions) {
		opts.constLabels = labels
	}
}

func WithByteMetrics(enabled bool) MetricsOption {
	return func(opts *metricsOptions) {
		opts.withByteMetrics = enabled
	}
}

func evaluateMetricsOptions(defaults *metricsOptions, opts ...MetricsOption) *metricsOptions {
	for _, opt := range opts {
		opt(defaults)
	}

	return defaults
}
