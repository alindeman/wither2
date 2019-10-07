# wither2

Inspired by [wither](https://github.com/pickaxe-club/wither), **wither2** is a similar implementation in Golang.

Currently **wither2** only supports ingesting Minecraft logs and sending messages _to_ Slack, but in future it may support sending Slack messages to Minecraft.

## Build

```bash
go install -v ./cmd/wither2
```

Or the Docker image:

```bash
docker build .
```

## Running

### Ingest

The `ingest` subcommand expects the Minecraft server log tailed into it. It will then forward relevant messages to Slack.

```bash
export SLACK_WEBHOOK_URL=...

tail -F /path/to/minecraft/logs/latest.log | wither2 ingest
```
