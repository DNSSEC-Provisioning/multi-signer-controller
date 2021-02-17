package main

import (
    "bytes"
    "encoding/base32"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"

    "github.com/google/uuid"
)

func init() {
    Command["test-desec"] = TestDesecCmd
    CommandHelp["test-desec"] = "Test deSEC.io API by inserting and removing a TXT record, requires <zone> <token name>"
}

func TestDesecCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 2 {
        return fmt.Errorf("Missing <zone> <token name>")
    }
    zone := args[0]
    token_name := args[1]

    if zone[len(zone)-1] == '.' {
        zone = zone[:len(zone)-1]
    }

    token := Config.Get("desectoken-"+token_name, "")
    if token == "" {
        return fmt.Errorf("Missing deSEC token %s, use conf-set desectoken-<name> <token>", token_name)
    }

    b := uuid.New()
    id := base32.HexEncoding.EncodeToString(b[:])
    id = strings.ToLower(strings.Replace(id, "=", "", -1))

    rrset := &DesecRRset{
        Subname: id,
        Type:    "TXT",
        Records: []string{"\"test-api\""},
        Ttl:     3600,
    }

    body, err := json.Marshal(rrset)
    if err != nil {
        return err
    }

    *output = append(*output, "Sending POST for creation of "+id)

    req, err := http.NewRequest("POST", fmt.Sprintf("https://desec.io/api/v1/domains/%s/rrsets/", zone), bytes.NewReader(body))
    if err != nil {
        return err
    }
    req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
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
    *output = append(*output, fmt.Sprintf("  %v", rrset))

    *output = append(*output, "Sending DELETE")

    req, err = http.NewRequest("DELETE", fmt.Sprintf("https://desec.io/api/v1/domains/%s/rrsets/%s/TXT/", zone, id), nil)
    if err != nil {
        return err
    }
    req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

    resp, err = client.Do(req)
    if err != nil {
        return err
    }

    if resp.StatusCode == 204 {
        *output = append(*output, "Status 204, deleted OK")
    } else {
        *output = append(*output, "Unknown status returned: "+resp.Status)
    }

    return nil
}
