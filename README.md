# External Initiator

Initiate Chainlink job runs from external sources.

## Installation

`go install`

## Usage

```bash
external-initiator [flags]
```

Flags:

- `-e`/`--endpoint`: Provide endpoint to use in `blockchain:URL` format.
    Eg: `--endpoint eth:wss://127.0.0.1:8546`. Flag is only required on the first run. Providing the flag
    will overwrite any previous endpoints for the same blockchain.
