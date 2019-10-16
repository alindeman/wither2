package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	server           = app.Command("server", "Serve an endpoint for Slack outgoing webhooks, forwarding relevant messages to Minecraft")
	addr             = server.Flag("addr", "Address to listen on").Envar("ADDR").Default("127.0.0.1").String()
	port             = server.Flag("port", "Port to listen on").Envar("PORT").Default("8080").Int()
	rconDialTimeout  = server.Flag("rcon-dial-timeout", "RCON dial timeout").Envar("RCON_DIAL_TIMEOUT").Default("10s").Duration()
	rconTimeout      = server.Flag("rcon-timeout", "RCON timeout for login and commands").Envar("RCON_TIMEOUT").Default("10s").Duration()
	rconIP           = server.Flag("rcon-ip", "RCON IP address").Envar("RCON_IP").Default("127.0.0.1").String()
	rconPort         = server.Flag("rnon-port", "RCON port").Envar("RCON_PORT").Default("25575").Int()
	rconPassword     = server.Flag("rcon-password", "RCON password").Envar("RCON_PASSWORD").Default("minecraft").String()
	slackToken       = server.Flag("slack-token", "Slack token for outgoing webhook").Envar("SLACK_TOKEN").Required().String()
	slackIgnoreUsers = server.Flag("slack-ignore-users", "Slack users to ignore").Envar("SLACK_IGNORE_USERS").Default("slackbot").Strings()
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigC
		fmt.Fprintf(os.Stderr, "caught %v, shutting down\n", sig)
		cancel()
	}()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case ingest.FullCommand():
		return runIngest(ctx)
	case server.FullCommand():
		return runServer(ctx)
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

		if minecraft.IsChatMessage(msg.Message) || minecraft.IsJoinLeaveMessage(msg.Message) || minecraft.IsDeathMessage(msg.Message) || minecraft.IsAdvancementMessage(msg.Message) {
			cctx, cancel := context.WithTimeout(ctx, *slackWebhookTimeout)

			if err := slackClient.Post(cctx, msg.Message); err != nil {
				fmt.Fprintf(os.Stderr, "error posting to Slack: %v\n", err)
			}

			cancel()
		}
	}

	return scanner.Err()
}

func runServer(ctx context.Context) error {
	conn, err := dialRCON(ctx)
	if err != nil {
		return err
	}

	minecraftClient := minecraft.New(conn)
	if err := minecraftClient.Login(*rconTimeout, *rconPassword); err != nil {
		return fmt.Errorf("failed to login to RCON: %w", err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *addr, *port),
		Handler: rconSlackForwarderHandler(minecraftClient, *slackToken, *slackIgnoreUsers),
	}

	errC := make(chan error)
	go func() {
		errC <- server.ListenAndServe()
	}()

	select {
	case err := <-errC:
		return err
	case <-ctx.Done():
		return server.Shutdown(ctx)
	}
}

func dialRCON(ctx context.Context) (net.Conn, error) {
	dctx, cancel := context.WithTimeout(ctx, *rconDialTimeout)
	defer cancel()

	var d net.Dialer
	return d.DialContext(dctx, "tcp", fmt.Sprintf("%s:%d", *rconIP, *rconPort))
}

func rconSlackForwarderHandler(client *minecraft.Client, slackToken string, slackIgnoreUsers []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		payload := slack.ParseWebhookPayload(r)

		if payload.Token != slackToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		for _, ignoreUser := range slackIgnoreUsers {
			if ignoreUser == payload.UserName {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		text := fmt.Sprintf("<%s> %s", payload.UserName, payload.Text)
		msg, err := json.Marshal(map[string]interface{}{
			"text": text,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := client.Command(*rconTimeout, "tellraw @a "+string(msg)); err != nil {
			fmt.Fprintf(os.Stderr, "failed to forward message to Minecraft: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
