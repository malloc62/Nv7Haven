package eod

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const maxComboLength = 20
const guild = "" // 819077688371314718 for testing

var combs = []string{
	"\n",
	"+",
	",",
}

func (b *EoD) cmdHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := b.newMsgNormal(m)
	rsp := b.newRespNormal(m)

	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	if strings.HasPrefix(m.Content, "?") {
		if len(m.Content) < 2 {
			return
		}
		b.infoCmd(strings.TrimSpace(m.Content[1:]), msg, rsp)
		return
	}

	if strings.HasPrefix(m.Content, "*2") {
		if !b.checkServer(msg, rsp) {
			return
		}
		lock.RLock()
		dat, exists := b.dat[msg.GuildID]
		lock.RUnlock()
		if !exists {
			return
		}
		if dat.combCache == nil {
			dat.combCache = make(map[string]comb)
		}
		comb, exists := dat.combCache[msg.Author.ID]
		if !exists {
			return
		}
		if comb.elem3 != "" {
			b.combine([]string{comb.elem3, comb.elem3}, msg, rsp)
			return
		}
		b.combine(comb.elems, msg, rsp)
		return
	}

	for _, comb := range combs {
		if strings.Contains(m.Content, comb) {
			if !b.checkServer(msg, rsp) {
				return
			}
			parts := strings.Split(m.Content, comb)
			if len(parts) < 2 {
				return
			}
			for i, part := range parts {
				parts[i] = strings.TrimSpace(strings.Replace(part, "\\", "", -1))
			}
			if len(parts) > maxComboLength {
				rsp.ErrorMessage(fmt.Sprintf("You can only combine up to %d elements!", maxComboLength))
				return
			}
			b.combine(parts, msg, rsp)
			return
		}
	}
}

func (b *EoD) initHandlers() {
	// Debugging
	var err error
	datafile, err = os.OpenFile("eodlogs.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	b.initInfoChoices()

	cmds, err := b.dg.ApplicationCommands(clientID, guild)
	if err != nil {
		panic(err)
	}
	cms := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range cmds {
		cms[cmd.Name] = cmd
	}
	for _, val := range commands {
		if val.Name == "elemsort" {
			val.Options[0].Choices = infoChoices
		}
		cmd, exists := cms[val.Name]
		if !exists || !commandsAreEqual(cmd, val) {
			_, err := b.dg.ApplicationCommandCreate(clientID, guild, val)
			if err != nil {
				fmt.Printf("Failed to update command %s\n", val.Name)
			} else {
				fmt.Printf("Updated command %s\n", val.Name)
			}
		}
	}

	b.dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		rsp := b.newRespSlash(i)
		if (i.Data.Name != "suggest") && (i.Data.Name != "mark") && (i.Data.Name != "image") && (i.Data.Name != "inv") && (i.Data.Name != "lb") && (i.Data.Name != "addcat") && (i.Data.Name != "cat") && (i.Data.Name != "hint") && (i.Data.Name != "stats") && (i.Data.Name != "idea") && (i.Data.Name != "about") && (i.Data.Name != "path") && (i.Data.Name != "get") && (i.Data.Name != "rmcat") {
			isMod, err := b.isMod(i.Member.User.ID, i.GuildID, bot.newMsgSlash(i))
			if rsp.Error(err) {
				return
			}
			if !isMod {
				rsp.ErrorMessage("You need to have permission `Administrator`!")
				return
			}
		}
		if i.Data.Name == "path" {
			isMod, err := b.isMod(i.Member.User.ID, i.GuildID, bot.newMsgSlash(i))
			if rsp.Error(err) {
				return
			}
			if !isMod {
				lock.RLock()
				dat, exists := b.dat[i.GuildID]
				lock.RUnlock()
				if !exists {
					rsp.ErrorMessage("You need to have permission `Administrator`!")
					return
				}
				inv, exists := dat.invCache[i.Member.User.ID]
				if !exists {
					rsp.ErrorMessage("You need to have permission `Administrator`!")
					return
				}
				_, exists = inv[strings.ToLower(i.Data.Options[0].StringValue())]
				if !exists {
					rsp.ErrorMessage("You don't have that element!")
					return
				}
			}
		}
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})
	b.dg.AddHandler(b.cmdHandler)
	b.dg.AddHandler(b.reactionHandler)
	b.dg.AddHandler(b.unReactionHandler)
	b.dg.AddHandler(b.pageSwitchHandler)
}

func commandsAreEqual(a *discordgo.ApplicationCommand, b *discordgo.ApplicationCommand) bool {
	if a.Name != b.Name || a.Description != b.Description || len(a.Options) != len(b.Options) {
		return false
	}
	for i, o1 := range a.Options {
		o2 := b.Options[i]
		if o1.Type != o2.Type || o1.Name != o2.Name || o1.Description != o2.Description || len(o1.Choices) != len(o2.Choices) {
			return false
		}
		sort.Slice(o1.Choices, func(i, j int) bool {
			return o1.Choices[i].Name < o1.Choices[j].Name
		})
		sort.Slice(o2.Choices, func(i, j int) bool {
			return o2.Choices[i].Name < o2.Choices[j].Name
		})
		for i, c1 := range o1.Choices {
			c2 := o2.Choices[i]
			if c1.Name != c2.Name || c1.Value != c2.Value {
				return false
			}
		}
	}
	return true
}
