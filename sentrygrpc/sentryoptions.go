package sentrygrpc

import (
	"github.com/Flahmingo-Investments/helpers-go/ferrors"
)

// option is used to configure the interceptors.
// NOTE: Don't use it directly.
type option struct {
	// repanic configures whether to panic again after recovering from a
	// panic. Use this option if you have other panic handlers.
	repanic bool

	// reportOn configures whether to report an error. Defaults to
	// withDefaultCodes.
	reportOn ReportOn

	// relog configures whether sentry will log to terminal anything it catches
	// under the error codes provided to it
	relog bool
}

// InterceptorOption configuration overrider.
type InterceptorOption func(*option)

func buildOptions(interOptns ...InterceptorOption) option {
	opts := option{
		reportOn: withDefaultCodes,
		relog:    false,
		repanic:  true,
	}

	for _, interOptn := range interOptns {
		if interOptn != nil {
			interOptn(&opts)
		}
	}
	return opts
}

// WithRepanic configures whether to panic again after recovering from
// a panic. Use this option if you have other panic handlers.
func WithRepanic(repanic bool) InterceptorOption {
	return func(o *option) {
		o.repanic = repanic
	}
}

// WithReportOn configures which errors to report on.
func WithReportOn(repOn ReportOn) InterceptorOption {
	return func(o *option) {
		o.reportOn = repOn
	}
}

// WithReport configures which errors to report on.
func WithRelog(reLog bool) InterceptorOption {
	return func(o *option) {
		o.relog = reLog
	}
}

// ReportOn decides error should be reported to sentry.
type ReportOn func(error) bool

var defaultCodes = []ferrors.ErrorCode{
	ferrors.Internal,
	ferrors.Unimplemented,
	ferrors.Unknown,
}

func withDefaultCodes(err error) bool {
	if err == nil {
		return false
	}

	currentErrorCode := ferrors.Code(err)
	for i := range defaultCodes {
		if currentErrorCode == defaultCodes[i] {
			return true
		}
	}
	return false
}

// ReportAlways returns true if err is non-nil.
func ReportAlways(err error) bool {
	return err != nil
}

// ReportOnCodes returns true if ferror code matches culprit error code.
func ReportOnCodes(codes ...ferrors.ErrorCode) ReportOn {
	return func(err error) bool {
		if err == nil {
			return false
		}
		currentErrorCode := ferrors.Code(err)
		for i := range codes {
			if currentErrorCode == codes[i] {
				return true
			}
		}
		return false
	}
}
