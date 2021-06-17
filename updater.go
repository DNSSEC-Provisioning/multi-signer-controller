package main

import (
    "log"

    "github.com/miekg/dns"
)

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
