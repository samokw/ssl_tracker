package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"
)

// Creating a basic program that will check the exipry of a predefined sercer
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	ctx := context.WithValue(context.Background(), "logger", logger)
	ctx.Done()





	hostname := "courselink.uoguelph.ca"
	// this is the https port
	port := "443"

	// Create a tcp connection to the server.
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostname, port), 3*time.Second)
	if err != nil {
		slog.Error("failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := tls.Client(conn, &tls.Config{
		ServerName: hostname,
	})

	err = client.Handshake()
	if err != nil {
		slog.Error("failed to complete TLS handshake: %v", err)
	}
	defer client.Close()

	certs := client.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		slog.Any("error", "no certificate found")
	}

	cert := certs[0]
	fmt.Printf("The SSL certificate for %s expires on: %s\n", hostname, cert.NotAfter.Format(time.RFC1123))
	remainingDays := time.Until(cert.NotAfter).Hours() / 24
	fmt.Printf("The certificate will expire in approximately %.0f days.\n", remainingDays)
}
