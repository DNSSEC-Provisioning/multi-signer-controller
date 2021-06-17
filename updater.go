package main

import (
    "log"

    "github.com/miekg/dns"
)

//
// TODO: See if there is a better ways to give the insert/remove RRset
//
// Current implementation mimics dns.Insert()/.Remove() in the way that each
// entry in the first array is a call to these functions with the second
// array.
//
type Updater interface {
    Update(fqdn, signer string, inserts, removes *[][]dns.RR, output *[]string) error
    RemoveRRset(fqdn, signer string, rrsets [][]dns.RR, output *[]string) error
}

var Updaters map[string]Updater = make(map[string]Updater)

func GetUpdater(type_ string) Updater {
    updater, ok := Updaters[type_]
    if !ok {
        log.Fatal("No updater type", type_)
    }
    return updater
}
