package main

import (
    "fmt"

    "github.com/miekg/dns"
)

func init() {
    Command["status"] = StatusCmd

    CommandHelp["status"] = "Check status of a signer group, requires <fqdn>"
}

func StatusCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    dnskeys := make(map[string][]*dns.DNSKEY)
    cdnskeys := make(map[string][]*dns.CDNSKEY)
    cdses := make(map[string][]*dns.CDS)
    nses := make(map[string][]*dns.NS)

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

            owner := Config.Get("dnskey-origin:"+fmt.Sprintf("%d-%d-%s", dnskey.Protocol, dnskey.Algorithm, dnskey.PublicKey), "")
            if owner != "" {
                owner = " (owner: " + owner + ")"
            }

            *output = append(*output, fmt.Sprintf("%s: found DNSKEY %d %d %d %s%s", signer, dnskey.Flags, dnskey.Protocol, dnskey.Algorithm, dnskey.PublicKey, owner))

            dnskeys[signer] = append(dnskeys[signer], dnskey)
        }

        m = new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeCDS)
        r, _, err = c.Exchange(m, ip)
        if err != nil {
            return err
        }
        cdses[signer] = []*dns.CDS{}
        for _, a := range r.Answer {
            cds, ok := a.(*dns.CDS)
            if !ok {
                continue
            }

            *output = append(*output, fmt.Sprintf("%s: found CDS %d %d %d %s", signer, cds.KeyTag, cds.Algorithm, cds.DigestType, cds.Digest))

            cdses[signer] = append(cdses[signer], cds)
        }

        m = new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeCDNSKEY)
        r, _, err = c.Exchange(m, ip)
        if err != nil {
            return err
        }
        cdnskeys[signer] = []*dns.CDNSKEY{}
        for _, a := range r.Answer {
            cdnskey, ok := a.(*dns.CDNSKEY)
            if !ok {
                continue
            }

            *output = append(*output, fmt.Sprintf("%s: found CDNSKEY %d %d %d %s", signer, cdnskey.Flags, cdnskey.Protocol, cdnskey.Algorithm, cdnskey.PublicKey))

            cdnskeys[signer] = append(cdnskeys[signer], cdnskey)
        }

        m = new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeNS)
        r, _, err = c.Exchange(m, ip)
        if err != nil {
            return err
        }
        nses[signer] = []*dns.NS{}
        for _, a := range r.Answer {
            ns, ok := a.(*dns.NS)
            if !ok {
                continue
            }

            owner := Config.Get("ns-origin:"+ns.Ns, "")
            if owner != "" {
                owner = " (owner: " + owner + ")"
            }

            *output = append(*output, fmt.Sprintf("%s: found NS %s%s", signer, ns.Ns, owner))

            nses[signer] = append(nses[signer], ns)
        }
    }

    group_dnskeys_synced := true
    for signer, keys := range dnskeys {
        *output = append(*output, fmt.Sprintf("Check sync status of %s DNSKEYs", signer))

        for _, key := range keys {
            if f := key.Flags & 0x101; f == 256 {
                for osigner, okeys := range dnskeys {
                    if osigner == signer {
                        continue
                    }

                    found := false
                    for _, okey := range okeys {
                        if f := okey.Flags & 0x101; f == 256 && okey.PublicKey == key.PublicKey {
                            if okey.Protocol != key.Protocol {
                                *output = append(*output, fmt.Sprintf("Found DNSKEY in %s but missmatch Protocol: %s", osigner, key.PublicKey))
                                break
                            }
                            if okey.Algorithm != key.Algorithm {
                                *output = append(*output, fmt.Sprintf("Found DNSKEY in %s but missmatch Protocol: %s", osigner, key.PublicKey))
                                break
                            }
                            found = true
                            break
                        }
                    }

                    if !found {
                        *output = append(*output, fmt.Sprintf("DNSKEY missing in %s: %s", osigner, key.PublicKey))
                        group_dnskeys_synced = false
                    }
                }
            }
        }
    }
    if group_dnskeys_synced {
        Config.Set("group-dnskeys-synced:"+args[0], "yes")
    } else {
        Config.Remove("group-dnskeys-synced:" + args[0])
    }

    ksks := []*dns.DNSKEY{}
    for _, keys := range dnskeys {
        for _, key := range keys {
            if f := key.Flags & 0x101; f == 257 {
                ksks = append(ksks, key)
            }
        }
    }

    group_cdcdnskeys_synced := true
    for signer, keys := range cdses {
        *output = append(*output, fmt.Sprintf("Check sync status of %s CDSes", signer))

        for _, ksk := range ksks {
            found := false
            for _, key := range keys {
                cds := ksk.ToDS(key.DigestType).ToCDS()
                if cds.KeyTag == key.KeyTag && cds.Algorithm == key.Algorithm && cds.Digest == key.Digest {
                    found = true
                    break
                }
            }
            if !found {
                *output = append(*output, fmt.Sprintf("CDS missing for KSK: %s", ksk.PublicKey))
                group_cdcdnskeys_synced = false
            }
        }
    }

    for signer, keys := range cdnskeys {
        *output = append(*output, fmt.Sprintf("Check sync status of %s CDNSKEYs", signer))

        for _, ksk := range ksks {
            found := false
            for _, key := range keys {
                cdnskey := ksk.ToCDNSKEY()
                if cdnskey.Flags == key.Flags && cdnskey.Protocol == key.Protocol && cdnskey.Algorithm == key.Algorithm && cdnskey.PublicKey == key.PublicKey {
                    found = true
                    break
                }
            }
            if !found {
                *output = append(*output, fmt.Sprintf("CDNSKEY missing for KSK: %s", ksk.PublicKey))
                group_cdcdnskeys_synced = false
            }
        }
    }

    if group_cdcdnskeys_synced {
        Config.Set("group-cdcdnskeys-synced:"+args[0], "yes")
    } else {
        Config.Remove("group-cdcdnskeys-synced:" + args[0])
    }

    group_nses_synced := true

    nsmap := make(map[string]*dns.NS)
    for _, rrs := range nses {
        for _, rr := range rrs {
            nsmap[rr.Ns] = rr
        }
    }
    nsset := []*dns.NS{}
    for _, rr := range nsmap {
        nsset = append(nsset, rr)
    }

    for signer, keys := range nses {
        *output = append(*output, fmt.Sprintf("Check sync status of %s NSes", signer))

        for _, ns := range nsset {
            found := false
            for _, key := range keys {
                if ns.Ns == key.Ns {
                    found = true
                    break
                }
            }
            if !found {
                *output = append(*output, fmt.Sprintf("NS missing: %s", ns.Ns))
                group_nses_synced = false
            }
        }
    }
    if group_nses_synced {
        Config.Set("group-nses-synced:"+args[0], "yes")
    } else {
        Config.Remove("group-nses-synced:" + args[0])
    }

    parent := Config.Get("parent:"+args[0], "")
    if parent == "" {
        return fmt.Errorf("No ip|host for parent of %s", args[0])
    }

    *output = append(*output, fmt.Sprintf("Check sync status of parent %s", parent))

    m := new(dns.Msg)
    m.SetQuestion(args[0], dns.TypeDS)
    c := new(dns.Client)
    r, _, err := c.Exchange(m, parent)
    if err != nil {
        return err
    }
    dses := []*dns.DS{}
    for _, a := range r.Answer {
        ds, ok := a.(*dns.DS)
        if !ok {
            continue
        }

        *output = append(*output, fmt.Sprintf("  found DS %d %d %d %s", ds.KeyTag, ds.Algorithm, ds.DigestType, ds.Digest))

        dses = append(dses, ds)
    }

    group_parent_ds_synced := true
    cdsmap := make(map[string]*dns.CDS)
    for _, keys := range cdses {
        for _, key := range keys {
            cdsmap[fmt.Sprintf("%d %d %d %s", key.KeyTag, key.Algorithm, key.DigestType, key.Digest)] = key
        }
    }
    for _, ds := range dses {
        delete(cdsmap, fmt.Sprintf("%d %d %d %s", ds.KeyTag, ds.Algorithm, ds.DigestType, ds.Digest))
    }
    for _, cds := range cdsmap {
        *output = append(*output, fmt.Sprintf("  Missing DS for CDS: %d %d %d %s", cds.KeyTag, cds.Algorithm, cds.DigestType, cds.Digest))
        group_parent_ds_synced = false
    }
    if group_parent_ds_synced {
        Config.Set("group-parent-ds-synced:"+args[0], "yes")
    } else {
        Config.Remove("group-parent-ds-synced:" + args[0])
    }

    m = new(dns.Msg)
    m.SetQuestion(args[0], dns.TypeNS)
    r, _, err = c.Exchange(m, parent)
    if err != nil {
        return err
    }
    for _, a := range r.Ns {
        ns, ok := a.(*dns.NS)
        if !ok {
            continue
        }

        *output = append(*output, fmt.Sprintf("  found NS %s", ns.Ns))

        delete(nsmap, ns.Ns)
    }

    group_parent_ns_synced := true
    for ns, _ := range nsmap {
        *output = append(*output, fmt.Sprintf("  Missing NS: %s", ns))
        group_parent_ns_synced = false
    }
    if group_parent_ns_synced {
        Config.Set("group-parent-ns-synced:"+args[0], "yes")
    } else {
        Config.Remove("group-parent-ns-synced:" + args[0])
    }

    return nil
}
