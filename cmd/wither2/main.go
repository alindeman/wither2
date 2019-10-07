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

	ingest                  = app.Command("ingest", "Ingest Minecraft logs, forwarding relevant messages to Slack")
	slackWebhookURL         = ingest.Flag("slack-webhook-url", "Slack webhook URL").Envar("SLACK_WEBHOOK_URL").Required().String()
	slackWebhookTimeout     = ingest.Flag("slack-webhook-timeout", "Timeout for each Slack webhook request").Envar("SLACK_WEBHOOK_TIMEOUT").Default("5s").Duration()
	discardDurationInterval = ingest.Flag("discard-duration-interval", "Log messages will be discarded if they are more than this duration in the past or future").Envar("DISCARD_DURATION_INTERVAL").Default("1m").Duration()
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

		now := time.Now()
		diff := now.Sub(msg.Timestamp)
		if diff < -*discardDurationInterval || diff > *discardDurationInterval {
			fmt.Fprintf(os.Stderr, "skipping message too far into past/future: %v\n", scanner.Text())
			continue
		}

		if strings.HasPrefix(msg.Message, "<") || strings.HasSuffix(msg.Message, "joined the game") || strings.HasSuffix(msg.Message, "left the game") {
			cctx, cancel := context.WithTimeout(ctx, *slackWebhookTimeout)

			if err := slackClient.Post(cctx, msg.Message); err != nil {
				fmt.Fprintf(os.Stderr, "error posting to Slack: %v\n", err)
			}

			cancel()
		}
	}

	return scanner.Err()
}
