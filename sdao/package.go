package sdao

import (
	"time"

	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

type Conversation struct {
	ID   string
	Name string
}

type ConversationDao interface {
	ListConversations() ([]Conversation, error)
	RemoveMessages(channels []Conversation, cutoffDate time.Time, dryRun bool) (int, error)
}

func retry(times int, f func() error) error {
	var lastError error
	for i := 0; i < times; i++ {
		if err := f(); err != nil {
			lastError = err
			switch v := err.(type) {
			case *slack.RateLimitedError:
				logrus.Warn("Retrying %+v of %d", v.RetryAfter, times)
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
