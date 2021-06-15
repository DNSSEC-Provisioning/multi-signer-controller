package main

import (
    "fmt"
    "time"

    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"

    "github.com/miekg/dns"
)

func init() {
    Command["sync-dnskey"] = SyncDnskeyCmd
    Command["sync-cdcdnskeys"] = SyncCdcdnskeysCmd
    Command["sync-ns"] = SyncNsCmd

    CommandHelp["sync-dnskey"] = "Sync DNSKEYs between signers in a group, requires <fqdn>"
    CommandHelp["sync-cdcdnskeys"] = "Create CDS/CDNSKEYs from DNSKEYs and sync them between signers in a group, requires <fqdn>"
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

        csk := Config.Get("signer-csk:"+signer, "")

        for _, a := range r.Answer {
            dnskey, ok := a.(*dns.DNSKEY)
            if !ok {
                continue
            }

            dnskeys[signer] = append(dnskeys[signer], dnskey)

            if f := dnskey.Flags & 0x101; f == 256 || csk != "" {
                Config.SetIfNotExists("dnskey-origin:"+fmt.Sprintf("%d-%d-%s", dnskey.Protocol, dnskey.Algorithm, dnskey.PublicKey), signer)
            }
        }
    }

    for signer, keys := range dnskeys {
        csk := Config.Get("signer-csk:"+signer, "")

        leaving := Config.Get("signer-leaving:"+signer, "")
        if leaving != "" {
            *output = append(*output, fmt.Sprintf("Signer %s is leaving, removing it's DNSKEYs from others", signer))
            for _, key := range keys {
                if f := key.Flags & 0x101; f == 256 || csk != "" {
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

                        ip := Config.Get("signer:"+osigner, "")
                        if ip == "" {
                            *output = append(*output, fmt.Sprintf("No ip|host for signer %s(???), can't sync it", osigner))
                            continue
                        }
                        stype := Config.Get("signer-type:"+osigner, "nsupdate")

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
                            switch stype {
                            case "nsupdate":
                                tsigkey := Config.Get("signer-tsigkey:"+osigner, "")
                                if tsigkey == "" {
                                    *output = append(*output, fmt.Sprintf("Missing signer %s TSIG key, can't sync %s keys to it", osigner, signer))
                                    continue
                                }

                                secret := Config.Get("tsigkey-"+tsigkey, "")
                                if secret == "" {
                                    *output = append(*output, fmt.Sprintf("Missing TSIG key %s, can't sync %s keys to %s", tsigkey, signer, osigner))
                                    continue
                                }

                                m := new(dns.Msg)
                                m.SetUpdate(args[0])
                                m.Remove([]dns.RR{key})
                                m.SetTsig(tsigkey+".", dns.HmacSHA256, 300, time.Now().Unix())

                                *output = append(*output, m.String())

                                c := new(dns.Client)
                                c.TsigSecret = map[string]string{tsigkey + ".": secret}
                                in, rtt, err := c.Exchange(m, ip)
                                if err != nil {
                                    return err
                                }

                                *output = append(*output, fmt.Sprintf("Remove took %v", rtt))
                                *output = append(*output, in.String())

                                *output = append(*output, fmt.Sprintf("  Removed DNSKEY from %s", osigner))
                                break

                            default:
                                return fmt.Errorf("Unknown signer type %s for %s", stype, osigner)
                            }
                        }
                    }
                }
            }
            continue
        }

        *output = append(*output, fmt.Sprintf("Syncing %s DNSKEYs", signer))

        for _, key := range keys {
            if f := key.Flags & 0x101; f == 256 || csk != "" {
                *output = append(*output, fmt.Sprintf("- %s", key.PublicKey))

                for osigner, okeys := range dnskeys {
                    if osigner == signer {
                        continue
                    }
                    if leaving := Config.Get("signer-leaving:"+osigner, ""); leaving != "" {
                        continue
                    }

                    ip := Config.Get("signer:"+osigner, "")
                    if ip == "" {
                        *output = append(*output, fmt.Sprintf("No ip|host for signer %s(???), can't sync it", osigner))
                        continue
                    }
                    stype := Config.Get("signer-type:"+osigner, "nsupdate")

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
                        switch stype {
                        case "nsupdate":
                            tsigkey := Config.Get("signer-tsigkey:"+osigner, "")
                            if tsigkey == "" {
                                *output = append(*output, fmt.Sprintf("Missing signer %s TSIG key, can't sync %s keys to it", osigner, signer))
                                continue
                            }

                            secret := Config.Get("tsigkey-"+tsigkey, "")
                            if secret == "" {
                                *output = append(*output, fmt.Sprintf("Missing TSIG key %s, can't sync %s keys to %s", tsigkey, signer, osigner))
                                continue
                            }

                            m := new(dns.Msg)
                            m.SetUpdate(args[0])
                            m.Insert([]dns.RR{key})
                            m.SetTsig(tsigkey+".", dns.HmacSHA256, 300, time.Now().Unix())

                            *output = append(*output, m.String())

                            c := new(dns.Client)
                            c.TsigSecret = map[string]string{tsigkey + ".": secret}
                            in, rtt, err := c.Exchange(m, ip)
                            if err != nil {
                                return err
                            }

                            *output = append(*output, fmt.Sprintf("Insert took %v", rtt))
                            *output = append(*output, in.String())

                            *output = append(*output, fmt.Sprintf("  Added DNSKEY to %s", osigner))
                            break

                        case "desec":
                            token := Config.Get("signer-desec:"+osigner, "")
                            if token == "" {
                                *output = append(*output, fmt.Sprintf("Missing signer %s deSEC token, can't sync %s keys to it", osigner, signer))
                                continue
                            }

                            secret := Config.Get("desectoken-"+token, "")
                            if secret == "" {
                                *output = append(*output, fmt.Sprintf("Missing deSEC token %s, can't sync %s keys to %s", token, signer, osigner))
                                continue
                            }

                            zone := args[0]
                            if zone[len(zone)-1] == '.' {
                                zone = zone[:len(zone)-1]
                            }

                            rrset := &DesecRRset{
                                Subname: "",
                                Type:    "DNSKEY",
                                Records: []string{fmt.Sprintf("%d %d %d %s", key.Flags, key.Protocol, key.Algorithm, key.PublicKey)},
                                Ttl:     3600,
                            }

                            *output = append(*output, "POST:")
                            *output = append(*output, fmt.Sprintf("  %v", rrset))

                            body, err := json.Marshal(rrset)
                            if err != nil {
                                return err
                            }
                            *output = append(*output, string(body))

                            req, err := http.NewRequest("POST", fmt.Sprintf("https://desec.io/api/v1/domains/%s/rrsets/", zone), bytes.NewReader(body))
                            if err != nil {
                                return err
                            }
                            req.Header.Add("Authorization", fmt.Sprintf("Token %s", secret))
                            req.Header.Add("Content-Type", "application/json")

                            client := &http.Client{}
                            resp, err := client.Do(req)
                            if err != nil {
                                return err
                            }
                            defer resp.Body.Close()
                            body, err = ioutil.ReadAll(resp.Body)
                            if err != nil {
                                return err
                            }

                            rrset = &DesecRRset{}
                            json.Unmarshal(body, &rrset)
                            *output = append(*output, "Response:")
                            *output = append(*output, resp.Status)
                            *output = append(*output, fmt.Sprintf("  %v", rrset))
                            break

                        default:
                            return fmt.Errorf("Unknown signer type %s for %s", stype, osigner)
                        }
                    } else {
                        *output = append(*output, fmt.Sprintf("  Key exist in %s", osigner))
                    }
                }
            }
        }
    }

    return nil
}

func SyncCdcdnskeysCmd(args []string, remote bool, output *[]string) error {
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

    m := new(dns.Msg)
    m.SetUpdate(args[0])
    m.Insert(cdses)
    m.Insert(cdnskeys)

    *output = append(*output, m.String())

    for _, signer := range signers {
        if leaving := Config.Get("signer-leaving:"+signer, ""); leaving != "" {
            continue
        }

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

            *output = append(*output, fmt.Sprintf("  Added CD/CDNSKEYs to %s", signer))
            break

        default:
            return fmt.Errorf("Unknown signer type %s for %s", stype, signer)
        }
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
            rr.Hdr = dns.RR_Header{Name: args[0], Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 300}
            rr.Ns = ns
            nsset = append(nsset, rr)
        }
    }

    m := new(dns.Msg)
    m.SetUpdate(args[0])
    m.Insert(nsset)
    if len(nsrem) > 0 {
        m.Remove(nsrem)
    }

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

            *output = append(*output, fmt.Sprintf("  Added NSes to %s", signer))
            break

        default:
            return fmt.Errorf("Unknown signer type %s for %s", stype, signer)
        }
    }

    return nil
}
