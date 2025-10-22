package externaldns

import (
	"strings"
)

// Record represents a DNS record with its name, targets, type, and TTL.
type Record struct {
	DNSName    string   `json:"dnsName"`    // Fully qualified domain name (FQDN) of the record.
	Targets    []string `json:"targets"`    // List of target IPs or hostnames.
	RecordType string   `json:"recordType"` // Type of the DNS record (e.g., A, CNAME, etc.).
	TTL        int64    `json:"ttl"`        // Time-to-live for the DNS record.
}

// GetTopLevelDomain extracts the top-level domain (TLD) from the DNSName.
// For simplicity, it assumes the TLD is the last two segments of the DNSName.
func (r Record) GetTopLevelDomain() (string, error) {
	// Split the FQDN into parts using the dot as a delimiter.
	parts := strings.Split(r.DNSName, ".")
	// If there are fewer than two parts, return an error.
	if len(parts) < 2 {
		return "", ErrInvalidDomainName
	}
	// Join the last two parts to form the TLD.
	return strings.Join(parts[len(parts)-2:], "."), nil
}
