package domain

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrAccountSuspended   = errors.New("account suspended")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrContentBlocked     = errors.New("content blocked by moderation policy")
	ErrPhotoRejected      = errors.New("photo rejected by verification")
	ErrPhotoPendingReview = errors.New("photo submitted for manual review")
)
