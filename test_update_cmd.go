package main

import (
    "encoding/base32"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/miekg/dns"
)

func init() {
    Command["test-update"] = TestUpdateCmd
    CommandHelp["test-update"] = "Test if dynamic update work by inserting and removing a TXT record, requires <DNS server> <zone> <TSIG key>"
}

func TestUpdateCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 3 {
        return fmt.Errorf("Missing <DNS server> <zone> <TSIG key>")
    }
    server := args[0]
    zone := args[1]
    tsigkey := args[2]

    secret := Config.Get("tsigkey-"+tsigkey, "")
    if secret == "" {
        return fmt.Errorf("Missing TSIG key %s, use conf-set tsigkey-<name> <secret>", tsigkey)
    }

    b := uuid.New()
    id := base32.HexEncoding.EncodeToString(b[:])
    id = strings.ToLower(strings.Replace(id, "=", "", -1))

    rr := new(dns.TXT)
    rr.Hdr = dns.RR_Header{Name: id + "." + zone, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 3600}
    rr.Txt = []string{"test-update"}
    rrs := []dns.RR{rr}

    m := new(dns.Msg)
    m.SetUpdate(zone)
    m.Insert(rrs)
    m.SetTsig(tsigkey+".", dns.HmacSHA256, 300, time.Now().Unix())

    *output = append(*output, m.String())

    c := new(dns.Client)
    c.TsigSecret = map[string]string{"update.": secret}
    in, rtt, err := c.Exchange(m, server)
    if err != nil {
        return err
    }

    *output = append(*output, fmt.Sprintf("Insert took %v", rtt))
    *output = append(*output, in.String())

    m = new(dns.Msg)
    m.SetUpdate(zone)
    m.Remove(rrs)
    m.SetTsig("update.", dns.HmacSHA256, 300, time.Now().Unix())

    *output = append(*output, m.String())

    in, rtt, err = c.Exchange(m, server)
    if err != nil {
        return err
    }

    *output = append(*output, fmt.Sprintf("Remove took %v", rtt))
    *output = append(*output, in.String())

    return nil
}
