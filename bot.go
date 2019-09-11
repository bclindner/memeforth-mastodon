package main

import (
	"encoding/json"
	"github.com/mattn/go-mastodon"
	"html"
	"github.com/microcosm-cc/bluemonday"
	"io/ioutil"
	"log"
	"context"
	"regexp"
	"strings"
)

const maxChars = 500

// tag stripper
var (
	statusRegexp = regexp.MustCompile(" ([^@]+)$")
	striptags = bluemonday.StrictPolicy()
)

type Config struct {
	InstanceURL string `json:"instanceURL"`
	AccessToken string `json:"accessToken"`
}

func GetClient(cfg Config) *mastodon.Client {
	return mastodon.NewClient(&mastodon.Config{
		Server: cfg.InstanceURL,
		AccessToken: cfg.AccessToken,
	})
}

func main() {
	// read config from file
	cfgFile, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalf("Failed to read config file at config.json. Is it unavailable?\nFull error: %v\n", err)
	}
	var cfg Config
	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config file!\nFull error: %v\n", err)
	}
	// try to open Mastodon connection with this file
	client := mastodon.NewClient(&mastodon.Config{
		Server: cfg.InstanceURL,
		AccessToken: cfg.AccessToken,
	})
	ctx, ctxCancel := context.WithCancel(context.Background())
	// test connection by getting self
	self, err := client.GetAccountCurrentUser(ctx)
	if err != nil {
		log.Fatalf("Failed to get account information: %v\n", err)
	}
	log.Printf("Logged in as %v\n", self.Username)
	// open websocket connection
	wsclient := client.NewWSClient()
	evtStream, err := wsclient.StreamingWSUser(ctx)
	if err != nil {
		log.Fatalf("Failed to establish WebSocket connection!\nFull error: %v\n", err)
	}
	// enter event loop
	log.Println("Entering event loop.")
	for genericEvent := range evtStream {
		switch evt := genericEvent.(type) {
			// handle mention notifications
			case *mastodon.NotificationEvent:
				switch evt.Notification.Type {
				case "mention":
					status := evt.Notification.Status
					content := html.UnescapeString(striptags.Sanitize(status.Content))
					matches := statusRegexp.FindStringSubmatch(content)
					forthcode := strings.TrimSpace(matches[1])
					result, err := ProcessMemeForth(forthcode)
					if err != nil {
						log.Printf("Error processing code: %v\nError: %v\n", forthcode, err)
					}
					if len(result) < maxChars {
						_, err = client.PostStatus(ctx, &mastodon.Toot{
							InReplyToID: status.ID,
							Status: "@"+status.Account.Acct + " " + result,
							Visibility: status.Visibility,
						})
						if err != nil {
							log.Printf("Failed to send toot: %v", err)
						}
					}
				default:
					break
				}
			// handle errors
			case *mastodon.ErrorEvent:
				log.Printf("Error in WebSocket event loop: %v\n", evt.Error())
				ctxCancel()
				break
			// ignore any other events
			default:
				continue
		}
	}
}
