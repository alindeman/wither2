package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alindeman/wither2/internal/minecraft"
	"github.com/alindeman/wither2/internal/slack"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("wither2", "Minecraft <-> Slack bridge")

	ingest          = app.Command("ingest", "Ingest Minecraft logs, forwarding relevant messages to Slack")
	slackWebhookURL = ingest.Flag("slack-webhook-url", "Slack webhook URL").Envar("SLACK_WEBHOOK_URL").Required().String()
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case ingest.FullCommand():
		return runIngest(ctx)
	default:
		return errors.New("unknown subcommand")
	}
}

func runIngest(ctx context.Context) error {
	slackClient := &slack.Client{
		WebhookURL: *slackWebhookURL,
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg, err := minecraft.ParseLogMessage(scanner.Text())
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipping unparseable message: %v\n", scanner.Text())
			continue
		}

		if strings.HasPrefix(msg.Message, "<") || strings.HasSuffix(msg.Message, "joined the game") || strings.HasSuffix(msg.Message, "left the game") {
			cctx, cancel := context.WithTimeout(ctx, 5*time.Second)

			if err := slackClient.Post(cctx, msg.Message); err != nil {
				fmt.Fprintf(os.Stderr, "error posting to Slack: %v\n", err)
			}

			cancel()
		}
	}

	return scanner.Err()
}
