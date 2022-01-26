// Package ferrors provides simple error handling primitives.
//
// The traditional error handling idiom in Go is roughly akin to
//
//     if err != nil {
//             return err
//     }
//
// which when applied recursively up the call stack results in error reports
// without context or debugging information. The ferrors package allows
// programmers to add context to the failure path in their code in a way
// that does not destroy the original value of the error.
//
// It also provides some useful error primitives to reduce unnecessary burden
// and duplicacy from code.
package ferrors

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var ( // Separator for wrapped error.
	_separator = []byte(": ")

	// Line separator for multiline messages or details.
	_lineSeparator = []byte("\n-  ")

	// Line separator for nested messages or details.
	_nestedlineSeperator = []byte("\n\t-  ")
)

// buffer pool to reduce string allocations.
var _buffer = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// Used when we could not determine the file or function name in stack trace.
const _unknown = "unknown"

// ErrorCode is internal domain error codes
type ErrorCode uint32

// String returns the string representation of the error code
func (e ErrorCode) String() string {
	return codes.Code(e).String()
}

const (
	// Unknown error. Default error type if no error type is provided
	Unknown ErrorCode = ErrorCode(codes.Unknown)

	// InvalidArgument indicates client specified an invalid argument.
	// It indicates arguments that are problematic regardless of the state of the
	// system
	// (e.g., a malformed file name).
	InvalidArgument ErrorCode = ErrorCode(codes.InvalidArgument)

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	NotFound ErrorCode = ErrorCode(codes.NotFound)

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	AlreadyExists ErrorCode = ErrorCode(codes.AlreadyExists)

	// PermissionDenied indicates the caller does not have permission to
	// execute the specified operation.
	PermissionDenied ErrorCode = ErrorCode(codes.PermissionDenied)

	// FailedPrecondition indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	FailedPrecondition ErrorCode = ErrorCode(codes.FailedPrecondition)

	// OutOfRange means operation was attempted past the valid range.
	// E.g., seeking or reading past end of file.
	//
	// Unlike InvalidArgument, this error indicates a problem that may
	// be fixed if the system state changes. For example, a 32-bit file
	// system will generate InvalidArgument if asked to read at an
	// offset that is not in the range [0,2^32-1], but it will generate
	// OutOfRange if asked to read from an offset past the current
	// file size.
	OutOfRange ErrorCode = ErrorCode(codes.OutOfRange)

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	Unimplemented ErrorCode = ErrorCode(codes.Unimplemented)

	// Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	Internal ErrorCode = ErrorCode(codes.Internal)

	// Unavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	Unavailable ErrorCode = ErrorCode(codes.Unavailable)

	// Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	Unauthenticated ErrorCode = ErrorCode(codes.Unauthenticated)
)

// compile time check.
var (
	_ error = (*fundamental)(nil)
	_ error = (*withFields)(nil)
	_ error = (*wrapped)(nil)
)

// New returns an error with the supplied message.
// It also records the stack trace at the point it was called.
func New(message string) Ferror {
	return &fundamental{
		ErrorCode: Unknown,
		Msg:       message,
		stack:     callers(),
	}
}

// Newf formats according to a format specifier and returns the string
// as a value that satisfies error.
// It also records the stack trace at the point it was called.
func Newf(format string, args ...interface{}) Ferror {
	return &fundamental{
		ErrorCode: Unknown,
		Msg:       fmt.Sprintf(format, args...),
		stack:     callers(),
	}
}

// WithCode returns an error with the supplied message and error code
// It also records the stack trace at the point it was called.
func WithCode(code ErrorCode, message string, detail ...*ErrorDetail) error {
	var dtl *ErrorDetail
	if len(detail) > 0 {
		dtl = detail[0]
	}

	return (&fundamental{
		ErrorCode: code,
		Msg:       message,
		stack:     callers(),
	}).WithDetail(dtl)
}

// fundamental is an error that contains an error code, a message and stack trace
type fundamental struct {
	ErrorCode ErrorCode    `json:"errorCode"`
	Detail    *ErrorDetail `json:"detail,omitempty"`
	Msg       string       `json:"msg"`
	stack     *stack
}

// WithDetail adds error detail to Ferror.
func (f *fundamental) WithDetail(detail *ErrorDetail) Ferror {
	f.Detail = detail
	return f
}

// Code returns the error code.
func (f *fundamental) Code() ErrorCode { return f.ErrorCode }

// Error implements error interface for fundamental
func (f *fundamental) Error() string {
	buf, _ := _buffer.Get().(*bytes.Buffer)
	buf.Reset()

	buf.WriteByte('(')
	buf.WriteString(f.ErrorCode.String())
	buf.WriteByte(')')
	buf.WriteByte(' ')
	buf.WriteString(f.Msg)
	if f.Detail != nil {
		buf.Write(_lineSeparator)
		buf.WriteString(f.Detail.String())
	}

	s := buf.String()
	_buffer.Put(buf)

	return s
}

