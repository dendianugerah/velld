package mail

import (
	"errors"
	"io"
	"net"
	"net/textproto"
	"testing"
)

func TestSendEmailSMTPAddress(t *testing.T) {
	for _, tc := range []struct {
		name    string
		network string
		host    string
		address string
	}{
		{"IPv4", "tcp4", "127.0.0.1", "127.0.0.1:0"},
		{"IPv6", "tcp6", "::1", "[::1]:0"},
		{"empty host", "tcp4", "", "127.0.0.1:0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen(tc.network, tc.address)
			if err != nil {
				if tc.network == "tcp6" {
					t.Skipf("IPv6 loopback unavailable: %v", err)
				}
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = listener.Close() })
			go func() {
				conn, err := listener.Accept()
				if err == nil {
					defer conn.Close()
					_, _ = io.WriteString(conn, "421 Service unavailable\r\n")
				}
			}()

			err = SendEmail(&SMTPConfig{Host: tc.host, Port: listener.Addr().(*net.TCPAddr).Port}, &Message{})
			var smtpErr *textproto.Error
			if !errors.As(err, &smtpErr) || smtpErr.Code != 421 {
				t.Fatalf("expected an SMTP response, got %v", err)
			}
		})
	}
}
