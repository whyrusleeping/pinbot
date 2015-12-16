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

var prefix = "!"
var gateway = "https://ipfs.io"

var (
	cmdBotsnack = prefix + "botsnack"
	cmdFriends  = prefix + "friends"
	cmdBefriend = prefix + "befriend"
	cmdShun     = prefix + "shun"
	cmdPin      = prefix + "pin"
	cmdUnPin    = prefix + "unpin"
)

var friends FriendsList

func tryPin(path string, sh *shell.Shell) (string, error) {

    if strings.HasPrefix(path, "http") {
        fmt.Print("adding link to URL")
        hash, err := sh.AddLink(path)
        fmt.Printf("added link to URL with hash %s", hash)
        if err != nil {
            fmt.Printf("failed to pin %s: %s", hash, err)
            return "", fmt.Errorf("failed to pin %s: %s", path, err)

        }
        return hash, nil
    } else if !strings.HasPrefix(path, "/ipfs") && !strings.HasPrefix(path, "/ipns") {
		path = "/ipfs/" + path
    }
    if strings.HasPrefix(path, "/ipfs") || strings.HasPrefix(path, "/ipns") {
	    out, err := sh.Refs(path, true)
	    if err != nil {
		    return "none", fmt.Errorf("failed to grab refs for %s: %s", path, err)
        }
	    // throw away results
	    for _ = range out {
	        err = sh.Pin(path)
	    }
	    if err != nil {
		    return "none", fmt.Errorf("failed to pin %s: %s", path, err)
	    }
	}

	return "none", nil
}

func tryUnpin(path string, sh *shell.Shell) error {
	out, err := sh.Refs(path, true)
	if err != nil {
		return fmt.Errorf("failed to grab refs for %s: %s", path, err)
	}

	// throw away results
	for _ = range out {
	}

	err = sh.Unpin(path)
	if err != nil {
		return fmt.Errorf("failed to pin %s: %s", path, err)
	}

	return nil
}

func Pin(b *hb.Bot, actor, path string) {

	errs := make(chan error, len(shs))
	var wg sync.WaitGroup
    urls := make(chan string, 1)

	b.Msg(actor, fmt.Sprintf("now pinning %s", path))
	// pin to every node concurrently.
	for i, sh := range shs {
	    wg.Add(1)
        go func(i int, sh *shell.Shell) {
		    defer wg.Done()
                if path, err := tryPin(path, sh); err != nil {
			        errs <- fmt.Errorf("[host %d] %s", i, err)
                    if (path != "none") {
                        urls <- path
                    }
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
    path = <-urls
	successes := len(shs) - failed
	b.Msg(actor, fmt.Sprintf("pin %d/%d successes -- %s%s", successes, len(shs), gateway, path))
}

func Unpin(b *hb.Bot, actor, path string) {
	if !strings.HasPrefix(path, "/ipfs") && !strings.HasPrefix(path, "/ipns") {
		path = "/ipfs/" + path
	}

	errs := make(chan error, len(shs))
	var wg sync.WaitGroup

	b.Msg(actor, fmt.Sprintf("now unpinning %s", path))

	// pin to every node concurrently.
	for i, sh := range shs {
		wg.Add(1)
		go func(i int, sh *shell.Shell) {
			defer wg.Done()
			if err := tryUnpin(path, sh); err != nil {
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
		b.Msg(actor, err.Error())
		failed++
	}

	successes := len(shs) - failed
	b.Msg(actor, fmt.Sprintf("unpin %d/%d successes -- %s%s", successes, len(shs), gateway, path))
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
	server := flag.String("server", "irc.rizon.net:6667", "set server to connect to")
	flag.Parse()

	for _, h := range loadHosts() {
		shs = append(shs, shell.NewShell(h))
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

	connectToIRC(con)
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

		connectToIRC(con)
		fmt.Println("Connection lost! attempting to reconnect!")
		con.Close()
	}
}

func connectToIRC(con *hb.Bot) {
	con.AddTrigger(pinTrigger)
	con.AddTrigger(unpinTrigger)
	con.AddTrigger(listTrigger)
	con.AddTrigger(befriendTrigger)
	con.AddTrigger(shunTrigger)
	con.AddTrigger(OmNomNom)
	con.AddTrigger(EatEverything)
	con.Channels = []string{
		"#omgatestchannel",
	}
	con.Run()

	for _ = range con.Incoming {
	}
}
