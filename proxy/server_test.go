package proxy

import (
	"net/http"
	"testing"
)

func TestStripHopByHopHeaders(t *testing.T) {
	header := http.Header{}
	header.Set("Proxy-Authorization", "Basic secret")
	header.Set("Proxy-Authenticate", "Basic")
	header.Set("Proxy-Connection", "keep-alive")
	header.Set("Connection", "keep-alive")
	header.Set("Keep-Alive", "timeout=5")
	header.Set("Te", "trailers")
	header.Set("Trailer", "X-Test")
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Upgrade", "websocket")
	header.Set("X-Test", "keep")

	stripHopByHopHeaders(header)

	for _, key := range []string{
		"Proxy-Authorization",
		"Proxy-Authenticate",
		"Proxy-Connection",
		"Connection",
		"Keep-Alive",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	} {
		if header.Get(key) != "" {
			t.Fatalf("expected header %s to be removed", key)
		}
	}
	if got := header.Get("X-Test"); got != "keep" {
		t.Fatalf("expected non hop-by-hop header to remain, got %q", got)
	}
}
