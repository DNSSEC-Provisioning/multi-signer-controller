package main

import (
    "fmt"
    "time"

    "github.com/miekg/dns"
)

func init() {
    Command["wait-ds"] = WaitDsCmd
    Command["wait-ns"] = WaitNsCmd

    CommandHelp["wait-ds"] = "Gather DNSKEYs and DSes for a group, use largest TTL * 2 and set a waiting time, requires <fqdn>"
    CommandHelp["wait-ns"] = "Gather NSes for a group, use largest TTL * 2 and set a waiting time, requires <fqdn>"
}

func WaitDsCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    wait_until := Config.Get("group-wait-ds:"+args[0], "")
    if wait_until != "" {
        until, err := time.Parse(time.RFC3339, wait_until)
        if err != nil {
            return err
        }

        *output = append(*output, fmt.Sprintf("Wait until %s (%s)", until.String(), time.Until(until).String()))

        return nil
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    var ttl uint32

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
        *output = append(*output, r.String())

        for _, a := range r.Answer {
            dnskey, ok := a.(*dns.DNSKEY)
            if !ok {
                continue
            }

            if dnskey.Header().Ttl > ttl {
                ttl = dnskey.Header().Ttl
            }
        }
    }

    parent := Config.Get("parent:"+args[0], "")
    if parent == "" {
        return fmt.Errorf("No ip|host for parent of %s", args[0])
    }

    m := new(dns.Msg)
    m.SetQuestion(args[0], dns.TypeDS)
    c := new(dns.Client)
    r, _, err := c.Exchange(m, parent)
    if err != nil {
        return err
    }
    *output = append(*output, r.String())

    for _, a := range r.Answer {
        ds, ok := a.(*dns.DS)
        if !ok {
            continue
        }

        if ds.Header().Ttl > ttl {
            ttl = ds.Header().Ttl
        }
    }

    *output = append(*output, fmt.Sprintf("Largest TTL %d", ttl))

    until := time.Now().Add((time.Duration(ttl*2) * time.Second))

    *output = append(*output, fmt.Sprintf("Wait until %s (%s)", until.String(), time.Until(until).String()))

    Config.Set("group-wait-ds:"+args[0], until.Format(time.RFC3339))

    return nil
}

func WaitNsCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    wait_until := Config.Get("group-wait-ns:"+args[0], "")
    if wait_until != "" {
        until, err := time.Parse(time.RFC3339, wait_until)
        if err != nil {
            return err
        }

        *output = append(*output, fmt.Sprintf("Wait until %s (%s)", until.String(), time.Until(until).String()))

        return nil
    }

    if !Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet("signers:" + args[0])

    var ttl uint32

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
        *output = append(*output, r.String())

        for _, a := range r.Answer {
            ns, ok := a.(*dns.NS)
            if !ok {
                continue
            }

            if ns.Header().Ttl > ttl {
                ttl = ns.Header().Ttl
            }
        }
    }

    parent := Config.Get("parent:"+args[0], "")
    if parent == "" {
        return fmt.Errorf("No ip|host for parent of %s", args[0])
    }

    m := new(dns.Msg)
    m.SetQuestion(args[0], dns.TypeNS)
    c := new(dns.Client)
    r, _, err := c.Exchange(m, parent)
    if err != nil {
        return err
    }
    *output = append(*output, r.String())

    for _, a := range r.Ns {
        ns, ok := a.(*dns.NS)
        if !ok {
            continue
        }

        if ns.Header().Ttl > ttl {
            ttl = ns.Header().Ttl
        }
    }

    *output = append(*output, fmt.Sprintf("Largest TTL %d", ttl))

    until := time.Now().Add((time.Duration(ttl*2) * time.Second))

    *output = append(*output, fmt.Sprintf("Wait until %s (%s)", until.String(), time.Until(until).String()))

    Config.Set("group-wait-ns:"+args[0], until.Format(time.RFC3339))

    return nil
}
