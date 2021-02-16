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

    if !Config.Exists(fmt.Sprintf("signers:%s", args[0])) {
        return fmt.Errorf("group %s has no signers", args[0])
    }

    signers := Config.ListGet(fmt.Sprintf("signers:%s", args[0]))

    for _, signer := range signers {
        ip := Config.Get(fmt.Sprintf("signer:%s", signer), "")
        if ip == "" {
            return fmt.Errorf("No ip|host for signer %s", signer)
        }

        m := new(dns.Msg)
        m.SetQuestion(args[0], dns.TypeDNSKEY)

        c := new(dns.Client)
        in, rtt, err := c.Exchange(m, ip)

        if err != nil {
            return err
        }

        *output = append(*output, fmt.Sprintf("%v", rtt))
        *output = append(*output, in.String())
    }

    return nil
}
