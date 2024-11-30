package kzerror

import (
	"log/slog"
)

var _ error = (*Error)(nil)
var _ ErrorWithMsg = (*Error)(nil)
var _ ErrorWithAttrs = (*Error)(nil)
var _ ErrorUnwrappable = (*Error)(nil)

type Error struct {
	Msg    string
	Attrs  []slog.Attr
	SubErr error
}

func (e *Error) GetMsg() string {
	return e.Msg
}

func (e *Error) GetAttrs() []slog.Attr {
	return e.Attrs
}

func (e *Error) Error() string {
	if e.SubErr == nil {
		return e.Msg
	}
	if e.Msg == "" {
		return e.SubErr.Error()
	}
	return e.Msg + " <<< " + e.SubErr.Error()
}

func (e *Error) Unwrap() error {
	return e.SubErr
}

func (e *Error) LogValue() slog.Value {
	if e.SubErr == nil && len(e.Attrs) <= 0 {
		return slog.StringValue(e.Msg)
	}
	attrs := make([]slog.Attr, 0, len(e.Attrs)+2)
	attrs = append(attrs, slog.String("msg", e.Msg))
	attrs = append(attrs, e.Attrs...)

	subErr := e.SubErr
	if subErr != nil {
		attrs = append(attrs, NewSLogAttr("suberr", BuildSLogValue(subErr)))
	}

	return slog.GroupValue(attrs...)
}

func NewSLogAttr(key string, value slog.Value) slog.Attr {
	return slog.Attr{Key: key, Value: value}
}

func BuildSLogValue(wrkErr error) slog.Value {
	if wrkErrLogValue, ok := wrkErr.(slog.LogValuer); ok {
		return wrkErrLogValue.LogValue()
	}

	attrs := make([]slog.Attr, 0)
	if wrkErrWithMsg, ok := wrkErr.(ErrorWithMsg); ok {
		attrs = append(attrs, slog.String("msg", wrkErrWithMsg.GetMsg()))
	} else {
		attrs = append(attrs, slog.String("error-msg", wrkErr.Error()))
	}
	if wrkErrWithAttrs, ok := wrkErr.(ErrorWithAttrs); ok {
		attrs = append(attrs, wrkErrWithAttrs.GetAttrs()...)
	}
	if wrkErrUnwrappable, ok := wrkErr.(ErrorUnwrappable); ok {
		subErr := wrkErrUnwrappable.Unwrap()
		if subErr != nil {
			attrs = append(attrs, NewSLogAttr("suberr", BuildSLogValue(subErr)))
		}
	}
	if len(attrs) == 1 {
		return attrs[0].Value
	}
	return slog.GroupValue(attrs...)
}

func NewErr(msg string, attrs ...slog.Attr) *Error {
	return WrapErrMsg(nil, msg, attrs...)
}
func WrapErr(err error, attrs ...slog.Attr) *Error {
	return WrapErrMsg(err, "", attrs...)
}
func WrapErrMsg(err error, msg string, attrs ...slog.Attr) *Error {
	if err == nil {
		for _, attr := range attrs {
			if attr.Key == "err" || attr.Key == "error" {
				if e, ok := attr.Value.Any().(error); ok {
					err = e
					break
				}
			}
		}
	}
	return &Error{Msg: msg, Attrs: attrs, SubErr: err}
}
