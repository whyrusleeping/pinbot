package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	shell "github.com/whyrusleeping/pinbot/Godeps/_workspace/src/github.com/ipfs/go-ipfs-api"
	hb "github.com/whyrusleeping/pinbot/Godeps/_workspace/src/github.com/whyrusleeping/hellabot"
)

var prefix string
var gateway string

var (
	cmdBotsnack = prefix + "botsnack"
	cmdFriends  = prefix + "friends"
	cmdBefriend = prefix + "befriend"
	cmdShun     = prefix + "shun"
	cmdPin      = prefix + "pin"
	cmdUnPin    = prefix + "unpin"
)

var friends FriendsList

func formatError(action string, err error) error {
	if strings.Contains(err.Error(), "504 Gateway Time-out") {
		err = fmt.Errorf("504 Gateway Time-out")
	}
	if strings.Contains(err.Error(), "502 Bad Gateway") {
		err = fmt.Errorf("502 Bad Gateway")
	}

	return fmt.Errorf("%s failed: %s", action, err.Error())
}

func tryPin(path string, sh *shell.Shell) error {
	out, err := sh.Refs(path, true)
	if err != nil {
		return formatError("refs", err)
	}

	// throw away results
	for _ = range out {
	}

	err = sh.Pin(path)
	if err != nil {
		return formatError("pin", err)
	}

	return nil
}

func tryUnpin(path string, sh *shell.Shell) error {
	out, err := sh.Refs(path, true)
	if err != nil {
		return formatError("refs", err)
	}

	// throw away results
	for _ = range out {
	}

	err = sh.Unpin(path)
	if err != nil {
		return formatError("unpin", err)
	}

	return nil
}

var pinfile = "pins.log"

func writePin(pin, label string) error {
	fi, err := os.OpenFile(pinfile, os.O_APPEND|os.O_EXCL|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fi, "%s\t%s\n", pin, label)
	if err != nil {
		return err
	}
	return fi.Close()
}

func Pin(b *hb.Bot, actor, path, label string) {
	if !strings.HasPrefix(path, "/ipfs") && !strings.HasPrefix(path, "/ipns") {
		path = "/ipfs/" + path
	}

	errs := make(chan error, len(shs))
	var wg sync.WaitGroup

	b.Msg(actor, fmt.Sprintf("now pinning on %d nodes", len(shs)))

	// pin to every node concurrently.
	for i, sh := range shs {
		wg.Add(1)
		go func(i int, sh *shell.Shell) {
			defer wg.Done()
			if err := tryPin(path, sh); err != nil {
				errs <- fmt.Errorf("%s -- %s", shsUrls[i], err)
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
		b.Msg(actor, err.Error())
		failed++
	}

	successes := len(shs) - failed
	b.Msg(actor, fmt.Sprintf("pinned on %d of %d nodes (%d failures) -- %s%s",
		successes, len(shs), failed, gateway, path))

	if err := writePin(path, label); err != nil {
		b.Msg(actor, fmt.Sprintf("failed to write log entry for last pin: %s", err))
	}
}

func Unpin(b *hb.Bot, actor, path string) {
	if !strings.HasPrefix(path, "/ipfs") && !strings.HasPrefix(path, "/ipns") {
		path = "/ipfs/" + path
	}

	errs := make(chan error, len(shs))
	var wg sync.WaitGroup

	b.Msg(actor, fmt.Sprintf("now unpinning on %d nodes", len(shs)))

	// pin to every node concurrently.
	for i, sh := range shs {
		wg.Add(1)
		go func(i int, sh *shell.Shell) {
			defer wg.Done()
			if err := tryUnpin(path, sh); err != nil {
				errs <- fmt.Errorf("%s -- %s", shsUrls[i], err)
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
		b.Msg(actor, err.Error())
		failed++
	}

	successes := len(shs) - failed
	b.Msg(actor, fmt.Sprintf("unpinned on %d of %d nodes (%d failures) -- %s%s",
		successes, len(shs), failed, gateway, path))
}

var shs []*shell.Shell
var shsUrls []string

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

func ensurePinLogExists() error {
	_, err := os.Stat(pinfile)
	if os.IsNotExist(err) {
		fi, err := os.Create(pinfile)
		if err != nil {
			return err
		}

		fi.Close()
	}
	return nil
}

func main() {
	name := flag.String("name", "pinbot-test", "set pinbot's nickname")
	server := flag.String("server", "irc.freenode.net:6667", "set server to connect to")
	channel := flag.String("channel", "#pinbot-test", "set channel to join")
	pre := flag.String("prefix", "!", "prefix of command messages")
	gw := flag.String("gateway", "https://ipfs.io", "IPFS-to-HTTP gateway to use for success messages")
	flag.Parse()

	prefix = *pre
	gateway = *gw

	err := ensurePinLogExists()
	if err != nil {
		panic(err)
	}

	for _, h := range loadHosts() {
		shs = append(shs, shell.NewShell(h))
		shsUrls = append(shsUrls, "http://"+h)
	}

	if err := friends.Load(); err != nil {
		if os.IsNotExist(err) {
			friends = DefaultFriendsList
		} else {
			panic(err)
		}
	}
	fmt.Println("loaded", len(friends.friends), "friends")

	con, err := hb.NewBot(*server, *name, hb.ReconOpt())
	if err != nil {
		panic(err)
	}

	connectToFreenodeIpfs(con, *channel)
	fmt.Println("Connection lost! attempting to reconnect!")
	con.Close()

	recontime := time.Second
	for {
		// Dont try to reconnect this time
		con, err := hb.NewBot(*server, *name)
		if err != nil {
			fmt.Println("ERROR CONNECTING: ", err)
			time.Sleep(recontime)
			recontime += time.Second
			continue
		}
		recontime = time.Second

		connectToFreenodeIpfs(con, *channel)
		fmt.Println("Connection lost! attempting to reconnect!")
		con.Close()
	}
}

func connectToFreenodeIpfs(con *hb.Bot, channel string) {
	con.AddTrigger(pinTrigger)
	con.AddTrigger(unpinTrigger)
	con.AddTrigger(listTrigger)
	con.AddTrigger(befriendTrigger)
	con.AddTrigger(shunTrigger)
	con.AddTrigger(OmNomNom)
	con.AddTrigger(EatEverything)
	con.Channels = []string{channel}
	con.Run()

	for _ = range con.Incoming {
	}
}
