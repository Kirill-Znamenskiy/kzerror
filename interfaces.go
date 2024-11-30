package kzerror

import "log/slog"

type ErrorWithMsg interface {
	error
	GetMsg() string
}
type ErrorWithAttrs interface {
	error
	GetAttrs() []slog.Attr
}
type ErrorUnwrappable interface {
	error
	Unwrap() error
}
