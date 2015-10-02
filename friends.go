package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

var friendsFile = "friends"

const (
	AdminPerm = "admin"
	PinPerm   = "pin"
)

var DefaultFriendsList = FriendsList{
	friends: map[string]string{
		"whyrusleeping": "admin",
		"jbenet":        "admin",
		"lgierth":       "admin",
	},
}

type FriendsList struct {
	friends map[string]string
}

func (fl *FriendsList) CanPin(name string) bool {
	switch fl.friends[name] {
	case AdminPerm:
		return true
	case PinPerm:
		return true
	default:
		return false
	}
}

func (fl *FriendsList) CanAddFriends(name string) bool {
	switch fl.friends[name] {
	case AdminPerm:
		return true
	default:
		return false
	}
}

func (fl *FriendsList) AddFriend(name, perm string) error {
	if !validPerm(perm) {
		return fmt.Errorf("invalid perm: %s", perm)
	}

	fl.friends[name] = perm
	return fl.Write()
}

func (fl *FriendsList) RmFriend(name string) error {
	delete(fl.friends, name)
	return fl.Write()
}

func (fl *FriendsList) Write() error {
	f, err := os.Create(friendsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	for n, p := range fl.friends {
		_, err := fmt.Fprintf(f, "%s %s\n", n, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fl *FriendsList) Load() error {
	buf, err := ioutil.ReadFile(friendsFile)
	if err != nil {
		return err
	}

	// clear friends map
	f, err := fl.Parse(buf)
	if err != nil {
		return err
	}
	fl.friends = f
	return nil
}

func (fl *FriendsList) Parse(buf []byte) (f map[string]string, err error) {
	f = make(map[string]string)
	for _, l := range bytes.Split(buf, []byte("\n")) {
		if len(l) < 3 {
			continue
		}

		parts := bytes.Split(l, []byte(" "))
		if len(parts) != 2 {
			return f, fmt.Errorf("format error. too many parts. %s", parts)
		}

		user := string(parts[0])
		if len(user) < 1 {
			return f, fmt.Errorf("invalid user: %s", user)
		}

		perm := string(parts[1])
		if !validPerm(perm) {
			return f, fmt.Errorf("invalid perm: %s", perm)
		}
		f[user] = perm
	}

	// ok everything seems good
	return f, nil
}

func validPerm(perm string) bool {
	switch perm {
	case AdminPerm:
	case PinPerm:
	default:
		return false
	}
	return true
}
