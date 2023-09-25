
package main

import (
	"reflect"
	"testing"

	"github.com/siderolabs/go-procfs/procfs"
)

func TestDumpEnv(t *testing.T) {
	cfg := &TailscaleConfig{
		AuthKey:   "my-auth-key",
		Hostname:  "my-hostname",
		AcceptDNS: true,
		AuthOnce:  false,
	}
	expectedEnv := "TS_AUTHKEY=my-auth-key\nTS_HOSTNAME=my-hostname\nTS_ACCEPT_DNS=true\nTS_AUTH_ONCE=false\n"
	env, err := dumpEnv(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != expectedEnv {
		t.Errorf("unexpected env:\nexpected: %q\nactual: %q", expectedEnv, env)
	}
}

func TestLoadProcCmdline(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		cmdline  string
		expected TailscaleConfig
	}{
		{
			name: "all fields",
			cmdline: "tailscale.authkey=1234 tailscale.hostname=example.com tailscale.accept-dns=true tailscale.authonce=false",
			expected: TailscaleConfig{
				AuthKey:   "1234",
				Hostname:  "example.com",
				AcceptDNS: true,
				AuthOnce:  false,
			},
		},
		{
			name: "missing fields",
			cmdline: "tailscale.authkey=1234 tailscale.hostname=example.com",
			expected: TailscaleConfig{
				AuthKey:  "1234",
				Hostname: "example.com",
			},
		},
		{
			name: "empty fields",
			cmdline: "tailscale.authkey= tailscale.hostname= tailscale.accept-dns= tailscale.authonce=",
			expected: TailscaleConfig{},
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse cmdline string
			c := procfs.NewCmdline(tc.cmdline)

			// Load config from cmdline
			var cfg TailscaleConfig
			err := loadProcCmdline(&cfg, c)
			if err != nil {
				t.Fatalf("failed to load config from cmdline: %v", err)
			}

			// Compare actual vs expected
			if !reflect.DeepEqual(cfg, tc.expected) {
				t.Errorf("unexpected config:\nexpected: %+v\nactual:   %+v", tc.expected, cfg)
			}
		})
	}
}