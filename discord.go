package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}

const (
	myRole      = "772142260128710719"
	helpChannel = "772167961771245578"
)

func updateStatus(discord *discordgo.Session, quit <-chan struct{}) {
	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case <-ticker.C:
			e := discord.UpdateListeningStatus("tinju’i toi")
			if e != nil {
				fmt.Fprintf(os.Stderr, "%v", e)
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func main() {
	must(loadJVS())

	token, ok := os.LookupEnv("SAMCU_TOKEN")
	if !ok {
		panic(fmt.Sprintf("Need token in env var SAMCU_TOKEN"))
	}
	discord, err := discordgo.New("Bot " + token)
	must(err)
	discord.ShouldReconnectOnError = true

	must(discord.Open())
	discord.AddHandler(handleMessage)
	quitter := make(chan struct{}, 1)
	updateStatus(discord, quitter)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	quitter <- struct{}{}

	discord.Close()
}

var handlers = map[string]func(func(string), string, []string){
	"lujvo":    jvozba,
	"rafsi":    rafsi,
	"selrafsi": rafsi,
	"valsi":    lookup,
	"sisku":    reverseLookup,
	"katna":    katna,
	"gloss":    gloss,
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if s.State.User.ID == m.Author.ID {
		return
	}
	stripped := strings.TrimPrefix(m.Message.Content, "<@&"+myRole+">")
	if len(m.GuildID) > 0 && stripped == m.Message.Content {
		return
	}

	var respond = func(msg string) {
		fmt.Println("→", msg)
		for i := 0; i < len(msg); i += 1918 {
			_, err := s.ChannelMessageSend(m.Message.ChannelID, msg[i:min(i+1918, len(msg))])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
			}
		}
	}
	var errmsg = func() {
		respond("<#" + helpChannel + ">")
	}

	fields := strings.Fields(stripped)
	if len(fields) == 0 {
		errmsg()
		return
	}
	cmd := strings.TrimSuffix(fields[0], ":")
	fields = fields[1:]

	fmt.Println(cmd, fields)

	handler, ok := handlers[cmd]
	if !ok {
		errmsg()
		return
	}
	handler(respond, cmd, fields)
}
