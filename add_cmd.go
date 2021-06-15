package main

import (
    "fmt"
    "time"

    "github.com/miekg/dns"
)

func init() {
    Command["add-csync"] = AddCsyncCmd
    Command["add-ns"] = AddNsCmd

    CommandHelp["add-csync"] = "Add CSYNC records on all signers for a group, requires <fqdn>"
    CommandHelp["add-ns"] = "Add a custom NS to all signers in a group, requires <fqdn> <ns fqdn to add>"
}

func AddCsyncCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
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
            csync.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeCSYNC, Class: dns.ClassINET, Ttl: 300}
            csync.Serial = soa.Serial
            csync.Flags = 3
            csync.TypeBitMap = []uint16{dns.TypeA, dns.TypeNS, dns.TypeAAAA}

            m = new(dns.Msg)
            m.SetUpdate(args[0])
            m.Insert([]dns.RR{csync})

            *output = append(*output, m.String())

            stype := Config.Get("signer-type:"+signer, "nsupdate")

            switch stype {
            case "nsupdate":
                tsigkey := Config.Get("signer-tsigkey:"+signer, "")
                if tsigkey == "" {
                    *output = append(*output, fmt.Sprintf("Missing signer %s TSIG key, can't sync", signer))
                    continue
                }

                secret := Config.Get("tsigkey-"+tsigkey, "")
                if secret == "" {
                    *output = append(*output, fmt.Sprintf("Missing TSIG key %s, can't sync %s", tsigkey, signer))
                    continue
                }

                m.SetTsig(tsigkey+".", dns.HmacSHA256, 300, time.Now().Unix())

                c := new(dns.Client)
                c.TsigSecret = map[string]string{tsigkey + ".": secret}
                in, rtt, err := c.Exchange(m, ip)
                if err != nil {
                    return err
                }

                *output = append(*output, fmt.Sprintf("Insert took %v", rtt))
                *output = append(*output, in.String())

                *output = append(*output, fmt.Sprintf("  Added CSYNC to %s", signer))
                break

            default:
                return fmt.Errorf("Unknown signer type %s for %s", stype, signer)
            }

            break
        }
    }

    return nil
}

func AddNsCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 2 {
        return fmt.Errorf("requires <fqdn> <ns fqdn to add>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    ns := new(dns.NS)
    ns.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 300}
    ns.Ns = args[1]

    m := new(dns.Msg)
    m.SetUpdate(args[0])
    m.Insert([]dns.RR{ns})

    *output = append(*output, m.String())

    for _, signer := range signers {
        ip := Config.Get("signer:"+signer, "")
        if ip == "" {
            *output = append(*output, fmt.Sprintf("No ip|host for signer %s(???), can't sync it", signer))
            continue
        }
        stype := Config.Get("signer-type:"+signer, "nsupdate")

        switch stype {
        case "nsupdate":
            tsigkey := Config.Get("signer-tsigkey:"+signer, "")
            if tsigkey == "" {
                *output = append(*output, fmt.Sprintf("Missing signer %s TSIG key, can't sync", signer))
                continue
            }

            secret := Config.Get("tsigkey-"+tsigkey, "")
            if secret == "" {
                *output = append(*output, fmt.Sprintf("Missing TSIG key %s, can't sync %s", tsigkey, signer))
                continue
            }

            m.SetTsig(tsigkey+".", dns.HmacSHA256, 300, time.Now().Unix())

            c := new(dns.Client)
            c.TsigSecret = map[string]string{tsigkey + ".": secret}
            in, rtt, err := c.Exchange(m, ip)
            if err != nil {
                return err
            }

            *output = append(*output, fmt.Sprintf("Insert took %v", rtt))
            *output = append(*output, in.String())

            *output = append(*output, fmt.Sprintf("  Added NS to %s", signer))
            break

        default:
            return fmt.Errorf("Unknown signer type %s for %s", stype, signer)
        }
    }

    return nil
}
