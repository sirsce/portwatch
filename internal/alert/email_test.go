package alert

import (
	"io"
	"net"
	"strings"
	"testing"
)

// startFakeSMTP starts a minimal fake SMTP server that accepts one connection.
// It returns the listener address and a channel that receives the raw data sent by the client.
func startFakeSMTP(t *testing.T) (addr string, msgCh <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start fake SMTP server: %v", err)
	}

	ch := make(chan string, 1)
	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		var buf strings.Builder
		conn.Write([]byte("220 localhost SMTP\r\n"))
		tmp := make([]byte, 4096)
		for {
			n, err := conn.Read(tmp)
			if n > 0 {
				buf.Write(tmp[:n])
				data := buf.String()
				if strings.Contains(data, "EHLO") || strings.Contains(data, "HELO") {
					conn.Write([]byte("250 OK\r\n"))
				}
				if strings.Contains(data, "MAIL FROM") {
					conn.Write([]byte("250 OK\r\n"))
				}
				if strings.Contains(data, "RCPT TO") {
					conn.Write([]byte("250 OK\r\n"))
				}
				if strings.Contains(data, "DATA") && !strings.Contains(data, "\r\n.\r\n") {
					conn.Write([]byte("354 Start input\r\n"))
				}
				if strings.Contains(data, "\r\n.\r\n") {
					conn.Write([]byte("250 OK\r\n"))
				}
				if strings.Contains(data, "QUIT") {
					conn.Write([]byte("221 Bye\r\n"))
					ch <- buf.String()
					return
				}
			}
			if err == io.EOF || err != nil {
				ch <- buf.String()
				return
			}
		}
	}()

	return ln.Addr().String(), ch
}

func TestEmailNotifier_Notify_Success(t *testing.T) {
	addr, msgCh := startFakeSMTP(t)

	host, port := splitHostPort(t, addr)
	notifier := NewEmailNotifier(host, port, "", "", "portwatch@localhost", []string{"admin@localhost"})

	err := notifier.Notify("Port 8080 down", "Service on port 8080 is unreachable.")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	msg := <-msgCh
	if !strings.Contains(msg, "Port 8080 down") {
		t.Errorf("expected subject in SMTP data, got: %s", msg)
	}
}

func TestEmailNotifier_Notify_InvalidHost(t *testing.T) {
	notifier := NewEmailNotifier("invalid.host.local", 25, "", "", "from@local", []string{"to@local"})
	err := notifier.Notify("test", "body")
	if err == nil {
		t.Fatal("expected error for invalid host, got nil")
	}
}

func splitHostPort(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("splitHostPort: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return host, port
}
