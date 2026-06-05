package geo

import "testing"

func TestMockClient_Lookup(t *testing.T) {
	c := NewMockClient("US", "New York")
	country, cty, err := c.Lookup("8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if country != "US" {
		t.Errorf("expected US, got %q", country)
	}
	if cty != "New York" {
		t.Errorf("expected New York, got %q", cty)
	}
}

func TestMockClient_Lookup_Localhost(t *testing.T) {
	c := NewMockClient("RU", "Moscow")
	country, city, err := c.Lookup("127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if country != "RU" {
		t.Errorf("expected RU, got %q", country)
	}
	if city != "Moscow" {
		t.Errorf("expected Moscow, got %q", city)
	}
}

func TestClient_Lookup_Localhost(t *testing.T) {
	c := NewClient()
	country, city, err := c.Lookup("127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if country != "RU" || city != "Localhost" {
		t.Errorf("expected RU/Localhost, got %q/%q", country, city)
	}
}

func TestClient_Lookup_EmptyIP(t *testing.T) {
	c := NewClient()
	country, city, err := c.Lookup("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if country != "RU" || city != "Localhost" {
		t.Errorf("expected RU/Localhost, got %q/%q", country, city)
	}
}

func TestClient_Lookup_IPv6Localhost(t *testing.T) {
	c := NewClient()
	country, city, err := c.Lookup("::1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if country != "RU" || city != "Localhost" {
		t.Errorf("expected RU/Localhost, got %q/%q", country, city)
	}
}
