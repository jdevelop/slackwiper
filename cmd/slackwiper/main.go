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

type Answer int

const (
	dateformat = "2006/01/02"
	tokenEnv   = "SLACK_TOKEN"

	AnswerYes Answer = iota
	AnswerNo
	AnswerSkipRest
)

var (
	user          = flag.StringP("user", "u", "", "user id")
	cutoffDateStr = flag.StringP("cutoff", "t", "", "date to retain messages after ( date yyyy/mm/dd )")
	channelStr    = flag.StringP("chats", "c", "", "comma-separated chat names to process ( empty to interactivly select ones )")
	dryRun        = flag.Bool("dry-run", true, "dry-run ( do not delete anything )")
	verbose       = flag.BoolP("verbose", "v", false, "verbose")
	quiet         = flag.BoolP("quiet", "q", false, "less output")
	attempts      = flag.IntP("attempts", "a", 1, "number of loops over the chat history")
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
		br        = bufio.NewReader(os.Stdin)
		confirm   func(string) Answer
		readYesNo = func(hasSkip bool) Answer {
			l, _, err := br.ReadLine()
			if err != nil {
				logger.Fatal(err)
			}
			switch c := string(l); c {
			case "y", "Y":
				return AnswerYes
			default:
				if hasSkip && (c == "s" || c == "S") {
					return AnswerSkipRest
				}
				return AnswerNo
			}
		}
		void     = struct{}{}
		channels = make([]sdao.Conversation, 0)
	)

	switch {
	case *quiet:
		logger.SetLevel(logrus.WarnLevel)
	case *verbose:
		logger.SetLevel(logrus.DebugLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	dao, err := sdao.NewSlackDao(token, *dryRun, *user, logrus.NewEntry(logger).WithField("UserID", *user))
	if err != nil {
		logger.Fatal(err)
	}

	if *channelStr == "" {
		confirm = func(name string) Answer {
			fmt.Printf("Wipe '%s'? [y(es)/N(o)/s(kip)] ", name)
			return readYesNo(true)
		}
	} else {
		var names = make(map[string]struct{})
		namesSlice := strings.Split(*channelStr, ",")
		for _, v := range namesSlice {
			names[strings.TrimSpace(v)] = void
		}
		confirm = func(name string) Answer {
			if _, ok := names[name]; ok {
				return AnswerYes
			} else {
				return AnswerNo
			}
		}
	}

	logger.Info("Collecting data...")

	convs, err := dao.ListConversations()
	if err != nil {
		logger.Fatal(err)
	}

answers:
	for _, v := range convs {
		switch ans := confirm(v.Name); ans {
		case AnswerYes:
			channels = append(channels, v)
		case AnswerSkipRest:
			break answers
		}
	}

	if len(channels) == 0 {
		logger.Warn("No chats selected, quitting")
		os.Exit(2)
	}

	logger.Info("Chats to wipe:")
	for _, v := range channels {
		logger.Infof("\t%s", v.Name)
	}

	fmt.Printf("Last chance: wipe %d chats? [y(es)/N(o)] ", len(channels))
	if readYesNo(false) == AnswerNo {
		logger.Println("Aborting")
		os.Exit(0)
	}

	var removed int

	for i := 0; i < *attempts; i++ {
		logger.Infof("Removal loop %d of %d", i+1, *attempts)
		rmvd, err := dao.RemoveMessages(channels, cutoffDate, *dryRun)
		if err != nil {
			logger.Fatal(err)
		}
		removed += rmvd
		time.Sleep(10 * time.Second)
	}

	logger.Printf("Removed %d messages from %d channels/DMs", removed, len(channels))
}
