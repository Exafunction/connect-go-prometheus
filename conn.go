package connect_go_prometheus

import (
	"time"

	"google.golang.org/protobuf/proto"

	"connectrpc.com/connect"
)

type streamingConn struct {
	startTime                 time.Time
	callType, service, method string
	reporter                  *Metrics
}

func newStreamingConn(spec connect.Spec, reporter *Metrics) streamingConn {
	callPackage, callMethod := procedureToPackageAndMethod(spec.Procedure)
	conn := streamingConn{
		startTime: time.Now(),
		callType:  streamTypeString(spec.StreamType),
		service:   callPackage,
		method:    callMethod,
		reporter:  reporter,
	}
	reporter.requestStarted.WithLabelValues(conn.callType, conn.service, conn.method).Inc()
	return conn
}

func (conn *streamingConn) reportSend(message any) {
	conn.reporter.streamMsgSent.WithLabelValues(conn.callType, conn.service, conn.method).Inc()
	if conn.reporter.bytesSent != nil {
		conn.reporter.bytesSent.WithLabelValues(conn.callType, conn.service, conn.method).Add(float64(proto.Size(message.(proto.Message))))
	}
}

func (conn *streamingConn) reportReceive(message any) {
	conn.reporter.streamMsgReceived.WithLabelValues(conn.callType, conn.service, conn.method).Inc()
	if conn.reporter.bytesReceived != nil {
		conn.reporter.bytesReceived.WithLabelValues(conn.callType, conn.service, conn.method).Add(float64(proto.Size(message.(proto.Message))))
	}
}

type streamingClientConn struct {
	connect.StreamingClientConn
	streamingConn
	onClose func(error)
}

func newStreamingClientConn(conn connect.StreamingClientConn, i *Interceptor, onClose func(error)) *streamingClientConn {
	return &streamingClientConn{
		StreamingClientConn: conn,
		streamingConn:       newStreamingConn(conn.Spec(), i.client),
		onClose:             onClose,
	}
}

func (conn *streamingClientConn) Send(msg any) error {
	conn.reportSend(msg)
	return conn.StreamingClientConn.Send(msg)
}

func (conn *streamingClientConn) Receive(msg any) error {
	err := conn.StreamingClientConn.Receive(msg)
	if err == nil {
		conn.reportReceive(msg)
	}
	return err
}

func (conn *streamingClientConn) CloseResponse() error {
	err := conn.StreamingClientConn.CloseResponse()
	conn.onClose(err)
	return err
}

var _ connect.StreamingClientConn = (*streamingClientConn)(nil)

type streamingHandlerConn struct {
	connect.StreamingHandlerConn
	streamingConn
}

func newStreamingHandlerConn(conn connect.StreamingHandlerConn, i *Interceptor) *streamingHandlerConn {
	return &streamingHandlerConn{
		StreamingHandlerConn: conn,
		streamingConn:        newStreamingConn(conn.Spec(), i.server),
	}
}

func (conn *streamingHandlerConn) Send(msg any) error {
	conn.reportSend(msg)
	return conn.StreamingHandlerConn.Send(msg)
}

func (conn *streamingHandlerConn) Receive(msg any) error {
	err := conn.StreamingHandlerConn.Receive(msg)
	if err == nil {
		conn.reportReceive(msg)
	}
	return err
}

var _ connect.StreamingHandlerConn = (*streamingHandlerConn)(nil)
