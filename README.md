# External Initiator

Initiate Chainlink job runs from external sources.

## Installation

`go install`

## Configuration

### Environment variables

| Key                        | Description                                                                                | Example                                                            |
| -------------------------- | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `EI_DATABASEURL`           | Postgres connection URL                                                                    | `postgresql://user:pass@localhost:5432/ei`                         |
| `EI_CHAINLINKURL`          | The URL of the Chainlink Core service                                                      | `http://localhost:6688`                                            |
| `EI_IC_ACCESSKEY`          | The Chainlink access key, used for traffic flowing from this service to Chainlink          | `0b7d4a293bff4baf8de852bfa1f1f78a`                                 |
| `EI_IC_SECRET`             | The Chainlink secret, used for traffic flowing from this service to Chainlink              | `h23MjHx17UJKBf3b0MWNI2P/UPh3c3O7/j8ivKCBhvcWH3H+xso4Gehny/lgpAht` |
| `EI_CI_ACCESSKEY`          | The External Initiator access key, used for traffic flowing from Chainlink to this service | `0b7d4a293bff4baf8de852bfa1f1f78a`                                 |
| `EI_CI_SECRET`             | The External Initiator secret, used for traffic flowing from Chainlink to this service     | `h23MjHx17UJKBf3b0MWNI2P/UPh3c3O7/j8ivKCBhvcWH3H+xso4Gehny/lgpAht` |
| `EI_KEEPER_BLOCK_COOLDOWN` | Number of blocks to cool down before triggering a new run for a Keeper job.                | `3`                                                                |

## Usage

```
$ ./external-initiator --help
Monitors external blockchains and relays events to Chainlink node. Supplying endpoint configs as args will delete all other stored configs. ENV variables can be set by prefixing flag with EI_: EI_ACCESSKEY

Usage:
  external-initiator [endpoint configs] [flags]

Flags:
      --chainlinkurl string         The URL of the Chainlink Core Service (default "localhost:6688")
      --ci_accesskey string         The External Initiator access key, used for traffic flowing from Chainlink to this Service
      --ci_secret string            The External Initiator secret, used for traffic flowing from Chainlink to this Service
      --cl_retry_attempts uint      The maximum number of attempts that will be made for job run triggers (default 3)
      --cl_retry_delay duration     The delay between attempts for job run triggers (default 1s)
      --cl_timeout duration         The timeout for job run triggers to the Chainlink node (default 5s)
      --databaseurl string          DatabaseURL configures the URL for external initiator to connect to. This must be a properly formatted URL, with a valid scheme (postgres://). (default "postgresql://postgres:password@localhost:5432/ei?sslmode=disable")
  -h, --help                        help for external-initiator
      --ic_accesskey string         The Chainlink access key, used for traffic flowing from this Service to Chainlink
      --ic_secret string            The Chainlink secret, used for traffic flowing from this Service to Chainlink
      --keeper_block_cooldown int   Number of blocks to cool down before triggering a new run for a Keeper job (default 3)
      --mock                        Set to true if the External Initiator should expect mock events from the blockchains
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

## Adding to Chainlink

In order to use external initiators in your Chainlink node, first enable the following config in your Chainlink node's environment:

```
FEATURE_EXTERNAL_INITIATORS=true
```

This unlocks the ability to run `chainlink initiators` commands. To add an initiator run:

```bash
chainlink initiators create NAME URL
```

Where NAME is the name that you can to assign the initiator (ex: chain-init), and URL is the URL of the `/jobs` endpoint of this external-initiator service (ex: http://localhost:8080/jobs).

Once created, the output will provide the authentication necessary to add to the [environment](#environment-variables) of this external-initiator service.

Once the initiator is created, you will be able to add jobs to your Chainlink node with the type of external, and the name in the param with the name that you assigned the initiator.

## Integration testing

The External Initiator has an integrated mock blockchain client that can be used to test blockchain implementations.

### Setup

Build Docker images and install dependencies by running the setup script:

```bash
./integration/setup
```

### Usage

Simply run the automated integration tests script:

```bash
./integration/run_test
```

### Stopping

All containers should automatically be stopped. If the integration tests script exits early, containers may not have
been stopped. You can manually stop them by running the stop script:

```bash
./integration/stop_docker
```
