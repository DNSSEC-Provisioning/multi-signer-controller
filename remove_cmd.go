package main

import (
    "fmt"

    "github.com/miekg/dns"
)

func init() {
    Command["remove-cdscdnskeys"] = RemoveCdscdnskeysCmd
    Command["remove-csync"] = RemoveCsyncCmd

    CommandHelp["remove-cdscdnskeys"] = "Remove all CDS/CDNSKEYs from signers in a group, requires <fqdn>"
    CommandHelp["remove-csync"] = "Remove all CSYNCs from signers in a group, requires <fqdn>"
}

func RemoveCdscdnskeysCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    cds := new(dns.CDS)
    cds.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeCDS, Class: dns.ClassINET, Ttl: 0}

    cdnskey := new(dns.CDNSKEY)
    cdnskey.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeCDNSKEY, Class: dns.ClassINET, Ttl: 0}

    for _, signer := range signers {
        updater := GetUpdater(Config.Get("signer-type:"+signer, "nsupdate"))
        if err := updater.RemoveRRset(args[0], signer, [][]dns.RR{[]dns.RR{cds}, []dns.RR{cdnskey}}, output); err != nil {
            return err
        }
        *output = append(*output, fmt.Sprintf("  Removed CDS/CDNSKEYs from %s", signer))
    }

    return nil
}

func RemoveCsyncCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    csync := new(dns.CSYNC)
    csync.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeCSYNC, Class: dns.ClassINET, Ttl: 0}

    for _, signer := range signers {
        updater := GetUpdater(Config.Get("signer-type:"+signer, "nsupdate"))
        if err := updater.RemoveRRset(args[0], signer, [][]dns.RR{[]dns.RR{csync}}, output); err != nil {
            return err
        }
        *output = append(*output, fmt.Sprintf("  Removed CSYNC from %s", signer))
    }

    return nil
}
