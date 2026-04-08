package custom

import "testing"

func TestValidateSubscriptionURLRejectsUnsafeTargets(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "public https ip", url: "https://1.1.1.1/sub", wantErr: false},
		{name: "public http ip", url: "http://8.8.8.8/sub", wantErr: false},
		{name: "localhost", url: "http://localhost:8080/sub", wantErr: true},
		{name: "loopback ip", url: "http://127.0.0.1/sub", wantErr: true},
		{name: "private ip", url: "http://10.0.0.1/sub", wantErr: true},
		{name: "link local ip", url: "http://169.254.169.254/latest/meta-data", wantErr: true},
		{name: "cgnat ip", url: "http://100.64.0.1/sub", wantErr: true},
		{name: "unsupported scheme", url: "ftp://example.com/sub", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateSubscriptionURL(tc.url)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %s", tc.url)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for %s, got %v", tc.url, err)
			}
		})
	}
}

func TestShouldKeepCustomProxyOnValidationFailure(t *testing.T) {
	t.Parallel()

	if !shouldKeepCustomProxyOnValidationFailure("proxyscrape") {
		t.Fatal("expected proxyscrape provider to keep proxies on validation failure")
	}
	if shouldKeepCustomProxyOnValidationFailure("manual") {
		t.Fatal("expected non-proxyscrape provider not to keep proxies on validation failure")
	}
}
