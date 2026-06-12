package geo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type GeoIPService interface {
	Lookup(ip string) (country, city string, err error)
}

type geoIPResponse struct {
	Status      string `json:"status"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *Client) Lookup(ip string) (string, string, error) {
	if ip == "" || isPrivateIP(ip) {
		return "RU", "Localhost", nil
	}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,countryCode,city", ip)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result geoIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	if result.Status != "success" {
		return "", "", fmt.Errorf("geo lookup failed for %s", ip)
	}

	return result.CountryCode, result.City, nil
}

type MockClient struct {
	Country string
	City    string
}

func NewMockClient(country, city string) *MockClient {
	return &MockClient{Country: country, City: city}
}

func (m *MockClient) Lookup(ip string) (string, string, error) {
	return m.Country, m.City, nil
}

var privateCIDRs = []*net.IPNet{
	mustCIDR("10.0.0.0/8"),
	mustCIDR("172.16.0.0/12"),
	mustCIDR("192.168.0.0/16"),
	mustCIDR("127.0.0.0/8"),
	mustCIDR("::1/128"),
}

func mustCIDR(s string) *net.IPNet {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return cidr
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, cidr := range privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
