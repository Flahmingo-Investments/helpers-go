// Package sentrygrpc provides interceptor for gRPC server.
package sentrygrpc

import (
	"context"

	"github.com/Flahmingo-Investments/helpers-go/flog"
	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
)

// SentryUnaryServerInterceptor is a middleware implementation of a GRPC server interceptor for panics in Unary operations
func SentryUnaryServerInterceptor(options ...InterceptorOption) grpc.UnaryServerInterceptor {
	opts := buildOptions(options...)
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		hub := sentry.CurrentHub().Clone()

		defer func() {
			if r := recover(); r != nil {
				hub.Recover(r)
				// If the option to throw panic after recovery is true
				if opts.repanic {
					panic(r)
				}
			}
		}()

		res, err := handler(ctx, req)

		// Checks if the error thrown is to be captured by Sentry
		// according to type of errors to be captured or not
		if opts.reportOn(err) {
			if opts.relog {
				flog.Errorf("sentry.relog: %+v", err)
			}
			hub.CaptureException(err)
		}

		return res, err
	}
}

// SentryStreamServerInterceptor is a middleware implementation of a GRPC server interceptor for panics in Stream operations
// TODO: NEEDS TO BE TESTED
func SentryStreamServerInterceptor(
	options ...InterceptorOption,
) grpc.StreamServerInterceptor {
	opts := buildOptions(options...)

	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		hub := sentry.CurrentHub().Clone()

		defer func() {
			if r := recover(); r != nil {
				hub.Recover(r)
				if opts.repanic {
					panic(r)
				}
			}
		}()

		err := handler(srv, stream)
		if opts.reportOn(err) {
			if opts.relog {
				flog.Errorf("sentry.relog: %+v", err)
			}
			hub.CaptureException(err)
		}

		return err
	}
}
