# External Initiator

Initiate Chainlink job runs from external sources.

## Installation

`go install`

## Configuration

### Environment variables

| Key | Description | Example |
|-----|-------------|---------|
| `EI_DATABASEURL` | Postgres connection URL | `postgresql://user:pass@localhost:5432/ei` |
| `EI_CHAINLINK` | URL for the CL node | `http://localhost:6688` |
| `EI_CLACCESSKEY` | CL node access key | `0b7d4a293bff4baf8de852bfa1f1f78a` |
| `EI_CLSECRET` | CL node access secret | `h23MjHx17UJKBf3b0MWNI2P/UPh3c3O7/j8ivKCBhvcWH3H+xso4Gehny/lgpAht` |

## Usage

1. Set environment variables (supports `.env` file)
2. Run the external initiator:

```bash
$ ./external-initiator --help
Monitors external blockchains and relays events to Chainlink node. ENV variables can be set by prefixing flag with EI_: EI_ACCESSKEY

Usage:
  external-initiator [endpoint configs] [flags]

Flags:
      --chainlink string     The URL of the Chainlink Core service (default "localhost:6688")
      --claccesskey string   The access key to identity the node to Chainlink
      --clsecret string      The secret to authenticate the node to Chainlink
      --databaseurl string   DatabaseURL configures the URL for chainlink to connect to. This must be a properly formatted URL, with a valid scheme (postgres://). (default "postgresql://postgres:password@localhost:5432/ei?sslmode=disable")
  -h, --help                 help for external-initiator
```
