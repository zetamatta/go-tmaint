package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ChimeraCoder/anaconda"

	"github.com/zetamatta/go-twopane"
)

type rowT struct {
	*anaconda.Tweet
	contents []string
}

func (row *rowT) Title() string {
	return fmt.Sprintf("@%s %s", row.Tweet.User.ScreenName, row.Tweet.FullText)
}

func (row *rowT) Contents() []string {
	if row.contents == nil {
		var buffer strings.Builder
		catTweet(row.Tweet, "", "", &buffer)
		row.contents = strings.Split(buffer.String(), "\n")
	}
	return row.contents
}

func view(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := api.GetHomeTimeline(url.Values{})
	if err != nil {
		return err
	}
	rows := make([]twopane.Row, 0, len(timeline))
	uniq := make(map[string]struct{})
	for i, t := range timeline {
		rows = append(rows, &rowT{Tweet: &timeline[i]})
		uniq[t.IdStr] = struct{}{}
	}
	for {
		var nextaction func() error
		err := twopane.View{
			Rows:  rows,
			Clear: true,
			Handler: func(_ *twopane.View, key string) bool {
				switch key {
				case "r", "R", "\x12":
					nextaction = func() error {
						timeline, err := api.GetHomeTimeline(url.Values{})
						if err != nil {
							return err
						}
						newrows := make([]twopane.Row, 0, len(timeline)+len(rows))
						for i, t := range timeline {
							if _, ok := uniq[t.IdStr]; ok {
								continue
							}
							uniq[t.IdStr] = struct{}{}
							newrows = append(newrows, &rowT{Tweet: &timeline[i]})
						}
						rows = append(newrows, rows...)
						return nil
					}
					return false
				default:
					return true
				}
			},
		}.Run()

		if err != nil {
			return err
		}
		if nextaction == nil {
			return nil
		}
		if err := nextaction(); err != nil {
			return err
		}
	}
}
