package domain

import "errors"

// HTTPStatus maps a domain error to an HTTP status code and client-safe message.
// Returns ok=false when the error is not a known domain error.
func HTTPStatus(err error) (status int, message string, ok bool) {
	switch {
	case errors.Is(err, ErrDuplicateEmail), errors.Is(err, ErrDuplicatePointTypeID):
		return 409, err.Error(), true
	case errors.Is(err, ErrMemberNotFound), errors.Is(err, ErrRewardNotFound), errors.Is(err, ErrPointNotFound):
		return 404, err.Error(), true
	case errors.Is(err, ErrInvalidPointType), errors.Is(err, ErrPointsNotPositive):
		return 400, err.Error(), true
	case errors.Is(err, ErrInsufficientBalance):
		return 422, err.Error(), true
	default:
		return 0, "", false
	}
}
