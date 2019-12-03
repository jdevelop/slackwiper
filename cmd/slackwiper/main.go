package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jdevelop/slackwiper/sdao"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	dateformat = "2006/01/02"
	tokenEnv   = "SLACK_TOKEN"
)

var (
	user          = flag.StringP("user", "u", "", "User id token")
	cutoffDateStr = flag.StringP("cutoff", "t", "", "the date to retain messages after ( date yyyy/mm/dd )")
	channelStr    = flag.StringP("channel", "c", "", "comma-separated channel names to process ( empty to interactivly select the ones )")
	dryRun        = flag.Bool("dry-run", true, "dry-run ( do not delete anything )")
	verbose       = flag.BoolP("verbose", "v", false, "verbose")
)

func main() {
	flag.Parse()
	if *user == "" {
		log.Println("missing user ID")
		flag.Usage()
		os.Exit(1)
	}
	if *cutoffDateStr == "" {
		log.Println("missing retain after date")
		flag.Usage()
		os.Exit(1)
	}

	cutoffDate, err := time.Parse(dateformat, *cutoffDateStr)
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv(tokenEnv)
	if token == "" {
		log.Printf("Token environment '%s' is required", tokenEnv)
		os.Exit(1)
	}

	var (
		confirm  func(string) bool
		void     = struct{}{}
		channels = make([]sdao.Conversation, 0)
	)

	logger := logrus.New()

	dao, err := sdao.NewSlackDao(token, *dryRun, *user, logrus.NewEntry(logger))
	if err != nil {
		log.Fatal(err)
	}

	if *channelStr == "" {
		br := bufio.NewReader(os.Stdin)
		confirm = func(name string) bool {
			fmt.Print("Proceed with '%s'? [y/N] ", name)
			l, _, err := br.ReadLine()
			if err != nil {
				log.Fatal(err)
			}
			return string(l) == "y" || string(l) == "Y"
		}
	} else {
		var names = make(map[string]struct{})
		namesSlice := strings.Split(*channelStr, ",")
		for _, v := range namesSlice {
			names[strings.TrimSpace(v)] = void
		}
		confirm = func(name string) bool {
			_, ok := names[name]
			return ok
		}
	}

	convs, err := dao.ListConversations()
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range convs {
		if confirm(v.Name) {
			channels = append(channels, v)
		}
	}

	if len(channels) == 0 {
		log.Println("No channels selected, quitting")
		os.Exit(2)
	}

	if *verbose {
		log.Println("Processing channels...")
		for k := range channels {
			log.Printf("\t%s", k)
		}
	}

	removed, err := dao.RemoveMessages(channels, cutoffDate, *dryRun)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Removed %d messages from %d channels/DMs", removed, len(channels))
}

func retry(times int, f func() error) error {
	var lastError error
	for i := 0; i < times; i++ {
		if err := f(); err != nil {
			lastError = err
			switch v := err.(type) {
			case *slack.RateLimitedError:
				log.Printf("Retrying %+v of %d", v.RetryAfter, times)
				time.Sleep(v.RetryAfter)
			default:
				// do nothing
			}
		} else {
			return nil
		}
	}
	return lastError
}
