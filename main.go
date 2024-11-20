package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/joho/godotenv"
)

var stats *Stats = &Stats{}
var mentionNone *api.AllowedMentions
var lastDeleteRequest map[discord.MessageID]time.Time

func init() {
	mentionNone = &api.AllowedMentions{
		Parse:       []api.AllowedMentionType{},
		Roles:       []discord.RoleID{},
		Users:       []discord.UserID{},
		RepliedUser: new(bool),
	}
	*mentionNone.RepliedUser = false
	lastDeleteRequest = make(map[discord.MessageID]time.Time)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	loadGuildLocaleMap()

	ctx := contextWithSigterm(context.Background())

	b, err := FetchAndLoadRules(repo)
	if err != nil {
		log.Fatal(err)
	}

	go StatsWorker(ctx, stats)

	s := state.NewWithIntents("Bot "+os.Getenv("BOT_TOKEN"), gateway.IntentGuildMessages+gateway.IntentMessageContent)
	s.AddHandler(
		// MessageCreate is called every time a message is sent in a server the bot has access to
		func(m *gateway.MessageCreateEvent) {
			defer func() {
				err := recover()
				if err != nil {
					log.Printf("Error when handling message: %v", err)
				}
			}()
			TryCleanMessage(m, b, s)
		},
	)

	s.AddHandler(func(m *gateway.ReadyEvent) {
		s.BulkOverwriteCommands(s.Ready().Application.ID, []api.CreateCommandData{
			{
				Name: "❌",
				Type: discord.MessageCommand,
			},
		})
	})

	s.AddHandler(func(m *gateway.InteractionCreateEvent) {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("Error when handling deletion request: %v", err)
			}
		}()
		data := m.Data.(*discord.CommandInteraction)
		if data == nil {
			return
		}

		switch data.Name {
		case "❌":
			if len(data.Resolved.Messages) == 0 {
				return
			}

			for _, toDel := range data.Resolved.Messages {
				me, err := s.Me()
				if err != nil {
					log.Printf("Failed to get me: %v", err)
					return
				}
				if toDel.Author.ID != me.ID {
					return
				}
				if (toDel.ReferencedMessage == nil && strings.HasPrefix(toDel.Content, m.Member.User.Mention())) ||
					(toDel.ReferencedMessage != nil && toDel.ReferencedMessage.Author.ID == m.Member.User.ID) {
					err := s.DeleteMessage(toDel.ChannelID, toDel.ID, "Requested by the original author")
					if err != nil {
						err = s.DeleteMessage(toDel.ChannelID, toDel.ID, "Requested by the original author")
					}
					if err == nil {
						s.RespondInteraction(m.ID, m.Token, api.InteractionResponse{
							Type: api.MessageInteractionWithSource,
							Data: &api.InteractionResponseData{
								Content: option.NewNullableString("OK ✨٩(ˊωˋ*)و✨"),
								Flags:   discord.EphemeralMessage,
							},
						})
					}
				} else {
					tryDeleteByOthersDeferred(s, m, toDel.ChannelID, toDel.ID)
					return
				}
			}

		}
	})

	// Wait for Ctrl+C or another termination signal to stop the bot
	log.Println("Bot is running...")
	err = s.Connect(ctx)
	if err != nil {
		log.Printf("Failed to open session: %v", err)
	}
	defer s.Close()
}

func tryDeleteByOthersDeferred(s *state.State, ev *gateway.InteractionCreateEvent, cId discord.ChannelID, mId discord.MessageID) {
	defer func() { // Clean up
		maxIt := 10
		for k, t := range lastDeleteRequest {
			if time.Since(t) > time.Second*6 {
				delete(lastDeleteRequest, k)
			}
			maxIt--
			if maxIt <= 0 {
				break
			}
		}
	}()
	waiting := false
	lastRequestedTime, foundRequest := lastDeleteRequest[mId]
	if !foundRequest {
		lastDeleteRequest[mId] = time.Now()

		s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.DeferredMessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString("Waiting for 2p...\\等待 2p...\nYou are not the OP, so you need to find someone and press this together to delete this!\n因為你不是原 PO，需要找人同時按這個才能刪除！"),
				Flags:   discord.EphemeralMessage,
			},
		})

		waiting = true
		<-time.After(time.Second * 3 / 2)
	}

	lastRequestedTime, foundRequest = lastDeleteRequest[mId]
	if !foundRequest { // Already deleted
		if waiting {
			s.EditInteractionResponse(ev.AppID, ev.Token, api.EditInteractionResponseData{
				Content: option.NewNullableString("💥COMBO💥合體技發動💥\n✨٩(ˊωˋ*)و✨"),
			})
		}
		return
	}

	if !waiting && time.Since(lastRequestedTime) < time.Second*2 {
		if s.DeleteMessage(cId, mId, "Requested by others") != nil {
			if s.DeleteMessage(cId, mId, "Requested by others") != nil {
				s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Content: option.NewNullableString("(*´･д･)? It failed... \\ 不知道為什麼失敗了..."),
						Flags:   discord.EphemeralMessage,
					},
				})
				return
			}
		}
		delete(lastDeleteRequest, mId)
		s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString("💥COMBO💥合體技發動💥\n✨٩(ˊωˋ*)و✨"),
				Flags:   discord.EphemeralMessage,
			},
		})
		return
	}

	_, err := s.EditInteractionResponse(ev.AppID, ev.Token, api.EditInteractionResponseData{
		Content: option.NewNullableString("You are not the OP, so you need to find someone and press this together to delete this!\n因為你不是原 PO，需要找人同時按這個才能刪除！"),
	})

	if err != nil {
		log.Printf("Error from tryDeleteByOthersDeferred: %v", err)
	}
}

func tryDeleteByOthers(s *state.State, cId discord.ChannelID, mId discord.MessageID) bool {
	defer func() { // Clean up
		maxIt := 10
		for k, t := range lastDeleteRequest {
			if time.Since(t) > time.Second*60 {
				delete(lastDeleteRequest, k)
			}
			maxIt--
			if maxIt <= 0 {
				break
			}
		}
	}()
	lastRequestedTime, requested := lastDeleteRequest[mId]
	if !requested {
		lastDeleteRequest[mId] = time.Now()
		return false
	}
	if time.Since(lastRequestedTime) < time.Second*2 {
		if s.DeleteMessage(cId, mId, "Requested by the original author") != nil {
			s.DeleteMessage(cId, mId, "Requested by the original author")
		}
		delete(lastDeleteRequest, mId)
		return true
	} else {
		lastDeleteRequest[mId] = time.Now()
		return false
	}

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
