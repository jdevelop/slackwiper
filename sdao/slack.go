package sdao

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

type slackDao struct {
	client   *slack.Client
	logger   *logrus.Entry
	username string
	userID   string
}

const (
	dateformat = "2006/01/02"
	dateprint  = "2006/01/02 15:04:05"
	pagesize   = 100
	retries    = 10
)

func NewSlackDao(token string, dryRun bool, userID string, log *logrus.Entry) (*slackDao, error) {
	api := slack.New(token)
	user, err := api.GetUserInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("can't get user information: %w", err)
	}
	return &slackDao{
		client:   api,
		logger:   log,
		username: user.Name,
		userID:   user.ID,
	}, nil
}

func (s *slackDao) RemoveMessages(conversations []Conversation, cutoffDate time.Time, dryRun bool) (int, error) {
	page, removed := 1, 0
	var channels = make(map[string]*Conversation)
	for i, c := range conversations {
		channels[c.ID] = &conversations[i]
	}
loop:
	for {
		var msgs *slack.SearchMessages
		err := retry(retries, func() error {
			if m, err := s.client.SearchMessages(`from:@`+s.username, slack.SearchParameters{
				Sort:          "timestamp",
				SortDirection: "asc",
				Count:         pagesize,
				Page:          page,
			}); err == nil {
				msgs = m
				return nil
			} else {
				return err
			}
		})
		if err != nil {
			return -1, fmt.Errorf("Failed to process page %d: %w", page, err)
		}
		if len(msgs.Matches) == 0 {
			break loop
		}
		s.logger.Infof("Processing page %d of %d", page, (msgs.TotalCount+pagesize-1)/pagesize)
		for _, m := range msgs.Matches {
			date, err := strconv.ParseFloat(m.Timestamp, 64)
			if err != nil {
				s.logger.Fatal(err)
			}
			if cutoffDate.Unix() < int64(date) {
				break loop
			}
			if _, ok := channels[m.Channel.ID]; ok {
				if m.User != s.userID {
					s.logger.Tracef("Skipping user %s:%s in channel %s", m.User, m.Username, m.Channel.Name)
					continue
				}
				s.logger.Infof("Removing: [%s]: %s > %s", time.Unix(int64(date), 0).Format(dateprint), m.Channel.Name, m.Text)
				if !dryRun {
					if err := retry(retries, func() error {
						_, _, err := s.client.DeleteMessage(m.Channel.ID, m.Timestamp)
						return err
					}); err != nil {
						s.logger.Errorf("Can't remove message %s from channel %s: %+v", m.Timestamp, m.Channel.Name, err)
					} else {
						removed += 1
					}
				}
			}
		}
		page += 1
		time.Sleep(500 * time.Millisecond)
	}
	return removed, nil
}

func (s *slackDao) ListConversations() ([]Conversation, error) {
	var (
		prms = slack.GetConversationsParameters{
			Types: []string{"public_channel", "private_channel", "mpim", "im"},
		}
		conversations = make([]Conversation, 0, 20)
	)
	for {
		groups, next, err := s.client.GetConversations(&prms)
		if err != nil {
			return nil, fmt.Errorf("can't get conversations: %w", err)
		}
		for _, group := range groups {
			if !group.IsMember {
				continue
			}
			s.logger.Debugf("Processing group: ID: %s, Name: %s\n", group.ID, group.Name)
			conversations = append(conversations, Conversation{ID: group.ID, Name: group.Name})
		}
		if next == "" {
			break
		}
		prms.Cursor = next
	}
	ims, err := s.client.GetIMChannels()
	if err != nil {
		return nil, fmt.Errorf("can't get IM conversations: %w", err)
	}
	for _, im := range ims {
		u, err := s.client.GetUserInfo(im.Conversation.User)
		if err != nil {
			return nil, fmt.Errorf("Can't find user: %s: %w", im.Conversation.User, err)
		}
		s.logger.Debugf("Processing direct message: ID: %s, Name: %s\n", im.ID, u.Name)
		conversations = append(conversations, Conversation{ID: u.ID, Name: u.Name})
	}
	return conversations, nil
}

var _ ConversationDao = &slackDao{}
