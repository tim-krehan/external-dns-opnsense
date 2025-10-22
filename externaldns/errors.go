package externaldns

import (
	"errors"
)

// ErrInvalidDomainName is returned when the DNSName is invalid or does not contain enough parts.
var ErrInvalidDomainName = errors.New("invalid domain name")
