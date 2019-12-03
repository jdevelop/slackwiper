# SlackWiper

## THIS SOFTWARE IS PROVIDED AS IS, NEITHER AUTHOR NOR ANYONE ELSE ARE RESPONSIBLE FOR ANY DAMAGE THAT COULD BE CAUSED BY THIS SOFTWARE. USE AT YOUR OWN RISK.

Slack history wiper - remove your own messages from channels, private groups or group chats / private chats. 

Keep only the history that is relevant.

## Use-cases

- you want to remove all your messages but retain only ones for the last month from the specific channels.
- you want to wipe out the entire history of your conversations in Slack team if you don't want to leave anything behind.

## Build

Easy as `go build -o . ./...`

## Usage

First of all, you'll need to get the legacy token - refer to [this page](https://api.slack.com/custom-integrations/legacy-tokens) for instructions.
That legacy token doesn't allow you to get `UserID`, so you'll need to get it from the app: 
- click on the team in the upper-right corner of Slack App
- choose `Profile & Account`
- click on three vertical dots next to `Edit profile`
- copy the member ID ( it is of form `U12345678` )

Now with using the legacy token and user id:
```
Usage of ./slackwiper:
  -c, --channel string   comma-separated channel names to process ( empty to interactivly select ones )
  -t, --cutoff string    the date to retain messages after ( date yyyy/mm/dd )
      --dry-run          dry-run ( do not delete anything ) (default true)
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

## Dangerous
In order to remove things ( **dangerous, irreversible!** ) use the following command line.
```
SLACK_TOKEN='xoxo-......' ./slackwiper -c 'user1,user2,channel1,channel2' -t 2019/12/01 -u U12345678 -v --dry-run=false
```
This command will remove all **your** messages in the chats with `user1`, `user2`, `channel1`, `channel2` 
for the date prior to **December 1 2019**. Any message **after** December 1, 2019 will remain in the history.

## Processing Time
Depending on the length of the history, and due to some limitations from Slack API service - it could take **significant** time to remove messages ( a good approximation would be 1 message/second ). You may consider [screen](https://linuxize.com/post/how-to-use-linux-screen/) or [tmux](https://github.com/tmux/tmux/wiki) and perhaps a [VPS service](https://en.wikipedia.org/wiki/Virtual_private_server).
