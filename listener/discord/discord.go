package discord

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/riking/ArchiveBot"
)

var (
	rgxHasLink = regexp.MustCompile(`https?:[^ ]+`)
	// goldmark
	rgxHasLink2 = regexp.MustCompile(`^(?:http|https|ftp):\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]+(?:(?:/|[#?])[-a-zA-Z0-9@:%_+.~#$!?&//=\(\);,'">\^{}\[\]` + "`" + `]*)?`)
)

type Listener struct {
	Config

	s  *discordgo.Session
	db *sql.DB

	uploader *struct{}
}

type Config struct {
	ClientID     string
	ClientSecret string
	BotToken     string

	// If Shards is 0, sharding is disabled.
	Shards  int
	MyShard int
}

func NewListener(config json.RawMessage, db *sql.DB, uploader *struct{}) (*Listener, error) {
	var c Config
	err := json.Unmarshal([]byte(config), &c)
	if err != nil {
		return nil, err
	}
	l := &Listener{
		Config:   c,
		db:       db,
		uploader: uploader,
	}

	if l.Config.ClientID == "" || l.Config.ClientSecret == "" || l.Config.BotToken == "" {
		return nil, fmt.Errorf("missing required configuration")
	}
	if l.Config.Shards != 0 && (l.Config.MyShard < 0 || l.Config.MyShard >= l.Config.Shards) {
		return nil, fmt.Errorf("invalid shard number %d (max %d)", l.Config.MyShard, l.Config.Shards)
	}

	return l, nil
}

func (l *Listener) Start() error {
	s, err := discordgo.New("Bot " + l.Config.BotToken)
	if err != nil {
		return err
	}
	l.s = s
	state := discordgo.NewState()
	state.TrackChannels = true
	state.TrackEmojis = false
	state.TrackMembers = false
	state.TrackRoles = false
	state.TrackVoice = false
	state.TrackPresences = false
	state.MaxMessageCount = 0
	s.State = state

	// s.Identify.LargeThreshold = 50
	// s.Identify.GuildSubscriptions = false
	// s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds|discordgo.IntentsGuildMessages|discordgo.IntentsGuildMessageReactions|discordgo.IntentsDirectMessages|discordgo.IntentsDirectMessageReactions)

	s.UserAgent = archivebot.TrueUserAgent
	s.Client = &http.Client{
		Timeout:   20 * time.Second,
		Transport: nil,
	}

	gb, err := s.GatewayBot()
	if err != nil {
		return err
	}

	if !(l.Config.Shards == 0 && gb.Shards == 1) && (l.Config.Shards*2 < gb.Shards) {
		return fmt.Errorf("Discord shard count must increase: have %d, want %d", l.Config.Shards, gb.Shards)
	}

	if l.Config.Shards != 0 {
		s.ShardID = l.Config.MyShard
		s.ShardCount = l.Config.Shards
	}

	s.AddHandler(l.OnMessage)
	s.AddHandler(l.OnMessageUpdate)
	s.AddHandler(l.OnGuildRemove)
	s.AddHandler(l.OnResume)

	err = s.Open()
	if err != nil {
		return fmt.Errorf("open socket: %w", err)
	}
	return nil
}

func (l *Listener) Channel(channelID string) (*discordgo.Channel, error) {
	ch, err := l.s.State.Channel(channelID)
	if err == discordgo.ErrStateNotFound {
		ch, err = l.s.Channel(channelID)
		if err != nil {
			return nil, err
		}
		l.s.State.ChannelAdd(ch)
	}
	return ch, nil
}

func (l *Listener) OnResume(_ *discordgo.Session, m *discordgo.Ready) {
	fmt.Println("discord: connected")
}

func (l *Listener) OnGuildRemove(_ *discordgo.Session, ev *discordgo.GuildDelete) {
	// unimplemented!

	// Wipe configuration records from database
}

func (l *Listener) OnMessage(_ *discordgo.Session, ev *discordgo.MessageCreate) {
	m := ev.Message
	if link := rgxHasLink.FindString(m.Content); link != "" {
		link2 := rgxHasLink2.FindString(m.Content)
		fmt.Printf("[DEBUG] found link\n  link1: %s\n  link2: %s\nmessage: %s\n", link, link2, m.Content)
	}
	if m.Embeds != nil {
		json.NewEncoder(os.Stdout).Encode(m.Embeds)
	}
}

func (l *Listener) OnMessageUpdate(_ *discordgo.Session, ev *discordgo.MessageUpdate) {
	if ev.Message.Embeds != nil {
		json.NewEncoder(os.Stdout).Encode(ev.Message.Embeds)
	}
}
