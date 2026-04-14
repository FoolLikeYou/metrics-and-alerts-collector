package main

import "testing"

func TestResolveServerAddr_SERVERPORTOverridesTemplate8080(t *testing.T) {
	t.Setenv("ADDRESS", "")
	t.Setenv("SERVER_PORT", "54321")

	cases := []struct {
		flag string
		want string
	}{
		{"localhost:8080", "localhost:54321"},
		{"http://localhost:8080", "localhost:54321"},
		{"https://127.0.0.1:8080", "localhost:54321"},
		{"http://[::1]:8080", "localhost:54321"},
	}
	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			if got := resolveServerAddr(tc.flag); got != tc.want {
				t.Fatalf("resolveServerAddr(%q) = %q, want %q", tc.flag, got, tc.want)
			}
		})
	}
}

func TestResolveServerAddr_ADDRESSWinsOverSERVERPORT(t *testing.T) {
	t.Setenv("ADDRESS", "localhost:11111")
	t.Setenv("SERVER_PORT", "22222")
	if got := resolveServerAddr("http://localhost:8080"); got != "localhost:11111" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveServerAddr_non8080FlagUnchangedWithoutAddress(t *testing.T) {
	t.Setenv("ADDRESS", "")
	t.Setenv("SERVER_PORT", "33333")
	if got := resolveServerAddr("localhost:9090"); got != "localhost:9090" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveServerAddr_emptyFlagUsesEnv(t *testing.T) {
	t.Setenv("ADDRESS", "")
	t.Setenv("SERVER_PORT", "44444")
	if got := resolveServerAddr(""); got != "localhost:44444" {
		t.Fatalf("got %q", got)
	}
}
