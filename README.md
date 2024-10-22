# gochatmail

A [chatmail](https://delta.chat/en/chatmail) server written in Go, based on [the
original Python implementation](https://github.com/deltachat/chatmail).

## What is this?

I want to improve my skills with Go, and I think that it would be good for
everyone if it were much easier to deploy a chatmail server.  Imagine if you
could stand up a chatmail server for you and your friends in 30 minutes or less
with minimal technical knowledge.  Maybe this would help put a dent in the
market shares of services like Discord and WhatsApp.  I don't think I can get it
to be as easy as the "new guild" button in Discord.  I'm hoping that I can
help make computing into a tool for creativity and expression again, instead of
a tool for surveillance and control.

In order to make that happen, I'm building a Go implementation of the chatmail
server components, wiring them up to the Go email server
[Maddy](https://maddy.email), and wrapping it all up with
[GoKrazy](https://gokrazy.org).  This makes a Raspberry Pi image or x86 virtual
machine image that you can deploy as a chatmail server very easily.

## How do I use it?

This isn't anywhere close to ready yet. Once there is something runnable, I'll
update this section with directions.

## What's on the to-do list?

### Deployment tooling
- [x] Generate chatmail server sign-up/privacy webpages
- [ ] Check for correctly set-up chatmail DNS records
- [ ] Automatically configure DNS records with popular registrars (Cloudflare,
Porkbun, Gandi, etc.)
- [ ] Build a clear, user-friendly UI for setting things up (maybe with
[bubbletea](https://github.com/charmbracelet/bubbletea) if that works well on
Windows too, or possibly a cross-platform GUI toolkit like Qt)
- [ ] Build the deployment command that actually generates a GoKrazy image with
  everything in it

### Chatmail server programs
- [x] Implement [milter](https://en.wikipedia.org/wiki/Milter) to reject
outgoing unencrypted email
- [ ] Implement SASL authentication plugin that creates accounts on first use
- [ ] Build a tiny web server that serves the sign-up/privacy webpages and
obtains HTTP-01 LetsEncrypt certificates
- [ ] Build a TLS ALPN sniffing proxy to multiplex HTTP, SMTP, and IMAP on port
  443 (for beating firewalls and increasing censorship resistance)
- [ ] Build inactive user cleanup process
- [ ] Build prometheus/openmetrics metrics endpoint
- [ ] Add `/new` endpoint to the tiny web server to generate new accounts
automatically.
- [ ] Implement push notification support for iOS/Android

### Testing
- [ ] CI of some description, so that I don't break it by accident
- [ ] Add more unit/integration tests, and check coverage metrics
- [ ] Test the documentation and UI with less technically-inclined friends to
make sure that it is clear, understandable, and easy enough to use.

## How do I help?

Open an issue or a PR!

## What is the license?

MIT, as with the upstream project. See
[LICENSE.txt](https://github.com/s0ph0s-2/gochatmail/blob/main/LICENSE.txt).
Note that this project also follows [the Delta Chat community standards and
practices](https://delta.chat/en/community-standards).
