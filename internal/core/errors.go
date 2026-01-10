package core

import "errors"

const WrapErrFmt = "%w: %v"

var ErrInternal = errors.New("internal error")
