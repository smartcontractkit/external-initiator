# External Initiator

Initiate Chainlink job runs from external sources.

## Installation

`go install`

## Configuration

### Environment variables

| Key | Description | Example |
|-----|-------------|---------|
| `DATABASE_URL` | Postgres connection URL | `postgresql://user:pass@localhost:5432/ei` |
| `CL_URL` | URL for the CL node | `http://localhost:6688` |
| `CL_ACCESS_KEY` | CL node access key | `0b7d4a293bff4baf8de852bfa1f1f78a` |
| `CL_ACCESS_SECRET` | CL node access secret | `h23MjHx17UJKBf3b0MWNI2P/UPh3c3O7/j8ivKCBhvcWH3H+xso4Gehny/lgpAht` |

## Usage

1. Set environment variables (supports `.env` file)
2. Run the external initiator (takes no arguments):

```bash
external-initiator
```
