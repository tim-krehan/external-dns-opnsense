package main

import "log"

func listEntries() []Record {
	return []Record{
		{
			DNSName:    "example.com",
			Targets:    []string{"1.2.3.4"},
			RecordType: "A",
			TTL:        300,
		},
	}
}

func createEntry(rec Record) {
	log.Printf("Create: %+v\n", rec)
}

func deleteEntry(rec Record) {
	log.Printf("Delete: %+v\n", rec)
}
