package myresolver

import (
	"net/netip"

	"github.com/oschwald/maxminddb-golang"
)

type IPDB struct {
	db *maxminddb.Reader
}

func ParseIPDBFromFile(file string) (IPDB, error) {
	r, err := maxminddb.Open(file)
	if err != nil {
		return IPDB{}, err
	}
	return IPDB{db: r}, nil
}

func (db *IPDB) Close() error {
	return db.db.Close()
}

func (db *IPDB) LookupIP(addr netip.Addr) (asn uint64, desc string, err error) {
	var record struct {
		Asn  uint64 `maxminddb:"autonomous_system_number"`
		Desc string `maxminddb:"autonomous_system_organization"`
	}

	a := addr.As16()
	if err := db.db.Lookup(a[:], &record); err != nil {
		return 0, "", err
	}

	return record.Asn, record.Desc, nil
}
