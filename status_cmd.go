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

            *output = append(*output, fmt.Sprintf("%s: found DNSKEY %d %d %d %s", signer, dnskey.Flags, dnskey.Protocol, dnskey.Algorithm, dnskey.PublicKey))

            dnskeys[signer] = append(dnskeys[signer], dnskey)
        }
    }

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
                    }
                }
            }
        }
    }

    return nil
}
