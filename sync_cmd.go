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

    CommandHelp["sync-dnskey"] = "Sync DNSKEYs between signers in a group, requires <fqdn>"
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
        }
    }

    for signer, keys := range dnskeys {
        *output = append(*output, fmt.Sprintf("Syncing %s DNSKEYs", signer))

        csk := Config.Get("signer-csk:"+signer, "")

        for _, key := range keys {
            if f := key.Flags & 0x101; f == 256 || csk != "" {
                *output = append(*output, fmt.Sprintf("- %s", key.PublicKey))

                for osigner, okeys := range dnskeys {
                    if osigner == signer {
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
