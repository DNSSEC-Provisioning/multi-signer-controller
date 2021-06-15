package main

import (
    "fmt"
    "time"

    "github.com/miekg/dns"
)

func init() {
    Command["remove-cdcdnskeys"] = RemoveCdcdnskeysCmd
    Command["remove-csync"] = RemoveCsyncCmd
    Command["remove-ns"] = RemoveNsCmd

    CommandHelp["remove-cdcdnskeys"] = "Remove all CDS/CDNSKEYs from signers in a group, requires <fqdn>"
    CommandHelp["remove-csync"] = "Remove all CSYNCs from signers in a group, requires <fqdn>"
    CommandHelp["remove-ns"] = "Remove a custom NS from all signers in a group, requires <fqdn> <ns fqdn to remove>"
}

func RemoveCdcdnskeysCmd(args []string, remote bool, output *[]string) error {
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

    m := new(dns.Msg)
    m.SetUpdate(args[0])
    m.RemoveRRset([]dns.RR{cds})
    m.RemoveRRset([]dns.RR{cdnskey})

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

            *output = append(*output, fmt.Sprintf("Remove took %v", rtt))
            *output = append(*output, in.String())

            *output = append(*output, fmt.Sprintf("  Removed CD/CDNSKEYs from %s", signer))
            break

        default:
            return fmt.Errorf("Unknown signer type %s for %s", stype, signer)
        }
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

    m := new(dns.Msg)
    m.SetUpdate(args[0])
    m.RemoveRRset([]dns.RR{csync})

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

            *output = append(*output, fmt.Sprintf("Remove took %v", rtt))
            *output = append(*output, in.String())

            *output = append(*output, fmt.Sprintf("  Removed CSYNC from %s", signer))
            break

        default:
            return fmt.Errorf("Unknown signer type %s for %s", stype, signer)
        }
    }

    return nil
}

func RemoveNsCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 2 {
        return fmt.Errorf("requires <fqdn> <ns fqdn to remove>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    ns := new(dns.NS)
    ns.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 0}
    ns.Ns = args[1]

    m := new(dns.Msg)
    m.SetUpdate(args[0])
    m.Remove([]dns.RR{ns})

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

            *output = append(*output, fmt.Sprintf("Remove took %v", rtt))
            *output = append(*output, in.String())

            *output = append(*output, fmt.Sprintf("  Removed NS from %s", signer))
            break

        default:
            return fmt.Errorf("Unknown signer type %s for %s", stype, signer)
        }
    }

    return nil
}
