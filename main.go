package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-go/client"
	"github.com/deevus/truenas-tui/app"
	"github.com/deevus/truenas-tui/config"
	"github.com/deevus/truenas-tui/internal"
	"golang.org/x/crypto/ssh"
)

func main() {
	serverFlag := flag.String("server", "", "server profile name from config")
	configFlag := flag.String("config", config.DefaultPath(), "path to config file")
	flag.Parse()

	cfg, err := config.LoadFrom(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	serverName := *serverFlag
	if serverName == "" {
		names := cfg.ServerNames()
		if len(names) == 1 {
			serverName = names[0]
		} else {
			fmt.Fprintf(os.Stderr, "Multiple servers configured. Use --server flag.\nAvailable: %v\n", names)
			os.Exit(1)
		}
	}

	serverCfg, ok := cfg.Servers[serverName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: server %q not found in config\n", serverName)
		os.Exit(1)
	}

	ctx := context.Background()

	wsCfg := client.WebSocketConfig{
		Host:               serverCfg.Host,
		Port:               serverCfg.Port,
		Username:           serverCfg.Username,
		APIKey:             serverCfg.APIKey,
		InsecureSkipVerify: serverCfg.InsecureSkipVerify,
	}

	if serverCfg.SSH != nil {
		sshHost := serverCfg.SSH.Host
		if sshHost == "" {
			sshHost = serverCfg.Host
		}

		if serverCfg.SSH.HostKeyFingerprint == "" {
			fingerprint, err := scanHostKey(sshHost, serverCfg.SSH.Port)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: host_key_fingerprint is required for SSH.\n")
				fmt.Fprintf(os.Stderr, "Could not auto-detect: %v\n", err)
				fmt.Fprintf(os.Stderr, "Get it with: ssh-keyscan -p %d %s 2>/dev/null | ssh-keygen -lf -\n", serverCfg.SSH.Port, sshHost)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error: host_key_fingerprint is required for SSH.\n")
			fmt.Fprintf(os.Stderr, "Detected fingerprint for %s:\n\n", sshHost)
			fmt.Fprintf(os.Stderr, "  host_key_fingerprint = %q\n\n", fingerprint)
			fmt.Fprintf(os.Stderr, "Add this to [servers.%s.ssh] in your config.\n", serverName)
			os.Exit(1)
		}

		privateKey, err := os.ReadFile(serverCfg.SSH.PrivateKeyPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading SSH private key %s: %v\n", serverCfg.SSH.PrivateKeyPath, err)
			os.Exit(1)
		}

		sshClient, err := client.NewSSHClient(&client.SSHConfig{
			Host:               sshHost,
			Port:               serverCfg.SSH.Port,
			User:               serverCfg.SSH.Username,
			PrivateKey:         string(privateKey),
			HostKeyFingerprint: serverCfg.SSH.HostKeyFingerprint,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating SSH client: %v\n", err)
			os.Exit(1)
		}
		wsCfg.Fallback = sshClient
	}

	wsClient, err := client.NewWebSocketClient(wsCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	if err := wsClient.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to %s: %v\n", serverCfg.Host, err)
		os.Exit(1)
	}

	version := wsClient.Version()
	svc := internal.NewServices(
		truenas.NewDatasetService(wsClient, version),
		truenas.NewSnapshotService(wsClient, version),
	)

	root := app.New(svc, serverName)

	if err := root.LoadActiveView(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}

	vxApp, err := vxfw.NewApp(vaxis.Options{})
	if err != nil {
		log.Fatal(err)
	}

	if err := vxApp.Run(root); err != nil {
		log.Fatal(err)
	}
}

// scanHostKey connects to an SSH server and returns the host key fingerprint.
func scanHostKey(host string, port int) (string, error) {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	var fingerprint string
	cfg := &ssh.ClientConfig{
		User: "probe",
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			fingerprint = ssh.FingerprintSHA256(key)
			return nil
		},
		Timeout: 5 * time.Second,
	}
	conn, err := ssh.Dial("tcp", addr, cfg)
	if conn != nil {
		conn.Close()
	}
	if fingerprint != "" {
		return fingerprint, nil
	}
	return "", fmt.Errorf("could not connect to %s: %v", addr, err)
}
