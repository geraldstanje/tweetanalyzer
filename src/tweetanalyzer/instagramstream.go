package tweetanalyzer

import (
	"github.com/carbocation/go-instagram/instagram"
	"log"
	"net/url"
	"time"
)

type InstagramStream struct {
	strChan        chan string
	errChan        chan error
	client         *instagram.Client
	subscriptionId []string
	dictInstagram  map[string]bool
	config         Config
	httpsender     *HttpSender
}

func NewInstagramStream(strChan_ chan string, errChan_ chan error, config_ Config) *InstagramStream {
	in := InstagramStream{
		strChan: strChan_,
		errChan: errChan_,
		config:  config_,
	}

	in.dictInstagram = make(map[string]bool)
	in.httpsender = NewHttpSender()
	return &in
}

func (in *InstagramStream) SetRedirectIP(ip string) {
	header_data := make(map[string]string)
	header_data["Host"] = "instagram.com"
	header_data["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
	header_data["User-Agent"] = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:16.0) Gecko/20100101 Firefox/16.0"
	header_data["Referer"] = "https://instagram.com"
	header_data["Host"] = "instagram.com"
	header_data["Accept"] = "*/*"

	str, _, _ := in.httpsender.Send("https://instagram.com/", false, nil, header_data)
	csrf_token, _ := in.httpsender.GetData(str, ",\"csrf_token\":\"", "\"}}")
	header_data["X-CSRFToken"] = csrf_token
	header_data["X-Instagram-AJAX"] = "1"
	header_data["X-Requested-With"] = "XMLHttpRequest"

	data := url.Values{"username": {in.config.InstagramConfig.Username},
		"password": {in.config.InstagramConfig.Password}}
	_, _, _ = in.httpsender.Send("https://instagram.com/accounts/login/ajax/", true, data, header_data)

	str2, _, _ := in.httpsender.Send("https://instagram.com/developer/clients/manage/?edited=RealtimeDataAnalysis", false, nil, header_data)
	csrfmiddlewaretoken, _ := in.httpsender.GetData(str2, "\"csrfmiddlewaretoken\" value=\"", "\"/>")
	client_id, _ := in.httpsender.GetData(str2, "<th>Client ID</th>\n"+CreateString(" ", 24)+"<td>", "</td>")

	data2 := url.Values{"csrfmiddlewaretoken": {csrfmiddlewaretoken},
		"name":         {"RealtimeDataAnalysis"},
		"description":  {"I do research in big data and want to get access to the Instagram API"},
		"website_url":  {"http://github.com/geraldstanje"},
		"redirect_uri": {"http://" + in.config.IPAddress + "/instagram"}}

	_, _, _ = in.httpsender.Send("https://instagram.com/developer/clients/"+client_id+"/edit/", true, data2, header_data)
}

func (in *InstagramStream) InstagramStream() {
	for {
		for i, _ := range in.config.InstagramConfig.Location {
			in.getRecentMedia(int64(i))
		}

		time.Sleep(5000 * time.Millisecond)
	}
}

func (in *InstagramStream) getRecentMedia(subscriptionId int64) {
	opt := &instagram.Parameters{
		Lat:      in.config.InstagramConfig.Location[subscriptionId].Lat,
		Lng:      in.config.InstagramConfig.Location[subscriptionId].Long,
		Distance: 5000,
	}

	media, _, err := in.client.Media.Search(opt)

	if err != nil {
		log.Println("Error: ", instagram.ErrorResponse(*in.client.Response))
		return
	}

	if len(media) > 0 {
		for i, _ := range media {
			if in.isDuplicate(media[i].ID) {
				return
			}

			comment := in.formatInstagramData(media[i])
			in.strChan <- comment
		}
	}
}

func (in *InstagramStream) isDuplicate(mediaId string) bool {
	if _, ok := in.dictInstagram[mediaId]; ok {
		return true
	} else {
		in.dictInstagram[mediaId] = true
		return false
	}
}

func (in *InstagramStream) formatInstagramData(media instagram.Media) string {
	var comment string

	comment = IntToString(InstagramID) + ", " +
		FloatToString(media.Location.Latitude) + ", " +
		FloatToString(media.Location.Longitude) + ", " +
		media.User.Username + ": "

	if media.Caption != nil {
		comment += "<br>" + media.Caption.Text + " " +
			"<br><a href=\"" + media.Link + "\">" + media.Link + "</a>"
	} else {
		comment += "<br><a href=\"" + media.Link + "\">" + media.Link + "</a>"
	}

	return comment
}

func (in *InstagramStream) Create() {
	in.client = instagram.NewClient(nil)
	in.client.ClientID = in.config.InstagramConfig.ClientID
	in.client.ClientSecret = in.config.InstagramConfig.ClientSecret
	in.client.AccessToken = in.config.InstagramConfig.AccessToken
}
