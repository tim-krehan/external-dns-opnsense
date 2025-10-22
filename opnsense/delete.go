package opnsense

import (
	externaldns "external-dns-opnsense/externaldns"
	"log"
)

// deleteEntry is a placeholder function for deleting a DNS entry.
// Currently, it only logs the record to be deleted.
func (cfg OpnSenseApi) DeleteEntry(rec externaldns.Record) {
	log.Printf("Delete: %+v\n", rec)
}
