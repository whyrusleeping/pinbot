package main

import (
	"strings"

	hb "github.com/whyrusleeping/hellabot"
)

var EatEverything = &hb.Trigger{
	func(mes *hb.Message) bool {
		return true
	},
	func(irc *hb.IrcCon, mes *hb.Message) bool {
		return true
	},
}

var OmNomNom = &hb.Trigger{
	func(mes *hb.Message) bool {
		return mes.Content == cmdBotsnack
	},
	func(irc *hb.IrcCon, mes *hb.Message) bool {
		irc.Channels[mes.To].Say("om nom nom")
		return true
	},
}

var authTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return true
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		if friends.CanPin(mes.From) {
			// do not consume messages from authed users
			return false
		}
		return true
	},
}

var pinTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return friends.CanPin(mes.From) && strings.HasPrefix(mes.Content, cmdPin)
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		parts := strings.Split(mes.Content, " ")
		if len(parts) == 1 {
			con.Channels[mes.To].Say("what do you want me to pin?")
		} else {
			Pin(con.Channels[mes.To], parts[1])
		}
		return true
	},
}

var unpinTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return friends.CanPin(mes.From) && strings.HasPrefix(mes.Content, cmdUnPin)
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		parts := strings.Split(mes.Content, " ")
		if len(parts) == 1 {
			con.Channels[mes.To].Say("what do you want me to unpin?")
		} else {
			Unpin(con.Channels[mes.To], parts[1])
		}
		return true
	},
}

var listTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return mes.Content == cmdFriends
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		out := "my friends are: "
		for n, _ := range friends.friends {
			out += n + " "
		}
		con.Channels[mes.To].Say(out)
		return true
	},
}

var befriendTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return friends.CanAddFriends(mes.From) &&
			strings.HasPrefix(mes.Content, cmdBefriend)
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		parts := strings.Split(mes.Content, " ")
		if len(parts) != 3 {
			con.Channels[mes.To].Say(cmdBefriend + " <name> <perm>")
			return true
		}
		name := parts[1]
		perm := parts[2]

		if err := friends.AddFriend(name, perm); err != nil {
			con.Channels[mes.To].Say("failed to befriend: " + err.Error())
			return true
		}
		con.Channels[mes.To].Say("Hey " + name + ", let's be friends! You can " + perm)
		return true
	},
}

var shunTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return friends.CanAddFriends(mes.From) &&
			strings.HasPrefix(mes.Content, cmdShun)
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		parts := strings.Split(mes.Content, " ")
		if len(parts) != 2 {
			con.Channels[mes.To].Say("who do you want me to shun?")
			return true
		}

		name := parts[1]
		if err := friends.RmFriend(name); err != nil {
			con.Channels[mes.To].Say("failed to shun: " + err.Error())
			return true
		}
		con.Channels[mes.To].Say("shun " + name + " the non believer! Shuuuuuuuun")
		return true
	},
}
