package main

type Record struct {
    DNSName    string   `json:"dnsName"`
    Targets    []string `json:"targets"`
    RecordType string   `json:"recordType"`
    TTL        int64    `json:"ttl"`
}

type ApplyChangesRequest struct {
    Create []Record `json:"create"`
    Delete []Record `json:"delete"`
}
