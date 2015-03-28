package tweetanalyzer

import (
	"fmt"
	"github.com/darkhelmet/twitterstream"
	"log"
	"regexp"
	"strings"
	"time"
)

type TwitterStream struct {
	strChan chan string
	errChan chan error
	config  Config
}

func NewTwitterStream(strChan_ chan string, errChan_ chan error, config_ Config) *TwitterStream {
	return &TwitterStream{
		strChan: strChan_,
		errChan: errChan_,
		config:  config_,
	}
}

func (t *TwitterStream) formatTwitterData(tweet *twitterstream.Tweet) (string, error) {
	var comment string
	var link string

	if tweet.Coordinates == nil {
		return "", fmt.Errorf("Tweet without geolocation set")
	}

	reg1 := regexp.MustCompile(`https?://t\.co/[a-zA-Z0-9]{0,10}`)
	reg2 := regexp.MustCompile(`pic.twitter.com/[a-zA-Z0-9]{0,10}`)

	comment = tweet.Text
	linkReg1 := reg1.FindAllString(comment, -1)
	linkReg2 := reg2.FindAllString(comment, -1)

	if linkReg1 != nil {
		comment = strings.Replace(comment, linkReg1[0], "", -1)
		link = "<a href=\"" + linkReg1[0] + "\">" + linkReg1[0] + "</a>"
	} else {
		if linkReg2 != nil {
			comment = strings.Replace(comment, linkReg2[0], "", -1)
			link = "<a href=\"" + linkReg2[0] + "\">" + linkReg2[0] + "</a>"
		}
	}

	comment = IntToString(TwitterID) + ", " +
		FloatToString(float64(tweet.Coordinates.Lat)) + ", " +
		FloatToString(float64(tweet.Coordinates.Long)) + ", " +
		tweet.User.ScreenName + ": " +
		"<br>" + comment +
		"<br>" + link
	return comment, nil
}

func (t *TwitterStream) decode(conn *twitterstream.Connection) {
	for {
		if tweet, err := conn.Next(); err == nil {
			comment, err := t.formatTwitterData(tweet)

			if err == nil {
				t.strChan <- comment
			}
		} else {
			fmt.Printf("Failed decoding tweet: %s", err)
			return
		}
	}
}

func (t *TwitterStream) TwitterStream() {
	var wait = 1
	var maxWait = 600 // Seconds

	client := twitterstream.NewClient(t.config.TwitterConfig.ConsumerKey, t.config.TwitterConfig.ConsumerSecret, t.config.TwitterConfig.AccessToken, t.config.TwitterConfig.AccessSecret)
	client.Timeout = 0

	for {
		// latitude/longitude of the locations defined in the config.xml file
		conn, err := client.Locations(twitterstream.Point{twitterstream.Latitude(t.config.TwitterConfig.Location[0].Lat), twitterstream.Longitude(t.config.TwitterConfig.Location[0].Long)},
			twitterstream.Point{twitterstream.Latitude(t.config.TwitterConfig.Location[1].Lat), twitterstream.Longitude(t.config.TwitterConfig.Location[1].Long)})

		if err != nil {
			wait = wait << 1 // exponential backoff
			log.Printf(err.Error())
			log.Printf("waiting for %d seconds before reconnect", Min(wait, maxWait))
			time.Sleep(time.Duration(Min(wait, maxWait)) * time.Second)
			continue
		} else {
			wait = 1
		}

		t.decode(conn)
	}
}
