package tweetanalyzer

import (
	"github.com/carbocation/go-instagram/instagram"
	"log"
	"net/url"
	"sync"
	"time"
)

type InstagramStream struct {
	strChan        chan string
	errChan        chan error
	client         *instagram.Client
	subscriptionId []string
	dictInstagram  map[string]bool
	muStart        sync.Mutex
	muMedia        sync.Mutex
	muDupl         sync.Mutex
	start          bool
	//muStart         sync.Mutex
	config     Config
	httpsender *HttpSender
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
	str, _, _ := in.httpsender.Send("https://instagram.com/", false, nil)
	in.httpsender.csrf_token, _ = in.httpsender.GetData(str, ",\"csrf_token\":\"", "\"}}")

	data := url.Values{"username": {in.config.InstagramConfig.Username},
		"password": {in.config.InstagramConfig.Password}}
	_, _, _ = in.httpsender.Send("https://instagram.com/accounts/login/ajax/", true, data)

	str2, _, _ := in.httpsender.Send("https://instagram.com/developer/clients/manage/?edited=RealtimeDataAnalysis", false, nil)
	csrfmiddlewaretoken, _ := in.httpsender.GetData(str2, "\"csrfmiddlewaretoken\" value=\"", "\"/>")
	client_id, _ := in.httpsender.GetData(str2, "<th>Client ID</th>\n"+CreateString(" ", 24)+"<td>", "</td>")

	data2 := url.Values{"csrfmiddlewaretoken": {csrfmiddlewaretoken},
		"name":         {"RealtimeDataAnalysis"},
		"description":  {"I do research in big data and want to get access to the Instagram API"},
		"website_url":  {"http://github.com/geraldstanje"},
		"redirect_uri": {"http://" + in.config.IPAddress + "/instagram"}}

	_, _, _ = in.httpsender.Send("https://instagram.com/developer/clients/"+client_id+"/edit/", true, data2)
}

func (in *InstagramStream) InstagramStream() {
	for {
		for i, _ := range in.config.InstagramConfig.Location {
			go in.getRecentMedia(int64(i))
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

	in.muMedia.Lock()
	media, _, err := in.client.Media.Search(opt)
	in.muMedia.Unlock()

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
	in.muDupl.Lock()
	defer in.muDupl.Unlock()

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

	// delete all existing subscriptions
	/*res, err := rt.instagramClient.Realtime.DeleteAllSubscriptions()
	  if err != nil {
	    fmt.Println("client.Realtime.DeleteAllSubscriptions returned error: ", err)
	    return
	  }

	  time.Sleep(1 * time.Second)

	  // subscribe to Manhattan uptown area
	  // subscribe to Manhattan downtown area
	  // subscribe to Brooklyn area
	  // subscribe to Queens area
	  // subscribe to Flushing area
	  for _, location := range rt.config.InstagramConfig.Location {
	    res, err = rt.instagramClient.Realtime.SubscribeToGeography(tweetanalyzer.FloatToString(location.Lat), tweetanalyzer.FloatToString(location.Long), "5000", "http://"+rt.config.IPAddress+":"+rt.config.Port+rt.config.InstagramConfig.CallbackURL)
	    if err != nil {
	      fmt.Println("client.Realtime.SubscribeToGeography returned error: ", err)
	      return
	    }

	    rt.subscriptionId = append(rt.subscriptionId, res.ID) //ObjectID)
	    time.Sleep(1 * time.Second)
	  }

	  time.Sleep(2 * time.Second)
	*/

	in.muStart.Lock()
	in.start = true
	in.muStart.Unlock()
}

/*
func (in *InstagramStream) InstagramHandler(w http.ResponseWriter, r *http.Request) {
  // To create a subscription, you make a POST request to the subscriptions endpoint.
  // The received GET request is the response of the subscription
  if r.Method == "GET" && r.FormValue("hub.mode") == "subscribe" && r.FormValue("hub.challenge") != "" {
    r.ParseForm()
    fmt.Fprintf(w, r.FormValue("hub.challenge"))
    // When someone posts a new photo and it triggers an update of one of your subscriptions,
    // instagram makes a POST request to the callback URL that you defined in the subscription.
    // The post body contains a raw text JSON body with update objects:
    //  {
    //      "subscription_id": "1",
    //      "object": "user",
    //      "object_id": "1234",
    //      "changed_aspect": "media",
    //      "time": 1297286541
    //  },
  } else {
    return

    defer r.Body.Close()

    var m = []instagram.RealtimeResponse{}
    err := json.NewDecoder(r.Body).Decode(&m)
    if err != nil {
      log.Println("Error: " + err.Error())
      return
    }

    in.muStart.Lock()
    defer in.muStart.Unlock()

    if in.start && len(m) > 0 {

      //found := false

      //for i, location := range rt.subscriptionId {
      //  if stringToInt(location) == int(m[0].SubscriptionID) {
        //if location == m[0].ObjectID {
      //    found = true
      //    go rt.getRecentMedia(int64(i))
      //  }
      //}

      //if !found {
      //  fmt.Println("not found Error")
      //}
    }

    w.WriteHeader(200)
  }
}
*/
