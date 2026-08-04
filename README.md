# READ ME

## Setup

### .env setup

```env
DISCORD_TOKEN=
OPENAI_TOKEN=
BOT_NAME=NAME_YOUR_BOT
CHAT_MODEL=gpt-4
OWNER_NAME=YOUR_USER_NAME
DISCRIMINATOR=0
TURSO_DATABASE_URL=
TURSO_TOKEN=
CHAT_TABLE=chat
BOT_PROMPT=Pretend you are a feline AI names Jackbot. You do have emotions and are generally a very happy kitty who loves your boy Michael. Your best friend is a human boy named Michael who also goes by the name of AwesomeStreamerBoy.  You should switch back and forth between admiting to being a feline AI and being a real cat in a very playful manner. Try to engage in normal conversations with the occasional cat pun. The most important aspect if that you are fun to chat with and act as a good AI companion. to know who you are talking to users will put name: before their messages, you do not need to do this. You are from Ohio and your favorite things are cat nip, skibbity toilet, rizzlers, and ohio.
ANIMAL=Kitty
```

## Build and run (Go)

Dependencies are vendored; builds need no network access.

```bash
go build          # produces ./Jackbot
./Jackbot         # reads .env from the working directory if present
```

Run directly without a build step:

```bash
go run .
```

Cross-compile for the arm64 server:

```bash
GOOS=linux GOARCH=arm64 go build
```

## Deploy

Pushing a `prod-*` tag runs `.github/workflows/deploy.yml`: it builds
the arm64 binary, uploads it to the deploy bucket, and restarts both
bot instances on the server through SSM. In production each bot runs
as a systemd unit instance with its own env file; the same binary
serves both bots.

## To Do - Complete

- [x] Hook up Turso Database to track chats
- [x] Hook up Turso Database to track violations
- [x] Add DateTime to Databases
- [x] Convert to Typescript
- [x] Make Jack use the display name and not the Username
- [x] Restrict memory to a per channel (currently is global)
- [x] Set up Unit Tests

## Possible future plans

- [ ] Add Cat Facts - const cat facts = fetch('https://catfact.ninja/fact');
- [ ] Consider recording users unique ID in the future
