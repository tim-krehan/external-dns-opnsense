package opnsense

import (
	"log"

	externaldns "external-dns-opnsense/externaldns"
)

// createEntry is a placeholder function for creating a new DNS entry.
// Currently, it only logs the record to be created.
func (cfg OpnSenseApi) CreateEntry(rec externaldns.Record) {
	log.Printf("Create: %+v\n", rec)
}
