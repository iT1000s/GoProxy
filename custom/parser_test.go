package custom

import (
	"encoding/base64"
	"testing"
)

func TestParseAutoDetectPlainProxyList(t *testing.T) {
	t.Parallel()

	data := []byte(`
# proxyscrape export
1.1.1.1:80
socks5://2.2.2.2:1080
https://3.3.3.3:443
invalid-line
`)

	nodes, err := Parse(data, "auto")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	if nodes[0].Type != "http" || nodes[0].Server != "1.1.1.1" || nodes[0].Port != 80 {
		t.Fatalf("unexpected first node: %+v", nodes[0])
	}
	if nodes[1].Type != "socks5" || nodes[1].Server != "2.2.2.2" || nodes[1].Port != 1080 {
		t.Fatalf("unexpected second node: %+v", nodes[1])
	}
	if nodes[2].Type != "http" || nodes[2].Server != "3.3.3.3" || nodes[2].Port != 443 {
		t.Fatalf("unexpected third node: %+v", nodes[2])
	}
}

func TestParseAutoDetectBase64PlainProxyList(t *testing.T) {
	t.Parallel()

	raw := "8.8.8.8:8080\nsocks4://9.9.9.9:1080\n"
	data := []byte(base64.StdEncoding.EncodeToString([]byte(raw)))

	nodes, err := Parse(data, "auto")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Type != "http" || nodes[0].Server != "8.8.8.8" || nodes[0].Port != 8080 {
		t.Fatalf("unexpected first node: %+v", nodes[0])
	}
	if nodes[1].Type != "socks5" || nodes[1].Server != "9.9.9.9" || nodes[1].Port != 1080 {
		t.Fatalf("unexpected second node: %+v", nodes[1])
	}
}

func TestParseAutoDetectPlainProxyListWithDefaultProtocol(t *testing.T) {
	t.Parallel()

	data := []byte("4.4.4.4:1080\n5.5.5.5:2080\n")

	nodes, err := ParseWithOptions(data, "auto", ParseOptions{DefaultProtocol: "socks5"})
	if err != nil {
		t.Fatalf("ParseWithOptions returned error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	for i, node := range nodes {
		if node.Type != "socks5" {
			t.Fatalf("node %d expected socks5, got %+v", i, node)
		}
	}
}

func TestParseAutoDetectRejectsInvalidText(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte("just some random text"), "auto")
	if err == nil {
		t.Fatal("expected error for invalid content")
	}
}
