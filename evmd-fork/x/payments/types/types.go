package types

import "time"

// IsExpired checks expiration against the provided time.
func (p PaymentRequest) IsExpired(now time.Time) bool {
	if p.ExpiresAtUnix == 0 {
		return false
	}
	return now.Unix() >= int64(p.ExpiresAtUnix)
}

// IsPending returns true when the request is awaiting payment.
func (p PaymentRequest) IsPending() bool {
	return p.Status == PaymentStatus_PAYMENT_STATUS_PENDING
}