// Format implements Formatter interface for fundamental.
func (f *fundamental) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, f.Error())
			f.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, f.Error())
	}
}

// GRPCStatus is implements GRPCStatus interface for fundamental.
func (f *fundamental) GRPCStatus() *status.Status {
	st := status.New(codes.Code(f.ErrorCode), f.Msg)
	if f.Detail != nil {
		std, err := st.WithDetails(f.Detail)
		// check where there was an error while attaching the metadata to status in
		// above switch block
		if err != nil {
			// If this errored, it will always error here, so better panic so we can
			// figure out why this was silently passing.
			panic(fmt.Sprintf("unable to attach metadata: %+v", err))
		}
		st = std
	}
	return st
}

// withFields is same as fundamental error but it can hold fields that caused the
// error.
type withFields struct {
	*fundamental
	Fields []Field
}

// WithDetail adds error detail to Ferror.
func (w *withFields) WithDetail(detail *ErrorDetail) Ferror {
	w.Detail = detail
	return w
}

// Format implements Formatter interface for withFields.
func (w *withFields) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, w.Error())
			w.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, w.Error())
	}
}

// Error implements error interface for withFields
func (w *withFields) Error() string {
	if len(w.Fields) > 0 {
		buf, _ := _buffer.Get().(*bytes.Buffer)
		buf.Reset()

		buf.WriteString(w.fundamental.Error())

		if len(w.Fields) > 0 {
			buf.Write(_lineSeparator)
			buf.WriteString("error fields:")
		}

		for _, field := range w.Fields {
			buf.Write(_nestedlineSeperator)
			buf.WriteString(field.Name)
			buf.Write(_separator)
			buf.WriteString(field.Description)
		}

		s := buf.String()
		_buffer.Put(buf)

		return s
	}

	return w.fundamental.Error()
}

// GRPCStatus is implements GRPCStatus interface for withFields.
func (w *withFields) GRPCStatus() *status.Status {
	st := status.New(codes.Code(w.ErrorCode), w.Msg)

	var std *status.Status
	var err error

	// We do not care about other error codes in withFields
	//
	// nolint:exhaustive
	switch w.ErrorCode {
	case InvalidArgument:
		br := &errdetails.BadRequest{}
		for _, f := range w.Fields {
			v := &errdetails.BadRequest_FieldViolation{
				Description: f.Description,
				Field:       f.Name,
			}

			br.FieldViolations = append(br.FieldViolations, v)
		}
		std, err = st.WithDetails(br)

	case FailedPrecondition:
		pf := &errdetails.PreconditionFailure{}
		for _, f := range w.Fields {
			v := &errdetails.PreconditionFailure_Violation{
				Description: f.Description,
				Subject:     f.Name,
			}

			pf.Violations = append(pf.Violations, v)
		}

		std, err = st.WithDetails(pf)

	default:
		return st
	}

	// check where there was an error while attaching the metadata to status in
	// above switch block
	if err != nil {
		// If this errored, it will always error here, so better panic so we can
		// figure out why this was silently passing.
		panic(fmt.Sprintf("unable to attach metadata: %+v", err))
	}

	return std
}

// wrapped wraps an error and add stack traces.
type wrapped struct {
	msgs   []string
	stacks []*stack
	cause  error
}

// Error implements the error interface for wrapped.
func (w *wrapped) Error() string {
	if len(w.msgs) > 0 {
		// We can optimize the buffer using buffer pool
		buf, _ := _buffer.Get().(*bytes.Buffer)
		buf.Reset()

		for _, m := range w.msgs {
			buf.WriteString(m)
			buf.Write(_separator)
		}

		buf.WriteString(w.cause.Error())

		s := buf.String()
		_buffer.Put(buf)

		return s
	}

	return w.cause.Error()
}

// Format implements Formatter interface for wrapped.
func (w *wrapped) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, w.Error())
			for _, stack := range w.stacks {
				stack.Format(s, verb)
			}
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, w.Error())
	}
}

// Code returns the error code.
func (w *wrapped) Code() ErrorCode {
	if e, ok := w.cause.(Ferror); ok {
		return e.Code()
	}
	return Unknown
}

// GRPCStatus is implements GRPCStatus interface for wrapped.
func (w *wrapped) GRPCStatus() *status.Status {
	return status.Convert(w.cause)
}

// WithStack add stack trace to an error
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	if werr, ok := err.(*wrapped); ok {
		werr.stacks = append(werr.stacks, callers())
		return werr
	}

	return &wrapped{
		cause:  err,
		stacks: []*stack{callers()},
	}
}

// Cause return the original wrapped error
func (w *wrapped) Cause() error { return w.cause }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *wrapped) Unwrap() error { return w.cause }

