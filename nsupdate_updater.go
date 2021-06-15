package main

import (
    "fmt"
    "time"

    "github.com/miekg/dns"
)

type NsupdateUpdater struct {
}

func init() {
    Updaters["nsupdate"] = &NsupdateUpdater{}
}

func (n *NsupdateUpdater) Update(fqdn, signer string, inserts, removes *[][]dns.RR, output *[]string) error {
    inserts_len := 0
    removes_len := 0
    if inserts != nil {
        for _, insert := range *inserts {
            inserts_len += len(insert)
        }
    }
    if removes != nil {
        for _, remove := range *removes {
            removes_len += len(remove)
        }
    }
    if inserts_len == 0 && removes_len == 0 {
        return fmt.Errorf("Inserts and removes empty, nothing to do")
    }

    ip := Config.Get("signer:"+signer, "")
    if ip == "" {
        return fmt.Errorf("No ip|host for signer %s", signer)
    }

    tsigkey := Config.Get("signer-tsigkey:"+signer, "")
    if tsigkey == "" {
        return fmt.Errorf("Missing signer %s TSIG key %s", signer, tsigkey)
    }

    secret := Config.Get("tsigkey-"+tsigkey, "")
    if secret == "" {
        return fmt.Errorf("Missing TSIG key secret for %s", tsigkey)
    }

    m := new(dns.Msg)
    m.SetUpdate(fqdn)
    if inserts != nil {
        for _, insert := range *inserts {
            m.Insert(insert)
        }
    }
    if removes != nil {
        for _, remove := range *removes {
            m.Remove(remove)
        }
    }
    m.SetTsig(tsigkey+".", dns.HmacSHA256, 300, time.Now().Unix())

    debug := false
    if Config.Get("debug-updater", "") == "yes" {
        debug = true
    }

    *output = append(*output, fmt.Sprintf("nsupdate: Sending inserts %d, removals %d to signer %s", inserts_len, removes_len, signer))
    if debug {
        *output = append(*output, m.String())
    }

    c := new(dns.Client)
    c.TsigSecret = map[string]string{tsigkey + ".": secret}
    in, rtt, err := c.Exchange(m, ip)
    if err != nil {
        return err
    }

    if debug {
        *output = append(*output, in.String())
    }
    *output = append(*output, fmt.Sprintf("nsupdate: Update took %v, rcode %s", rtt, dns.RcodeToString[in.MsgHdr.Rcode]))

    return nil
}

func (d *NsupdateUpdater) RemoveRRset(fqdn, signer string, rrsets [][]dns.RR, output *[]string) error {
    rrsets_len := 0
    for _, rrset := range rrsets {
        rrsets_len += len(rrset)
    }
    if rrsets_len == 0 {
        return fmt.Errorf("rrset(s) is empty, nothing to do")
    }

    ip := Config.Get("signer:"+signer, "")
    if ip == "" {
        return fmt.Errorf("No ip|host for signer %s", signer)
    }

    tsigkey := Config.Get("signer-tsigkey:"+signer, "")
    if tsigkey == "" {
        return fmt.Errorf("Missing signer %s TSIG key %s", signer, tsigkey)
    }

    secret := Config.Get("tsigkey-"+tsigkey, "")
    if secret == "" {
        return fmt.Errorf("Missing TSIG key secret for %s", tsigkey)
    }

    m := new(dns.Msg)
    m.SetUpdate(fqdn)
    for _, rrset := range rrsets {
        m.RemoveRRset(rrset)
    }
    m.SetTsig(tsigkey+".", dns.HmacSHA256, 300, time.Now().Unix())

    debug := false
    if Config.Get("debug-updater", "") == "yes" {
        debug = true
    }

    *output = append(*output, fmt.Sprintf("nsupdate: Sending remove rrset(s) %d to signer %s", rrsets_len, signer))
    if debug {
        *output = append(*output, m.String())
    }

    c := new(dns.Client)
    c.TsigSecret = map[string]string{tsigkey + ".": secret}
    in, rtt, err := c.Exchange(m, ip)
    if err != nil {
        return err
    }

    if debug {
        *output = append(*output, in.String())
    }
    *output = append(*output, fmt.Sprintf("nsupdate: Update took %v, rcode %s", rtt, dns.RcodeToString[in.MsgHdr.Rcode]))

    return nil
}
