package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dutil/commandsystem"
)

var Commands = []commandsystem.CommandHandler{
	&commandsystem.SimpleCommand{
		Name:        "Help",
		Description: "Shows help abut all or one specific command",
		Arguments: []*commandsystem.ArgumentDef{
			&commandsystem.ArgumentDef{Name: "command", Description: "Optionally specify a command to show help for", Type: commandsystem.ArgumentTypeString},
		},
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			target := ""
			if parsed.Args[0] != nil {
				target = parsed.Args[0].Str()
			}
			help := cmdSystem.GenerateHelp(target, 0)
			dgo.ChannelMessageSend(m.ChannelID, "**Help** - *(For problems/whatever contact jonas747#3124)*\n"+help+"\n"+VERSION)
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Invite",
		Description: "Responds with bot invite link",
		RunInDm:     true,
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			dgo.ChannelMessageSend(m.ChannelID, "You smell bad https://discordapp.com/oauth2/authorize?client_id=201163424485343232&scope=bot&permissions=101376")
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Stats",
		Aliases:     []string{"st"},
		Description: "Shows stats for a user or yourself",
		Arguments: []*commandsystem.ArgumentDef{
			&commandsystem.ArgumentDef{Name: "User", Description: "Optionally specify a user", Type: commandsystem.ArgumentTypeUser},
		},
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			user := m.Author
			if parsed.Args[0] != nil {
				user = parsed.Args[0].DiscordUser()
			}

			player := playerManager.GetCreatePlayer(user.ID, user.Username)

			player.Lock()
			stats := fmt.Sprintf("Stats for **%s**\n - Money: **$%d**", user.Username, player.Money)
			player.Unlock()

			go SurelySend(m.ChannelID, stats)
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "FreeMoney",
		Aliases:     []string{"fm", "giefmoney", "gief", "mmm"},
		Description: "Gives you $50 if you have less than that",
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			player := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)

			player.Lock()

			if player.Money < 50 {
				player.Money += 50
				stats := fmt.Sprintf("Stats for **%s**\n - Money: **$%d**", m.Author.Username, player.Money)
				go SurelySend(m.ChannelID, stats)
			} else {
				go SurelySend(m.ChannelID, "You have too much money already >:{")
			}

			player.Unlock()

			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Create",
		Aliases:     []string{"c"},
		Description: "Creates a table",
		Arguments: []*commandsystem.ArgumentDef{
			&commandsystem.ArgumentDef{Name: "Buy in", Description: "Your buy in amount", Type: commandsystem.ArgumentTypeNumber},
			&commandsystem.ArgumentDef{Name: "Stakes-small", Description: "Small stakes for this table", Type: commandsystem.ArgumentTypeNumber},
			&commandsystem.ArgumentDef{Name: "Stakes-min", Description: "Big stakes for this table", Type: commandsystem.ArgumentTypeNumber},
		},
		RequiredArgs: 3,
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			privateChannel, err := GetCreatePrivateChannel(m.Author.ID)
			if err != nil {
				return err
			}

			buyin := parsed.Args[0].Int()
			small := parsed.Args[1].Int()
			big := parsed.Args[2].Int()

			evt := &CreateTableEvt{
				PlayerID:       m.Author.ID,
				PrivateChannel: privateChannel,
				Name:           m.Author.Username,
				Channel:        m.ChannelID,
				BuyIn:          buyin,
				Small:          small,
				Big:            big,
			}

			tableManager.EvtChan <- evt
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Join",
		Aliases:     []string{"j"},
		Description: "Joins a table",
		Arguments: []*commandsystem.ArgumentDef{
			&commandsystem.ArgumentDef{Name: "Buy in", Description: "Buy in amount, has to be larger than 50*min-bet", Type: commandsystem.ArgumentTypeNumber},
		},
		RequiredArgs: 1,
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			privateChannel, err := GetCreatePrivateChannel(m.Author.ID)
			if err != nil {
				return err
			}

			money := parsed.Args[0].Int()

			evt := &AddPlayerEvt{
				PlayerID:       m.Author.ID,
				PrivateChannel: privateChannel,
				Name:           m.Author.Username,
				Channel:        m.ChannelID,
				BuyIn:          money,
			}

			tableManager.EvtChan <- evt
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Start",
		Aliases:     []string{"s"},
		Description: "Starts a table",
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			evt := &StartEvt{
				Channel: m.ChannelID,
			}

			tableManager.EvtChan <- evt
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Stop",
		Description: "Stops a table, maybe to take a break or something",
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			tableManager.EvtChan <- &StopTableEvt{PlayerID: m.Author.ID, Channel: m.ChannelID}
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Kick",
		Description: "Kicks a player from your table",
		Arguments: []*commandsystem.ArgumentDef{
			&commandsystem.ArgumentDef{Name: "Target", Description: "Player to kick", Type: commandsystem.ArgumentTypeUser},
		},
		RequiredArgs: 1,
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			target := parsed.Args[0].DiscordUser()
			tableManager.EvtChan <- &KickPlayerEvt{PlayerID: m.Author.ID, KickPlayerID: target.ID, Channel: m.ChannelID}
			return nil
		},
	},
	&commandsystem.SimpleCommand{
		Name:         "Ban",
		Description:  "Bans a player from your table >:O",
		RequiredArgs: 1,
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			target := parsed.Args[0].DiscordUser()
			tableManager.EvtChan <- &BanPlayerEvt{PlayerID: m.Author.ID, BanPlayerID: target.ID, Channel: m.ChannelID}
			return nil
		},
	},
	&commandsystem.CommandContainer{
		Name:        "Config",
		Aliases:     []string{"conf"},
		Description: "Changes a table configuration",
		Children: []commandsystem.CommandHandler{
			&commandsystem.SimpleCommand{
				Name:        "Get",
				Aliases:     []string{"g", "print", "show"},
				Description: "Shows the current config",
				RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
					tableManager.EvtChan <- &PrintInfoEvt{Channel: m.ChannelID}
					return nil
				},
			},
			&commandsystem.SimpleCommand{
				Name:        "Set",
				Aliases:     []string{"s"},
				Description: "Chages a config setting",
				Arguments: []*commandsystem.ArgumentDef{
					&commandsystem.ArgumentDef{Name: "Key", Description: "What to change", Type: commandsystem.ArgumentTypeString},
					&commandsystem.ArgumentDef{Name: "Value", Description: "The new value", Type: commandsystem.ArgumentTypeString},
				},
				RequiredArgs: 2,
				RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
					tableManager.EvtChan <- &ChangeSettingsEvt{Channel: m.ChannelID, PlayerID: m.Author.ID, Settings: map[string]string{parsed.Args[0].Str(): parsed.Args[1].Str()}}
					return nil
				},
			},
		},
	},
	&commandsystem.SimpleCommand{
		Name:        "Leave",
		Description: "Leaves a table",
		RunFunc: func(parsed *commandsystem.ParsedCommand, m *discordgo.MessageCreate) error {
			evt := &RemovePlayerEvt{
				PlayerID: m.Author.ID,
				Channel:  m.ChannelID,
			}

			tableManager.EvtChan <- evt
			return nil
		},
	},
}

func GetCreatePrivateChannel(userID string) (string, error) {
	dgo.State.RLock()
	for _, channel := range dgo.State.PrivateChannels {
		if channel.Recipient.ID == userID {
			dgo.State.RUnlock()
			return channel.ID, nil
		}
	}
	dgo.State.RUnlock()

	channel, err := dgo.UserChannelCreate(userID)
	if err != nil {
		return "", err
	}

	go dgo.State.ChannelAdd(channel)
	return channel.ID, nil
}
