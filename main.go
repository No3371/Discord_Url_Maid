package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/joho/godotenv"
)

// cleanTrackingParams removes tracking parameters from any URLs in the message
// CleanMessageAndReport function that processes a message string and cleans up URLs based on providers' rules
func CleanMessageAndReport(message string, data *Data) string {
	stats.TotalMessages++
	// Find all URLs in the message
	urlMatch, err := urlExtractor.FindStringMatch(message)
	if err != nil {
		log.Println("Failed to find URLs in message:", err)
		return ""
	}

	containsRedirect := false
	modified := false
	reply := ""

	// Loop through all matches (URLs)
	for urlMatch != nil {
		stats.TotalURLs++
		urlStr := urlMatch.String()
		urlModified := false
		urlMatched := false

		// Loop through each provider
		for name, provider := range data.Providers {
			if urlMatched && !strings.HasPrefix(name, "globalRules") {
				continue
			}

			if match, _ := provider.UrlPattern.MatchString(urlStr); !match {
				continue // next provider
			}

			if !strings.HasPrefix(name, "globalRules") {
				urlMatched = true
			}

			for _, rdr := range provider.Redirections {
				if ridrectFound, _ := rdr.MatchString(urlStr); ridrectFound {
					stats.Redirects++
					containsRedirect = true
					continue
				}
			}

			// Skip URL if it matches any exception pattern
			exceptionFound := false
			for _, exception := range provider.Exceptions {
				if exceptionMatch, _ := exception.MatchString(urlStr); exceptionMatch {
					exceptionFound = true
					break // next exception rule
				}
			}
			if exceptionFound {
				continue // next provider
			}

			paramMatch, err := paramExtracter.FindStringMatch(urlStr)
			if err != nil {
				log.Println("Failed to find parameters in URL:", err)
				continue
			}

			// Loop through all query parameters
			for paramMatch != nil {
				stats.TotalParams++
				// Extract the param key and value
				paramName := paramMatch.GroupByNumber(1).String()

				// Check if the paramValue matches any of the provider's rules
				for _, rule := range provider.Rules {
					if match, _ := rule.MatchString(paramName); match {
						// Remove or replace based on the initial character ('?' or '&')
						if strings.HasPrefix(paramMatch.String(), "&") {
							urlStr = strings.Replace(urlStr, paramMatch.String(), "", 1)
						} else if strings.HasPrefix(paramMatch.String(), "?") {
							urlStr = strings.Replace(urlStr, paramMatch.String(), "?", 1)
						}

						stats.CleanedParams++
						modified = true
						urlModified = true
						break
					}
				}

				paramMatch, err = paramExtracter.FindNextMatch(paramMatch)
				if err != nil {
					log.Println("Failed to find next parameter in URL:", err)
				}
			}
		}

		if urlModified {
			stats.CleanedURLs++
			if urlStr[len(urlStr)-1] == '?' {
				urlStr = urlStr[:len(urlStr)-1]
			}
			reply += urlStr + "\n"
		}

		// Move to the next match (URL)
		urlMatch, err = urlExtractor.FindNextMatch(urlMatch)
		if err != nil {
			log.Println("Failed to find next URL in message:", err)
		}
	}

	if containsRedirect {
		reply = reply + "↪️ Redirect Found / 此訊息包含轉址\n"
	}

	if modified {
		stats.CleanedMessages++
	}

	if len(reply) > 0 && reply[len(reply)-1] == '\n' {
		reply = reply[:len(reply)-1]
	}
	// Return the cleaned-up message
	return reply
}

