package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/joho/godotenv"
)


var stats *Stats = &Stats{}
var allowedMentions *api.AllowedMentions

func init () {
	allowedMentions = &api.AllowedMentions{
		Parse:       []api.AllowedMentionType{},
		Roles:       []discord.RoleID{},
		Users:       []discord.UserID{},
		RepliedUser: new(bool),
	}
	*allowedMentions.RepliedUser = false
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

    loadGuildLocaleMap()

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
			CleanMessage(m, b, s)
		},
	)

	// Wait for Ctrl+C or another termination signal to stop the bot
	log.Println("Bot is running...")
	err = s.Connect(ctx)
	if err != nil {
		log.Printf("Failed to open session: %v", err)
	}
	defer s.Close()

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

func getGuildLocale(guildID discord.GuildID) string {
    if v, ok := guildLocaleMap[int64(guildID)]; ok {
        return v
    }
    return ""
}

var guildLocaleMap map[int64]string

const GUILD_LOCALE_FILE = "guilds_locale.json"

func loadGuildLocaleMap() {
    b, err := os.ReadFile(GUILD_LOCALE_FILE)
	if err != nil {
        return
    }

    temp := make(map[string]string)
    err = json.Unmarshal(b, temp)
    if err != nil {
        fmt.Errorf("failed to unmarshal guild locale map: %w", err)
        return
    }

    for k, v := range temp {
        id, err := strconv.ParseInt(k, 10, 64)
        if err != nil {
            fmt.Errorf("skipping %s because failed to unmarshal parse guild id: %w", k, err)
            continue
        }
        guildLocaleMap[id] = v
    }

	fmt.Printf("Loaded Guild Locale Map: %+v", guildLocaleMap)
}
