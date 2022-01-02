// Package httpmw contains http middlewares.
package httpmw

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// _buffer is pool of bytes buffer.
var _buffer = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// allocate the space only once.
var _space = []byte(" ")

// compile time check.
var _ zapcore.ObjectMarshaler = (*HTTPPayload)(nil)

// HTTPPayload is the complete payload that can be logged.
type HTTPPayload struct {
	// The request method. Examples: "GET", "HEAD", "PUT", "POST".
	RequestMethod string `json:"requestMethod"`

	// The scheme (http, https), the host name, the path and the query portion of
	// the URL that was requested.
	//
	// Example: "http://example.com/some/info?color=red".
	RequestURL string `json:"requestUrl"`

	// The response code indicating the status of response.
	//
	// Examples: 200, 404.
	Status int `json:"status"`

	// The user agent sent by the client.
	//
	// Example: "Mozilla/4.0 (compatible; MSIE 6.0; Windows 98; Q312461; .NET CLR 1.0.3705)".
	UserAgent string `json:"userAgent"`

	// The IP address (IPv4 or IPv6) of the client that issued the HTTP request.
	//
	// Examples: "192.168.1.1", "FE80::0202:B3FF:FE1E:8329".
	RemoteIP string `json:"remoteIp"`

	// The IP address (IPv4 or IPv6) for whom the request is forwarded for.
	//
	// Examples: "192.168.1.1", "FE80::0202:B3FF:FE1E:8329".
	ForwardedFor string `json:"forwarded_for"`

	// The request processing duration on the server, from the time the request was
	// received until the response was sent.
	Duration string `json:"duration"`

	// The referrer URL of the request, as defined in HTTP/1.1 Header Field
	// Definitions.
	Referrer string `json:"referrer"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler interface.
func (req *HTTPPayload) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("requestMethod", req.RequestMethod)
	enc.AddString("requestUrl", req.RequestURL)
	enc.AddInt("status", req.Status)
	enc.AddString("userAgent", req.UserAgent)
	enc.AddString("remoteIp", req.RemoteIP)
	enc.AddString("duration", req.Duration)
	enc.AddString("referrer", req.Referrer)
	enc.AddString("forwardedFor", req.ForwardedFor)

	return nil
}

// toZapField adds the correct Stackdriver field.
//
// see: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
func toZapField(req *HTTPPayload) zap.Field {
	return zap.Object("httpRequest", req)
}

// RequestLogger provides method to log http requests.
type RequestLogger struct {
	logger *zap.Logger
}

// NewRequestLogger returns http handler to log requests.
func NewRequestLogger(l *zap.Logger) *RequestLogger {
	return &RequestLogger{
		logger: l.WithOptions(
			zap.WithCaller(false),
			zap.AddStacktrace(zap.DPanicLevel),
		),
	}
}

// WithLogger returns a logging middleware to log http request.
func (l *RequestLogger) WithLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// wrap the responseWriter so, we can track the status code.
		wrapped := responseWriter{ResponseWriter: w}
		next.ServeHTTP(&wrapped, r)

		payload := &HTTPPayload{
			RequestMethod: r.Method,
			RequestURL:    r.RequestURI,
			Status:        wrapped.status,
			UserAgent:     r.UserAgent(),
			RemoteIP:      r.RemoteAddr,
			Referrer:      r.Referer(),
			Duration:      time.Since(start).String(),
			ForwardedFor:  r.Header.Get("x-forwarded-for"),
		}

		// What we want to build using buffer
		// message := r.Method + " " + r.RequestURI + " " + http.StatusText(wrapped.status)

		// Get bytes buffer from the pool and reset it to reset the garbage.
		buf := _buffer.Get().(*strings.Builder)
		buf.Reset()

		buf.WriteString(r.Method)
		buf.Write(_space)
		buf.WriteString(r.RequestURI)
		buf.Write(_space)
		buf.WriteString(http.StatusText(wrapped.status))

		message := buf.String()
		_buffer.Put(buf)

		if wrapped.status >= http.StatusBadRequest {
			l.logger.Error(message, toZapField(payload))
			return
		}

		l.logger.Info(message, toZapField(payload))
	})
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// Status returns http status code.
func (rw *responseWriter) Status() int {
	return rw.status
}

// WriteHeader writes the http response header.
func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}