// Wrap wraps an error with custom message.
// It also records the stack trace at the point it was called.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	// This part is same as wrapf.
	// But we don't want wrapf in call stack.
	if werr, ok := err.(*wrapped); ok {
		werr.msgs = append(werr.msgs, msg)
		werr.stacks = append(werr.stacks, callers())
		return werr
	}

	return &wrapped{
		cause:  err,
		stacks: []*stack{callers()},
		msgs:   []string{msg},
	}
}

// Wrapf wraps an error with custom formatted message.
// It also records the stack trace at the point it was called.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if werr, ok := err.(*wrapped); ok {
		werr.msgs = append(werr.msgs, fmt.Sprintf(format, args...))
		werr.stacks = append(werr.stacks, callers())
		return werr
	}

	return &wrapped{
		cause:  err,
		stacks: []*stack{callers()},
		msgs:   []string{fmt.Sprintf(format, args...)},
	}
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

// Field is the field that caused the error
type Field struct {
	Name        string
	Description string
}

// NewInvalidArgumentError return an invalid argument error.
// It also records the stack trace at the point it was called.
func NewInvalidArgumentError(msg string, fields ...Field) Ferror {
	return &withFields{
		fundamental: &fundamental{
			ErrorCode: InvalidArgument,
			stack:     callers(),
			Msg:       msg,
		},
		Fields: fields,
	}
}

// NewAlreadyExistsError returns an already exists error
// It also records the stack trace at the point it was called.
func NewAlreadyExistsError(msg string, fields ...Field) Ferror {
	return &withFields{
		fundamental: &fundamental{
			ErrorCode: AlreadyExists,
			stack:     callers(),
			Msg:       msg,
		},
		Fields: fields,
	}
}

// NewInternalError returns an internal error.
// It also records the stack trace at the point it was called.
func NewInternalError(msg string) Ferror {
	return &fundamental{
		ErrorCode: Internal,
		stack:     callers(),
		Msg:       msg,
	}
}

// NewOutOfRangeError returns an out of range error.
// It also records the stack trace at the point it was called.
func NewOutOfRangeError(msg string, fields ...Field) Ferror {
	return &withFields{
		fundamental: &fundamental{
			ErrorCode: OutOfRange,
			stack:     callers(),
			Msg:       msg,
		},
		Fields: fields,
	}
}

// NewPermissionDeniedError returns permission denied error
// It also records the stack trace at the point it was called.
func NewPermissionDeniedError(msg string) Ferror {
	return &fundamental{
		ErrorCode: PermissionDenied,
		stack:     callers(),
		Msg:       msg,
	}
}

// NewUnauthenticatedError returns an unauthenticated error.
// It also records the stack trace at the point it was called.
func NewUnauthenticatedError(msg string) Ferror {
	return &fundamental{
		ErrorCode: Unauthenticated,
		stack:     callers(),
		Msg:       msg,
	}
}

// NewNotFoundError returns an not found error.
// It also records the stack trace at the point it was called.
func NewNotFoundError(msg string) Ferror {
	return &fundamental{
		ErrorCode: NotFound,
		stack:     callers(),
		Msg:       msg,
	}
}

// Ferror is an error that contains error code, details, and stack traces.
type Ferror interface {
	// Code returns the error code.
	Code() ErrorCode
	// WithDetail attaches an error detail to Ferror.
	WithDetail(*ErrorDetail) Ferror

	error
}

// Code returns error code
func Code(err error) ErrorCode {
	if e, ok := err.(Ferror); ok {
		return e.Code()
	}
	return Unknown
}

// ErrorDetail describes the cause of the error with structured details.
//
// Example of an error when creating an account with email, when email already exists.
// is not enabled:
//
//     { "reason": "EMAIL_ALREADY_EXISTS"
//       "metadata": {
//         "email": "email is already in use"
//       }
//     }
//
// This response indicates that the pubsub.googleapis.com API is not enabled.
//
// Example of an error that is returned when attempting to create a Spanner
// instance in a region that is out of stock:
//
//     { "reason": "MARKET_CLOSED"
//       "metadata": {
//         "info": "Market is closed."
//       }
//     }
type ErrorDetail errdetails.ErrorInfo

// Reset resets the ErrorDetail.
func (e *ErrorDetail) Reset() {
	(*errdetails.ErrorInfo)(e).Reset()
}

// String implements the fmt.Stringer interface.
func (e *ErrorDetail) String() string {
	return (*errdetails.ErrorInfo)(e).String()
}

// ProtoMessage implements proto Message interface.
func (*ErrorDetail) ProtoMessage() {}

// ProtoReflect returns a reflective view of message.
func (e *ErrorDetail) ProtoReflect() protoreflect.Message {
	return (*errdetails.ErrorInfo)(e).ProtoReflect()
}
