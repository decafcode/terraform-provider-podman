package api

type NetworkJson struct {
	DnsEnabled bool   `json:"dns_enabled"`
	Id         string `json:"id"`
	Internal   bool   `json:"internal"`
	Name       string `json:"name"`
}
