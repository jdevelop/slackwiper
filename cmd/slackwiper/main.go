package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jdevelop/slackwiper/sdao"
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
	quiet         = flag.BoolP("quiet", "q", false, "less output")
)

func main() {
	flag.Parse()
	logger := logrus.New()
	if *user == "" {
		logger.Error("missing user ID")
		flag.Usage()
		os.Exit(1)
	}
	if *cutoffDateStr == "" {
		logger.Error("missing cutoff date")
		flag.Usage()
		os.Exit(1)
	}

	cutoffDate, err := time.Parse(dateformat, *cutoffDateStr)
	if err != nil {
		logger.Fatal(err)
	}

	token := os.Getenv(tokenEnv)
	if token == "" {
		logger.Fatalf("Token environment '%s' is required", tokenEnv)
	}

	var (
		confirm  func(string) bool
		void     = struct{}{}
		channels = make([]sdao.Conversation, 0)
	)

	switch {
	case *quiet:
		logger.SetLevel(logrus.InfoLevel)
	case *verbose:
		logger.SetLevel(logrus.DebugLevel)
	default:
		logger.SetLevel(logrus.WarnLevel)
	}

	dao, err := sdao.NewSlackDao(token, *dryRun, *user, logrus.NewEntry(logger))
	if err != nil {
		logger.Fatal(err)
	}

	if *channelStr == "" {
		br := bufio.NewReader(os.Stdin)
		confirm = func(name string) bool {
			fmt.Print("Proceed with '%s'? [y/N] ", name)
			l, _, err := br.ReadLine()
			if err != nil {
				logger.Fatal(err)
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
		logger.Fatal(err)
	}

	for _, v := range convs {
		if confirm(v.Name) {
			channels = append(channels, v)
		}
	}

	if len(channels) == 0 {
		logger.Warn("No channels selected, quitting")
		os.Exit(2)
	}

	if *verbose {
		logger.Info("Processing channels...")
		for k := range channels {
			logger.Infof("\t%s", k)
		}
	}

	removed, err := dao.RemoveMessages(channels, cutoffDate, *dryRun)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("Removed %d messages from %d channels/DMs", removed, len(channels))
}
