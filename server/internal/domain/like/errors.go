package like

import "errors"

var (
	ErrInvalidTargetType = errors.New("invalid like target type")
	ErrInvalidTargetID   = errors.New("invalid like target id")
	ErrTargetNotFound    = errors.New("like target not found")
)