var stats *Stats = &Stats{}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := contextWithSigterm(context.Background())

	b, err := FetchAndLoadJSON(repo)
	if err != nil {
		log.Fatal(err)
	}

	go StatsWorker(ctx, stats)

	s := state.NewWithIntents("Bot "+os.Getenv("BOT_TOKEN"), gateway.IntentGuildMessages+gateway.IntentMessageContent)
	s.AddHandler(
		// MessageCreate is called every time a message is sent in a server the bot has access to
		func(m *gateway.MessageCreateEvent) {
			// Ignore bot messages
			if m.Author.Bot {
				return
			}

			if !strings.Contains(m.Content, "http") {
				return
			}

			// Check if the message contains URLs
			reply := CleanMessageAndReport(m.Content, b)

			if reply != "" {
				edit := api.EditMessageData{}
				edit.Flags = new(discord.MessageFlags)
				*edit.Flags = m.Flags
				*edit.Flags |= discord.SuppressEmbeds
				_, err := s.EditMessageComplex(m.ChannelID, m.ID, edit)
				if err != nil {
					log.Printf("Failed to edit message: %v", err)
				}

				_, err = s.SendMessageComplex(m.ChannelID, api.SendMessageData{
					Content:         locale(m.Author.Locale, "reply") + reply,
					AllowedMentions: nil,
					Reference: &discord.MessageReference{
						MessageID: m.ID,
						ChannelID: m.ChannelID,
						GuildID:   m.GuildID,
					},
					Flags: discord.SuppressNotifications,
				})
				if err != nil {
					log.Printf("Failed to reply: %v", err)
				}
			}
		})

	err = s.Connect(ctx)
	if err != nil {
		log.Printf("Failed to open session: %v", err)
	}
	defer s.Close()

	// Wait for Ctrl+C or another termination signal to stop the bot
	log.Println("Bot is running...")
}

// https://gist.github.com/matejb/87064825093c42c1e76e7175665d9a9b
func contextWithSigterm(ctx context.Context) context.Context {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

		select {
		case <-signalCh:
		case <-ctx.Done():
		}
	}()

	return ctxWithCancel
}

type Stats struct {
	CleanedMessages int
	TotalMessages   int
	CleanedURLs     int
	TotalURLs       int
	CleanedParams   int
	TotalParams     int
	Redirects       int
}

func StatsWorker(ctx context.Context, stats *Stats) {
	LoadStats(stats)
	t := time.NewTimer(time.Minute * 5)
	for {
		select {
		case <-ctx.Done():
			SaveStats(stats)
			return
		case <-t.C:
			SaveStats(stats)
			t.Reset(time.Minute * 5)
			continue
		}
	}
}

const STATS_FILE = "stats.json"

func LoadStats(stats *Stats) {

	backup := func() {
		data, err := os.ReadFile(STATS_FILE)
		if err != nil {
			fmt.Errorf("failed to read stats and failed to copy (read): %w", err)
			*stats = Stats{}
			return
		}
		// Write data to dst
		err = os.WriteFile(STATS_FILE, data, 0644)
		if err != nil {
			fmt.Errorf("failed to read stats and failed to copy (write): %w", err)
			*stats = Stats{}
			return
		}
	}

	b, err := os.ReadFile(STATS_FILE)
	if err == nil {
		err = json.Unmarshal(b, stats)
		if err != nil {
			fmt.Errorf("failed to unmarshal stats: %w", err)
			backup()
			*stats = Stats{}
			return
		}
	} else if err != nil && os.IsNotExist(err) {
		log.Println(err)
		*stats = Stats{}
	} else {
		backup()
		*stats = Stats{}
	}
}
func SaveStats(stats *Stats) {
	b, err := json.Marshal(stats)
	if err != nil {
		log.Println(err)
	}

	f, err := os.Create(STATS_FILE)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		log.Println(err)
	}
}

func locale(lang string, id string) string {
	switch id {
	case "reply":
		switch lang {
		case "zh-CN":
			fallthrough
		case "zh-TW":
			return "把主人的 URL 🧹掃乾淨✨✨\n"
		case "ja":
			return "御主人様の URL を🧹綺麗にしました✨✨\n"
		default:
			return "I made🧹 Master's URL Clean✨✨\n"
		}
	default:
		return ""
	}
}
