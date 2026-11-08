package mail

import "testing"

func TestRedactAddresses(t *testing.T) {
	cases := map[string]string{
		"550 5.1.1 <ramon@gerard.ca>: Recipient address rejected": "550 5.1.1 <redacted>: Recipient address rejected",
		"mailbox ramon@gerard.ca unavailable":                     "mailbox redacted unavailable",
		"421 service not available":                               "421 service not available",
	}
	for in, want := range cases {
		if got := redactAddresses(in); got != want {
			t.Errorf("redactAddresses(%q) = %q, want %q", in, got, want)
		}
	}
}
