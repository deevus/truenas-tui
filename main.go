package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/vxfw"
	"github.com/deevus/truenas-go"
	"github.com/deevus/truenas-go/client"
	"github.com/deevus/truenas-tui/app"
	"github.com/deevus/truenas-tui/config"
	"github.com/deevus/truenas-tui/internal"
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

	if serverCfg.SSH == nil {
		fmt.Fprintf(os.Stderr, "Error: server %q requires [servers.%s.ssh] config for connection setup\n", serverName, serverName)
		os.Exit(1)
	}

	privateKey, err := os.ReadFile(serverCfg.SSH.PrivateKeyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SSH private key %s: %v\n", serverCfg.SSH.PrivateKeyPath, err)
		os.Exit(1)
	}

	sshHost := serverCfg.SSH.Host
	if sshHost == "" {
		sshHost = serverCfg.Host
	}

	sshClient, err := client.NewSSHClient(&client.SSHConfig{
		Host:               sshHost,
		Port:               serverCfg.SSH.Port,
		User:               serverCfg.SSH.User,
		PrivateKey:         string(privateKey),
		HostKeyFingerprint: serverCfg.SSH.HostKeyFingerprint,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating SSH client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	wsClient, err := client.NewWebSocketClient(client.WebSocketConfig{
		Host:               serverCfg.Host,
		Port:               serverCfg.Port,
		Username:           serverCfg.Username,
		APIKey:             serverCfg.APIKey,
		InsecureSkipVerify: serverCfg.InsecureSkipVerify,
		Fallback:           sshClient,
	})
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

	vxApp, err := vxfw.NewApp(vaxis.Options{})
	if err != nil {
		log.Fatal(err)
	}

	if err := vxApp.Run(root); err != nil {
		log.Fatal(err)
	}
}
