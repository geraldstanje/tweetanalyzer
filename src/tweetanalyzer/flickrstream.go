package tweetanalyzer

import (
	"fmt"
	"github.com/manki/flickgo"
	"net/http"
	"time"
)

type FlickrStream struct {
	strChan   chan string
	errChan   chan error
	client    *flickgo.Client
	dict      map[string]bool
	timestamp int64
	config    Config
}

func NewFlickrStream(strChan_ chan string, errChan_ chan error, config_ Config) *FlickrStream {
	fl := FlickrStream{
		strChan: strChan_,
		errChan: errChan_,
		config:  config_,
	}

	fl.dict = make(map[string]bool)
	fl.timestamp = time.Now().Unix()
	return &fl
}

func (fl *FlickrStream) Create() {
	apiKey := fl.config.FlickrConfig.ApiKey
	secret := fl.config.FlickrConfig.Secret
	fl.client = flickgo.New(apiKey, secret, http.DefaultClient)
}

func (fl *FlickrStream) FlickrStream() {
	for {
		fl.request(fl.timestamp)

		time.Sleep(5000 * time.Millisecond)
	}
}

func (fl *FlickrStream) getUsername(userid string, username *string) {
	args := map[string]string{
		"user_id": userid,
	}

	resp, err := fl.client.GetPeopleInfo(args)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	*username = resp.Username
}

func (fl *FlickrStream) getLocation(photoid string, lat *string, long *string) {
	args := map[string]string{
		"photo_id": photoid,
	}

	resp, err := fl.client.GetLocation(args)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	*lat = resp.Location.Latitude
	*long = resp.Location.Longitude
}

func (fl *FlickrStream) request(currtimestamp int64) {
	// https://www.flickr.com/services/api/explore/flickr.photos.search
	args := map[string]string{
		"min_upload_date": IntToString(int(currtimestamp)),
		"lat":             FloatToString(fl.config.FlickrConfig.Location.Lat),
		"lon":             FloatToString(fl.config.FlickrConfig.Location.Long),
		"radius":          fl.config.FlickrConfig.Radius,
		"radius_units":    fl.config.FlickrConfig.RadiusUnits,
		"per_page":        "500",
	}

	resp, err := fl.client.Search(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	if resp.Photos != nil {
		for _, photo := range resp.Photos {
			photoId := photo.ID

			if _, ok := fl.dict[photoId]; !ok {
				fl.dict[photoId] = true

				var lat string
				var long string
				var username string

				fl.getLocation(photoId, &lat, &long)
				fl.getUsername(photo.Owner, &username)

				link := "https://www.flickr.com/photos/" + photo.Owner + "/" + photoId

				comment := IntToString(FlickrID) + ", " +
					lat + ", " +
					long + ", " +
					username + ": " +
					"<br>" + photo.Title +
					"<br><a href=\"" + link + "\">" + link + "</a>"

				fl.strChan <- comment
			}
		}
	}
}