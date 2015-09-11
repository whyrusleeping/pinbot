package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	hb "github.com/whyrusleeping/hellabot"
	shell "github.com/whyrusleeping/ipfs-shell"
)

var prefix = "!"
var gateway = "http://gateway.ipfs.io"

var (
	cmdBotsnack = prefix + "botsnack"
	cmdFriends  = prefix + "friends"
	cmdPin      = prefix + "pin"
)

var friends = []string{
	"whyrusleeping",
	"jbenet",
	"tperson",
	"krl",
	"kyledrake",
	"zignig",
	"lgierth",
	"cryptix",
	"daviddias",
	"mafintosh",
	"davidar"
}

func isFriend(name string) bool {
	for _, n := range friends {
		if n == name {
			return true
		}
	}
	return false
}

type sayer interface {
	Say(string)
}

func tryPin(s sayer, path string, sh *shell.Shell) error {
	out, err := sh.Refs(path, true)
	if err != nil {
		return fmt.Errorf("failed to grab refs for %s: %s", path, err)
	}

	// throw away results
	for _ = range out {
	}

	err = sh.Pin(path)
	if err != nil {
		return fmt.Errorf("failed to pin %s: %s", path, err)
	}

	return nil
}

func Pin(s sayer, path string) {
	if !strings.HasPrefix(path, "/ipfs") && !strings.HasPrefix(path, "/ipns") {
		path = "/ipfs/" + path
	}

	errs := make(chan error, len(shs))
	var wg sync.WaitGroup

	s.Say(fmt.Sprintf("now pinning %s", path))

	// pin to every node concurrently.
	for i, sh := range shs {
		wg.Add(1)
		go func(i int, sh *shell.Shell) {
			defer wg.Done()
			if err := tryPin(s, path, sh); err != nil {
				errs <- fmt.Errorf("[host %d] %s", i, err)
			}
		}(i, sh)
	}

	// close the err chan when done.
	go func() {
		wg.Wait()
		close(errs)
	}()

	// wait on the err chan and print every err we get as we get it.
	var failed int
	for err := range errs {
		s.Say(err.Error())
		failed++
	}

	successes := len(shs) - failed
	s.Say(fmt.Sprintf("pin %d/%d successes -- %s%s", successes, len(shs), gateway, path))
}

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
		if isFriend(mes.From) {
			// do not consume messages from authed users
			return false
		}
		return true
	},
}

var pinTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return isFriend(mes.From) && strings.HasPrefix(mes.Content, cmdPin)
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

var listTrigger = &hb.Trigger{
	func(mes *hb.Message) bool {
		return mes.Content == cmdFriends
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

var shs []*shell.Shell

func loadHosts() []string {
	fi, err := os.Open("hosts")
	if err != nil {
		fmt.Println("failed to open hosts file, defaulting to localhost:5001")
		return []string{"localhost:5001"}
	}

	var hosts []string
	scan := bufio.NewScanner(fi)
	for scan.Scan() {
		hosts = append(hosts, scan.Text())
	}
	return hosts
}

func main() {
	name := flag.String("name", "pinbot-test", "set pinbots name")
	server := flag.String("server", "irc.freenode.net:6667", "set server to connect to")
	flag.Parse()

	for _, h := range loadHosts() {
		shs = append(shs, shell.NewShell(h))
	}

	con, err := hb.NewIrcConnection(*server, *name, false, true)
	if err != nil {
		panic(err)
	}

	connectToFreenodeIpfs(con)
	fmt.Println("Connection lost! attempting to reconnect!")
	con.Close()

	recontime := time.Second
	for {
		// Dont try to reconnect this time
		con, err := hb.NewIrcConnection(*server, *name, false, false)
		if err != nil {
			fmt.Println("ERROR CONNECTING: ", err)
			time.Sleep(recontime)
			recontime += time.Second
			continue
		}
		recontime = time.Second

		connectToFreenodeIpfs(con)
		fmt.Println("Connection lost! attempting to reconnect!")
		con.Close()
	}
}

func connectToFreenodeIpfs(con *hb.IrcCon) {
	con.AddTrigger(pinTrigger)
	con.AddTrigger(listTrigger)
	con.AddTrigger(OmNomNom)
	con.AddTrigger(EatEverything)
	con.Start()
	con.Join("#ipfs")

	for _ = range con.Incoming {
	}
}
