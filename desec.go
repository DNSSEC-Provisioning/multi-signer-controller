package main

type DesecRRset struct {
    Created string   `json:"created,omitempty"`
    Domain  string   `json:"domain,omitempty"`
    Subname string   `json:"subname,omitempty"`
    Name    string   `json:"name,omitempty"`
    Type    string   `json:"type,omitempty"`
    Records []string `json:"records,omitempty"`
    Ttl     int      `json:"ttl,omitempty"`
    Touched string   `json:"touched,omitempty"`
}
