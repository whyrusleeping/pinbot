# ipfs pinbot

IRC bot that pins [IPFS](http://ipfs.io) files.

Can specify either a hash (can begin with ipfs/ or ipns/) or an http:// URL.

## Usage

In a terminal
```
go get -u github.com/whyrusleeping/pinbot
pinbot [-s <server>]
```

In IRC:

```irc
<jbenet> !friends
<pinbot> my friends are: whyrusleeping lgierth jbenet
<jbenet> !pin QmbTdsZpRdVC7au7jLtkMwD6PRJPvfPvdRzG817PnxR2pR
<pinbot> now pinning QmbTdsZpRdVC7au7jLtkMwD6PRJPvfPvdRzG817PnxR2pR
<pinbot> pin QmbTdsZpRdVC7au7jLtkMwD6PRJPvfPvdRzG817PnxR2pR successful! -- http://gateway.ipfs.io/ipfs/QmbTdsZpRdVC7au7jLtkMwD6PRJPvfPvdRzG817PnxR2pR
<jbenet> !botsnack
<pinbot> om nom nom
```

Make sure to change the friends array. (or bug us to make this better configurable in an issue)
