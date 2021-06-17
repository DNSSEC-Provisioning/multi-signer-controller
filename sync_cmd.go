package main

import (
    "fmt"
    "strconv"

    "github.com/miekg/dns"
)

func init() {
    Command["sync-dnskey"] = SyncDnskeyCmd
    Command["sync-cdscdnskeys"] = SyncCdscdnskeysCmd
    Command["sync-ns"] = SyncNsCmd

    CommandHelp["sync-dnskey"] = "Sync DNSKEYs between signers in a group, requires <fqdn>"
    CommandHelp["sync-cdscdnskeys"] = "Create CDS/CDNSKEYs from DNSKEYs and sync them between signers in a group, requires <fqdn>"
    CommandHelp["sync-ns"] = "Sync NSes between signers in a group, requires <fqdn>"
}

func SyncDnskeyCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    dnskeys := make(map[string][]*dns.DNSKEY)

    for _, signer := range signers {
        ip := Config.Get("signer:"+signer, "")
        if ip == "" {
            return fmt.Errorf("No ip|host for signer %s", signer)
        }

        m := new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeDNSKEY)

        c := new(dns.Client)
        r, _, err := c.Exchange(m, ip)

        if err != nil {
            return err
        }

        dnskeys[signer] = []*dns.DNSKEY{}

        for _, a := range r.Answer {
            dnskey, ok := a.(*dns.DNSKEY)
            if !ok {
                continue
            }

            dnskeys[signer] = append(dnskeys[signer], dnskey)

            if f := dnskey.Flags & 0x101; f == 256 {
                Config.SetIfNotExists("dnskey-origin:"+fmt.Sprintf("%d-%d-%s", dnskey.Protocol, dnskey.Algorithm, dnskey.PublicKey), signer)
            }
        }
    }

    for signer, keys := range dnskeys {
        leaving := Config.Get("signer-leaving:"+signer, "")
        if leaving != "" {
            *output = append(*output, fmt.Sprintf("Signer %s is leaving, removing it's DNSKEYs from others", signer))
            for _, key := range keys {
                if f := key.Flags & 0x101; f == 256 {
                    // check that it's our key
                    if Config.Get("dnskey-origin:"+fmt.Sprintf("%d-%d-%s", key.Protocol, key.Algorithm, key.PublicKey), "") != signer {
                        continue
                    }
                    *output = append(*output, fmt.Sprintf("- %s", key.PublicKey))

                    for osigner, okeys := range dnskeys {
                        if osigner == signer {
                            continue
                        }
                        if leaving := Config.Get("signer-leaving:"+osigner, ""); leaving != "" {
                            continue
                        }

                        found := false
                        for _, okey := range okeys {
                            if okey.PublicKey == key.PublicKey {
                                // if okey.Protocol != key.Protocol {
                                //     *output = append(*output, fmt.Sprintf("Found DNSKEY in %s but missmatch Protocol: %s", osigner, key.PublicKey))
                                //     break
                                // }
                                // if okey.Algorithm != key.Algorithm {
                                //     *output = append(*output, fmt.Sprintf("Found DNSKEY in %s but missmatch Protocol: %s", osigner, key.PublicKey))
                                //     break
                                // }
                                found = true
                                break
                            }
                        }
                        if found {
                            updater := GetUpdater(Config.Get("signer-type:"+osigner, "nsupdate"))
                            if err := updater.Update(args[0], osigner, nil, &[][]dns.RR{[]dns.RR{key}}, output); err != nil {
                                return err
                            }
                            *output = append(*output, fmt.Sprintf("  Removed DNSKEY from %s", osigner))
                        }
                    }
                }
            }
            continue
        }

        *output = append(*output, fmt.Sprintf("Syncing %s DNSKEYs", signer))

        for _, key := range keys {
            if f := key.Flags & 0x101; f == 256 {
                *output = append(*output, fmt.Sprintf("- %s", key.PublicKey))

                for osigner, okeys := range dnskeys {
                    if osigner == signer {
                        continue
                    }
                    if leaving := Config.Get("signer-leaving:"+osigner, ""); leaving != "" {
                        continue
                    }

                    found := false
                    for _, okey := range okeys {
                        if okey.PublicKey == key.PublicKey {
                            // if okey.Protocol != key.Protocol {
                            //     *output = append(*output, fmt.Sprintf("Found DNSKEY in %s but missmatch Protocol: %s", osigner, key.PublicKey))
                            //     break
                            // }
                            // if okey.Algorithm != key.Algorithm {
                            //     *output = append(*output, fmt.Sprintf("Found DNSKEY in %s but missmatch Protocol: %s", osigner, key.PublicKey))
                            //     break
                            // }
                            found = true
                            break
                        }
                    }

                    if !found {
                        updater := GetUpdater(Config.Get("signer-type:"+osigner, "nsupdate"))
                        if err := updater.Update(args[0], osigner, &[][]dns.RR{[]dns.RR{key}}, nil, output); err != nil {
                            return err
                        }
                        *output = append(*output, fmt.Sprintf("  Added DNSKEY to %s", osigner))
                    } else {
                        *output = append(*output, fmt.Sprintf("  Key exist in %s", osigner))
                    }
                }
            }
        }
    }

    return nil
}

func SyncCdscdnskeysCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    dnskeys := make(map[string][]*dns.DNSKEY)

    for _, signer := range signers {
        if leaving := Config.Get("signer-leaving:"+signer, ""); leaving != "" {
            continue
        }

        ip := Config.Get("signer:"+signer, "")
        if ip == "" {
            return fmt.Errorf("No ip|host for signer %s", signer)
        }

        m := new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeDNSKEY)

        c := new(dns.Client)
        r, _, err := c.Exchange(m, ip)

        if err != nil {
            return err
        }

        dnskeys[signer] = []*dns.DNSKEY{}

        for _, a := range r.Answer {
            dnskey, ok := a.(*dns.DNSKEY)
            if !ok {
                continue
            }

            dnskeys[signer] = append(dnskeys[signer], dnskey)
        }
    }

    cdses := []dns.RR{}
    cdnskeys := []dns.RR{}
    for _, keys := range dnskeys {
        for _, key := range keys {
            if f := key.Flags & 0x101; f == 257 {
                cdses = append(cdses, key.ToDS(dns.SHA256).ToCDS())
                cdnskeys = append(cdnskeys, key.ToCDNSKEY())
            }
        }
    }

    for _, signer := range signers {
        if leaving := Config.Get("signer-leaving:"+signer, ""); leaving != "" {
            continue
        }

        updater := GetUpdater(Config.Get("signer-type:"+signer, "nsupdate"))
        if err := updater.Update(args[0], signer, &[][]dns.RR{cdses, cdnskeys}, nil, output); err != nil {
            return err
        }
        *output = append(*output, fmt.Sprintf("  Added CDS/CDNSKEYs to %s", signer))
    }

    return nil
}

func SyncNsCmd(args []string, remote bool, output *[]string) error {
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

    nses := make(map[string][]*dns.NS)

    for _, signer := range signers {
        ip := Config.Get("signer:"+signer, "")
        if ip == "" {
            return fmt.Errorf("No ip|host for signer %s", signer)
        }

        m := new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeNS)
        c := new(dns.Client)
        r, _, err := c.Exchange(m, ip)
        if err != nil {
            return err
        }

        nses[signer] = []*dns.NS{}

        for _, a := range r.Answer {
            ns, ok := a.(*dns.NS)
            if !ok {
                continue
            }

            nses[signer] = append(nses[signer], ns)

            Config.SetIfNotExists("ns-origin:"+ns.Ns, signer)
        }
    }

    nsmap := make(map[string]*dns.NS)
    for _, rrs := range nses {
        for _, rr := range rrs {
            nsmap[rr.Ns] = rr
        }
    }
    nsset := []dns.RR{}
    nsrem := []dns.RR{}
    for _, rr := range nsmap {
        leaving := ""
        for _, signer := range signers {
            if Config.Get("signer-ns:"+signer, "") == rr.Ns {
                leaving = Config.Get("signer-leaving:"+signer, "")
                break
            }
        }
        if leaving != "" {
            *output = append(*output, "removing "+rr.Ns+", leaving signer")
            nsrem = append(nsrem, rr)
            continue
        }

        nsset = append(nsset, rr)
    }

    for _, signer := range signers {
        ns := Config.Get("signer-ns:"+signer, "")
        if ns == "" {
            continue
        }
        if _, ok := nsmap[ns]; !ok {
            rr := new(dns.NS)
            rr.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: uint32(ttl)}
            rr.Ns = ns
            nsset = append(nsset, rr)
        }
    }

    for _, signer := range signers {
        updater := GetUpdater(Config.Get("signer-type:"+signer, "nsupdate"))
        if err := updater.Update(args[0], signer, &[][]dns.RR{nsset}, &[][]dns.RR{nsrem}, output); err != nil {
            return err
        }
        *output = append(*output, fmt.Sprintf("  Add/rem'ed NSes to %s", signer))
    }

    return nil
}
