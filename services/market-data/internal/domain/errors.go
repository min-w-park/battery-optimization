package domain

import "errors"

var (
	// MarketPrice validation errors
	ErrInvalidPrice        = errors.New("price must be non-negative ($/MWh)")
	ErrInvalidDemand       = errors.New("demand must be greater than 0 (MW)")
	ErrInvalidRegion       = errors.New("region must be valid NEM region (NSW, VIC, QLD, SA, TAS)")
	ErrInvalidIntervalType = errors.New("interval type must be 5MIN_PREDISPATCH or 30MIN_PREDISPATCH")
	ErrIntervalInPast      = errors.New("interval start cannot be in the past")
	ErrPublishedInFuture   = errors.New("published time cannot be in the future")
	ErrIntervalAlignment   = errors.New("interval start must align with interval type (5-min or 30-min boundaries)")

	// Repository errors
	ErrNotFound          = errors.New("market price not found")
	ErrDuplicateInterval = errors.New("price for this region/interval/time already exists")
)
