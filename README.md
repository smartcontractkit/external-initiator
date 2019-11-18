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

```bash
$ ./external-initiator --help
Monitors external blockchains and relays events to Chainlink node. Supplying endpoint configs as args will delete all other stored configs. ENV variables can be set by prefixing flag with EI_: EI_ACCESSKEY

Usage:
  external-initiator [endpoint configs] [flags]

Flags:
      --chainlink string     The URL of the Chainlink Core service (default "localhost:6688")
      --claccesskey string   The access key to identity the node to Chainlink
      --clsecret string      The secret to authenticate the node to Chainlink
      --databaseurl string   DatabaseURL configures the URL for chainlink to connect to. This must be a properly formatted URL, with a valid scheme (postgres://). (default "postgresql://postgres:password@localhost:5432/ei?sslmode=disable")
  -h, --help                 help for external-initiator
```

### Supply Endpoint configs via HTTP

You can send a POST request with an Endpoint config to `/configs`.
These configs will be stored in the database, and be available when restarting the EI if no configs are passed as args.
Endpoint names are unique identifiers, and any previous record with the same name will be overwritten.

### Supply Endpoint configs as args

**WARNING:** Supplying Endpoint configs as args will permanently delete any previously stored Endpoint configs.

When running the External Initiator with Endpoint configs passed as args, the EI will delete any other configs and run only using the configs provided.
The configs are stored the same was as with HTTP, and the configs will persist if the EI is restarted without any configs passed as args.

### Example

```bash
$ ./external-initiator "{\"name\":\"eth-mainnet\",\"type\":\"ethereum\",\"url\":\"ws://localhost:8546/\"}" --chainlink "http://localhost:6688/"
```
