package storage

import (
	"path/filepath"
	"testing"
)

func TestAddSubscriptionStoresDefaultProtocol(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	defer store.Close()

	subID, err := store.AddSubscription("proxyscrape", "", "/tmp/proxyscrape.txt", "auto", "proxyscrape", "socks5", 60)
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}

	sub, err := store.GetSubscription(subID)
	if err != nil {
		t.Fatalf("GetSubscription: %v", err)
	}
	if sub.DefaultProtocol != "socks5" {
		t.Fatalf("expected default protocol socks5, got %q", sub.DefaultProtocol)
	}
	if sub.Provider != "proxyscrape" {
		t.Fatalf("expected provider proxyscrape, got %q", sub.Provider)
	}
}

func TestPickWeightedProxyByRollPrefersCustom(t *testing.T) {
	t.Parallel()

	proxies := []Proxy{
		{Address: "free-1", Source: "free"},
		{Address: "custom-1", Source: "custom"},
		{Address: "free-2", Source: "free"},
	}

	if got := pickWeightedProxyByRoll(proxies, 0); got == nil || got.Address != "free-1" {
		t.Fatalf("roll 0 expected free-1, got %+v", got)
	}
	for _, roll := range []int{1, 2, 3, 4} {
		got := pickWeightedProxyByRoll(proxies, roll)
		if got == nil || got.Address != "custom-1" {
			t.Fatalf("roll %d expected custom-1, got %+v", roll, got)
		}
	}
	if got := pickWeightedProxyByRoll(proxies, 5); got == nil || got.Address != "free-2" {
		t.Fatalf("roll 5 expected free-2, got %+v", got)
	}
}
