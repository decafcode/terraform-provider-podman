package api

type NetworkJson struct {
	DnsEnabled  bool   `json:"dns_enabled"`
	Id          string `json:"id"`
	Internal    bool   `json:"internal"`
	Ipv6Enabled bool   `json:"ipv6_enabled"`
	Name        string `json:"name"`
}
