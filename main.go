package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
  "time"
  "strconv"

	"github.com/bwmarrin/discordgo"
  "github.com/nlopes/slack"
)

var (
  voiceStateMap map[string]string
  q chan string
  slackChannelID string
  slackAPIToken string
  discordServerID string
  discordBotToken string
)

func init() {
  voiceStateMap = map[string]string{}
}

func main() {
  slackChannelID = os.Getenv("SLACK_CHANNEL_ID")
  slackAPIToken = os.Getenv("SLACK_API_TOKEN")
  discordServerID = os.Getenv("DISCORD_SERVER_ID")
  discordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
  sg := slack.New(slackAPIToken)
  q = make(chan string, 100)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
  dg.AddHandler(voiceStateUpdate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
  go deleteOldMessages(sg, q)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
    fmt.Printf("ping by : %v\n", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, "Pong!!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
    fmt.Printf("pong by : %v\n", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, "Ping!!")
	}
}

func voiceStateUpdate(s *discordgo.Session, m *discordgo.VoiceStateUpdate) {
  sg := slack.New(slackAPIToken)
  map_cid, _ := voiceStateMap[m.UserID]
  switch m.ChannelID {
  case map_cid:
    return
  case "":
    user, _ := s.User(m.UserID)
    username := user.Username
    text := fmt.Sprintf("[Discord]%vがボイスチャンネルから退出しました。", username)
    _, timestamp, err := sg.PostMessage(slackChannelID, text, slack.PostMessageParameters{AsUser:true})
    if err != nil {
      fmt.Printf("%v\n", err)
      return
    }
    q <- timestamp
    fmt.Printf("%v Message Successfully Sent: %v\n", timestamp, text)
    voiceStateMap[m.UserID] = ""
  default:
    user, _ := s.User(m.UserID)
    channel, _ := s.Channel(m.ChannelID)
    username := user.Username
    channelname := channel.Name
    text := fmt.Sprintf("[Discord]%vがボイスチャンネル「%v」に入室しました。", username, channelname)
    params := slack.PostMessageParameters{AsUser:true}
    attachment := slack.Attachment{
  		Title: text,
  		TitleLink: "https://discordapp.com/channels/" + discordServerID,
	  }
    params.Attachments = []slack.Attachment{attachment}
    _, timestamp, err := sg.PostMessage(slackChannelID, "", params)
    if err != nil {
      fmt.Printf("%v\n", err)
      return
    }
    q <- timestamp
    fmt.Printf("%v Message Successfully Sent: %v\n", timestamp, text)
    voiceStateMap[m.UserID] = m.ChannelID
  }
  /*tmp += 1
  status := fmt.Sprintf("[%v]UserID:%v\nSessionID:%v\nChannelID:%v\nSuppress:%v\nSelfMute:%v\nSelfDeaf:%v\nMute:%v\nDeaf:%v", tmp, m.UserID, m.SessionID, m.ChannelID, m.Suppress, m.SelfMute, m.SelfDeaf, m.Mute, m.Deaf)
  fmt.Println(status)
  s.ChannelMessageSend(m.ChannelID, status)*/
}

// delete messages 1 hour elapsed
func deleteOldMessages(api *slack.Client, q chan string){
  for {
    time.Sleep(1)
    if len(q) > 0 {
      timestamp := <- q
      f, _ := strconv.ParseFloat(timestamp, 64)
      t := int64(f)
      current := time.Now().Unix()
      fmt.Println(current, t, current - t)
      if (current - t) > 3600 {
        respChannel, respTimestamp, err := api.DeleteMessage(slackChannelID, timestamp)
        if err != nil {
          fmt.Printf("Slack API Error(%v): %v\n", timestamp, err)
        } else {
          fmt.Printf("Old message successfully deleted: %v, %v\n", respChannel, respTimestamp)
        }
      } else {
        q <- timestamp
        continue
      }
    } else {
      continue
    }
  }
}
