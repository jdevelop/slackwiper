# SlackWiper

## THIS SOFTWARE IS PROVIDED AS IS, NO WARRANTY, NO LIABILITY. NEITHER AUTHOR NOR ANYONE ELSE ARE RESPONSIBLE FOR ANY DAMAGE THAT COULD BE CAUSED BY THIS SOFTWARE. USE AT YOUR OWN RISK.

Slack history wiper - remove your messages from channels, private groups or group chats / private chats. 

Keep only the history that is relevant.

## Use-cases

- you want to remove all your messages but retain only ones for the last month from the specific channels.
- retain messages option is disabled by a Slack Workspace administrator
- you want to wipe out the entire history of your conversations in Slack team

## Build

Easy as `go build -o . ./...`

## Download

[Download](https://github.com/jdevelop/slackwiper/releases/tag/v1.0.0) binaries for Windows, MacOS, Linux (x64).

## Usage

First of all, you'll need to get the legacy token - refer to [this page](https://api.slack.com/custom-integrations/legacy-tokens) for instructions.
That legacy token doesn't allow you to get `UserID`, so you'll need to get it from the Slack app/Website: 
- click on the team in the upper-right corner of Slack App/Website
- choose `Profile & Account`
- click on three vertical dots next to `Edit profile`
- copy the member ID ( a string like `U12345678` )

Now with using the legacy token and user id:
```
Usage of ./slackwiper:
  -c, --chats string     comma-separated chat names to process ( empty to interactivly select ones )
  -t, --cutoff string    date to retain messages after ( date yyyy/mm/dd )
      --dry-run          dry-run ( do not delete anything ) (default true)
  -q, --quiet            less output
  -u, --user string      User id token
  -v, --verbose          verbose
pflag: help requested

```

`-t` option is **crucial** - the messages **before** this date will be removed. The messages after this date will be preserved.

an example:
```
SLACK_TOKEN='xoxo-......' ./slackwiper -c 'user1,user2,channel1,channel2' -t 2019/12/01 -u U12345678 -v
```
by default it won't remove anything ( that `--dry-run` option is on, just in case ).

## Interactive mode
If you don't provide `-c` argument - then slackwiper will read the available channels/groups and direct messages options from Slack and ask which ones to include:

```
SLACK_TOKEN='xoxo-......' ./slackwiper -c 'user1,user2,channel1,channel2' -t 2019/12/01 -u U12345678 -v
INFO[0000] Collecting data...
Wipe 'general'? [y/N/s] y
Wipe 'release'? [y/N/s]
Wipe 'golang'? [y/N/s] s
```

The available options are:
- y - yes, include this channel or chat
- n - no ( default one, used if no input provided )
- s - skip to the end

With `s` option you may want to skip the rest of the chats if you have already selected all the chats you want to clear.


## Dangerous
In order to actually remove things ( **dangerous**, **irreversible!** ) use the command line option `--dry-run=false`
```
SLACK_TOKEN='xoxo-......' ./slackwiper -c 'user1,user2,channel1,channel2' -t 2019/12/01 -u U12345678 -v --dry-run=false
```
This command will remove all **your** messages in the chats with `user1`, `user2`, `channel1`, `channel2` 
for the date prior to **December 1 2019**. Any message **after** December 1, 2019 will remain in the history.

## Processing Time
Depending on the length of the history, and due to some limitations from Slack API service - it could take **significant** time to remove messages ( a good approximation would be 1 message/second ). You may consider [screen](https://linuxize.com/post/how-to-use-linux-screen/) or [tmux](https://github.com/tmux/tmux/wiki) and perhaps a [VPS service](https://en.wikipedia.org/wiki/Virtual_private_server).
