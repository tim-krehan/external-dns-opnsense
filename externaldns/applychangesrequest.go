package externaldns

// ApplyChangesRequest represents a request to create or delete DNS records.
type ApplyChangesRequest struct {
	Create []Record `json:"create"` // Records to be created.
	Delete []Record `json:"delete"` // Records to be deleted.
}
