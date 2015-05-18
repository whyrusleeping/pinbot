package main

import (
	"fmt"
	"strings"

	hb "github.com/whyrusleeping/hellabot"
	shell "github.com/whyrusleeping/ipfs-shell"
)

var friends = []string{"whyrusleeping", "jbenet", "tperson"}

func isFriend(name string) bool {
	for _, n := range friends {
		if n == name {
			return true
		}
	}
	return false
}

var EatEverything = &hb.Trigger{
	func(mes *hb.Message) bool {
		return true
	},
	func(irc *hb.IrcCon, mes *hb.Message) bool {
		return true
	},
}

var authTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return true
	},
	func(con *hb.IrcCon, mes *hb.Message) {
		if isFriend(mes.From) {
			// do not consume messages from authed users
			return false
		}
		return true
	},
}

var pinTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return isFriend(mes.From) && strings.HasPrefix(mes.Content, "!pin")
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		parts := strings.Split(mes.Content, " ")
		if len(parts) == 1 {
			con.Channels[mes.To].Say("what do you want me to pin?")
		} else {
			con.Channels[mes.To].Say(fmt.Sprintf("now pinning %s", parts[1]))
			err := sh.Pin(parts[1])
			if err != nil {
				con.Channels[mes.To].Say(fmt.Sprintf("failed to pin %s: %s", parts[1], err))
			} else {
				con.Channels[mes.To].Say(fmt.Sprintf("pin %s successful!", parts[1]))
			}
		}
		return true
	},
}

var listTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return mes.Content == "!friends"
	},
	func(con *hb.IrcCon, mes *hb.Message) bool {
		out := "my friends are: "
		for _, n := range friends {
			out += n + " "
		}
		con.Channels[mes.To].Say(out)
		return true
	},
}

var sh *shell.Shell

func main() {
	sh = shell.NewShell("localhost:5001")

	con, err := hb.NewIrcConnection("irc.freenode.net:6667", "pinbot-test", false, true)
	if err != nil {
		panic(err)
	}

	con.AddTrigger(pinTrigger)
	con.AddTrigger(listTrigger)
	con.AddTrigger(EatEverything)
	con.Start()
	con.Join("#ipfs")

	for _ = range con.Incoming {
	}
}
