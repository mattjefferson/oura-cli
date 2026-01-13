# oura

Read-only CLI for the Oura Cloud API (v2) using OAuth2.

## Requirements

- Oura API application (OAuth2 client id/secret)
- Redirect URI registered in your Oura app (default: `http://127.0.0.1:8797/callback`)

## Install

```bash
go build -o bin/oura ./cmd/oura
```

## Quick start

```bash
export OURA_CLIENT_ID=your_client_id
export OURA_CLIENT_SECRET=your_client_secret
oura auth login --scopes daily heartrate
oura list sleep --start-date 2024-01-01 --end-date 2024-01-07
oura get daily_activity <document_id>
oura whoami
```

## OAuth2 login

Defaults:
- Redirect URI: `http://127.0.0.1:8797/callback`
- Scopes: `daily`

Manual flow:

```bash
oura auth login --paste --no-open
```

## Config

Default config path: `~/.config/oura/config.json`
Stored with mode `0600` and includes tokens and client secret.

Environment overrides:

- `OURA_CLIENT_ID`
- `OURA_CLIENT_SECRET`
- `OURA_REDIRECT_URI`
- `OURA_SCOPES`
- `OURA_ACCESS_TOKEN`
- `OURA_REFRESH_TOKEN`

## Commands

```text
oura auth login|status|logout
oura list <resource> [filters]
oura get <resource> [document_id]
oura resources
oura whoami
```

## Versioning

Use `-ldflags "-X github.com/mattjefferson/oura-cli/internal/app.version=..."` when building.
