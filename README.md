# truenas-tui

A terminal UI for managing TrueNAS servers.

## Install

```bash
go install github.com/deevus/truenas-tui@latest
```

## Configure

Create `~/.config/truenas-tui/config.toml`:

```toml
[servers.home]
host = "truenas.local"
port = 443
username = "admin"
api_key = "1-your-api-key"

[servers.home.ssh]
host = "truenas.local"
port = 22
user = "root"
private_key_path = "~/.ssh/id_ed25519"
host_key_fingerprint = "SHA256:..."
```

Generate an API key in the TrueNAS web UI under **Credentials > API Keys**.

The SSH section is required for the WebSocket client to detect the TrueNAS version on connect. The `host_key_fingerprint` can be obtained with:

```bash
ssh-keyscan truenas.local 2>/dev/null | ssh-keygen -lf -
```

## Usage

```bash
# Single server (auto-selected)
truenas-tui

# Multiple servers — specify which one
truenas-tui --server home

# Custom config path
truenas-tui --config /path/to/config.toml
```

## Keybindings

| Key | Action |
|-----|--------|
| `1` / `2` / `3` | Switch tabs (Pools / Datasets / Snapshots) |
| `Tab` / `Shift+Tab` | Next / previous tab |
| `j` / `k` or arrows | Navigate list |
| `q` | Quit |
