package main

import (
    "fmt"

    "github.com/miekg/dns"
)

type DesecUpdater struct {
}

func init() {
    Updaters["desec"] = &DesecUpdater{}
}

func (d *DesecUpdater) Update(fqdn, signer string, inserts, removes *[][]dns.RR, output *[]string) error {
    return fmt.Errorf("deSEC.io support not implemented")
}

func (d *DesecUpdater) RemoveRRset(fqdn, signer string, rrsets [][]dns.RR, output *[]string) error {
    return fmt.Errorf("deSEC.io support not implemented")
}

// token := Config.Get("signer-desec:"+osigner, "")
// if token == "" {
//     *output = append(*output, fmt.Sprintf("Missing signer %s deSEC token, can't sync %s keys to it", osigner, signer))
//     continue
// }
//
// secret := Config.Get("desectoken-"+token, "")
// if secret == "" {
//     *output = append(*output, fmt.Sprintf("Missing deSEC token %s, can't sync %s keys to %s", token, signer, osigner))
//     continue
// }
//
// zone := args[0]
// if zone[len(zone)-1] == '.' {
//     zone = zone[:len(zone)-1]
// }
//
// rrset := &DesecRRset{
//     Subname: "",
//     Type:    "DNSKEY",
//     Records: []string{fmt.Sprintf("%d %d %d %s", key.Flags, key.Protocol, key.Algorithm, key.PublicKey)},
//     Ttl:     ttl,
// }
//
// *output = append(*output, "POST:")
// *output = append(*output, fmt.Sprintf("  %v", rrset))
//
// body, err := json.Marshal(rrset)
// if err != nil {
//     return err
// }
// *output = append(*output, string(body))
//
// req, err := http.NewRequest("POST", fmt.Sprintf("https://desec.io/api/v1/domains/%s/rrsets/", zone), bytes.NewReader(body))
// if err != nil {
//     return err
// }
// req.Header.Add("Authorization", fmt.Sprintf("Token %s", secret))
// req.Header.Add("Content-Type", "application/json")
//
// client := &http.Client{}
// resp, err := client.Do(req)
// if err != nil {
//     return err
// }
// defer resp.Body.Close()
// body, err = ioutil.ReadAll(resp.Body)
// if err != nil {
//     return err
// }
//
// rrset = &DesecRRset{}
// json.Unmarshal(body, &rrset)
// *output = append(*output, "Response:")
// *output = append(*output, resp.Status)
// *output = append(*output, fmt.Sprintf("  %v", rrset))
// break
