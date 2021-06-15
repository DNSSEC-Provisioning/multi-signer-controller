package main

import (
    "fmt"
    "strconv"

    "github.com/miekg/dns"
)

func init() {
    Command["add-csync"] = AddCsyncCmd

    CommandHelp["add-csync"] = "Add CSYNC records on all signers for a group, requires <fqdn>"
}

func AddCsyncCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    ttl, err := strconv.Atoi(Config.Get("group-ttl:"+args[0], "300"))
    if err != nil {
        ttl = 300
    }

    signers := Config.ListGet("signers:" + args[0])

    for _, signer := range signers {
        ip := Config.Get("signer:"+signer, "")
        if ip == "" {
            return fmt.Errorf("No ip|host for signer %s", signer)
        }

        m := new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeSOA)
        c := new(dns.Client)
        r, _, err := c.Exchange(m, ip)
        if err != nil {
            return err
        }

        for _, a := range r.Answer {
            soa, ok := a.(*dns.SOA)
            if !ok {
                continue
            }

            csync := new(dns.CSYNC)
            csync.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeCSYNC, Class: dns.ClassINET, Ttl: uint32(ttl)}
            csync.Serial = soa.Serial
            csync.Flags = 3
            csync.TypeBitMap = []uint16{dns.TypeA, dns.TypeNS, dns.TypeAAAA}

            updater := GetUpdater(Config.Get("signer-type:"+signer, "nsupdate"))
            if err := updater.Update(args[0], signer, &[][]dns.RR{[]dns.RR{csync}}, nil, output); err != nil {
                return err
            }
            *output = append(*output, fmt.Sprintf("  Added CSYNC to %s", signer))
            break
        }
    }

    return nil
}
