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

# SSH is optional — used as fallback for version detection
[servers.home.ssh]
host = "truenas.local"        # defaults to server host
port = 22                     # defaults to 22
username = "root"             # defaults to server username
private_key_path = "~/.ssh/id_ed25519"
host_key_fingerprint = "SHA256:..."
```

Generate an API key in the TrueNAS web UI under **Credentials > API Keys**.

The SSH section is optional. When configured, it provides a fallback transport for version detection. Paths in `private_key_path` support `~` and environment variable expansion. The `host_key_fingerprint` can be obtained with:

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
