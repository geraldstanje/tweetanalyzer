package main

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/carbocation/go-instagram/instagram"
	"github.com/darkhelmet/twitterstream"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const errorCounterMax = 3

// Client connection consists of the websocket and the client ip
type Client struct {
	errorCount int
	websocket  *websocket.Conn
	clientIP   string
}

type RealtimeAnalyzer struct {
	config          Config
	start           bool
	activeClients   map[string]Client
	instagramClient *instagram.Client
	subscriptionId  []string
	strChan         chan string
	errChan         chan error
	points          [][]float64
	dictInstagram   map[string]bool
	muMedia         sync.Mutex
	muDupl          sync.Mutex
	muStart         sync.Mutex
}

type Config struct {
	IPAddress       string          `xml:"ip"`
	Port            string          `xml:"port"`
	TwitterConfig   TwitterConfig   `xml:"twitter"`
	InstagramConfig InstagramConfig `xml:"instagram"`
}

type TwitterConfig struct {
	ConsumerKey    string        `xml:"consumerKey"`
	ConsumerSecret string        `xml:"consumerSecret"`
	AccessToken    string        `xml:"accessToken"`
	AccessSecret   string        `xml:"accessSecret"`
	Location       []GeoLocation `xml:"location"`
}

type InstagramConfig struct {
	ClientID     string        `xml:"clientID"`
	ClientSecret string        `xml:"clientSecret"`
	AccessToken  string        `xml:"accessToken"`
	CallbackURL  string        `xml:"callbackURL"`
	Location     []GeoLocation `xml:"location"`
}

type GeoLocation struct {
	Lat  float64 `xml:"lat,attr"`
	Long float64 `xml:"long,attr"`
}

// to convert a float number to a string
func floatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func stringToFloat(str string) float64 {
	floatVal, _ := strconv.ParseFloat(str, 64)
	return floatVal
}

func random(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func parseXML(reader *bufio.Reader) (Config, error) {
	config := Config{}
	err := xml.NewDecoder(reader).Decode(&config)
	return config, err
}

func (rt *RealtimeAnalyzer) readConfig(filename string) error {
	xmlFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer xmlFile.Close()

	reader := bufio.NewReader(xmlFile)
	rt.config, err = parseXML(reader)

	return err
}

func (rt *RealtimeAnalyzer) changeIPAddress(filename string, newStr string) error {
	if len(filename) == 0 {
		return fmt.Errorf("Error: invalid len of file")
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	reg := regexp.MustCompile(`[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}:[0-9]{2,5}/sock";`)
	ips := reg.FindAllString(string(b), -1)

	for _, ip := range ips {
		newStr += "/sock\";"
		b = bytes.Replace(b, []byte(ip), []byte(newStr), -1)
	}

	err = ioutil.WriteFile(filename, b, 0644)
	return err
}

// handler for the main page
func HomeHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-type", "text/html")
	webpage, err := ioutil.ReadFile("home.html")

	if err != nil {
		http.Error(response, fmt.Sprintf("home.html file error %v", err), 500)
	}

	fmt.Fprint(response, string(webpage))
}

// reference: http://alienryderflex.com/polygon/
func (rt *RealtimeAnalyzer) isPointInPolygon(longitude float64, latitude float64) bool {
	rt.points = [][]float64{{40.703286, -74.017739},
		{40.735551, -74.010487},
		{40.752979, -74.007397},
		{40.815891, -73.960540},
		{40.800966, -73.929169},
		{40.783921, -73.94145},
		{40.776122, -73.941965},
		{40.739974, -73.972864},
		{40.729308, -73.971663},
		{40.711614, -73.978014},
		{40.706148, -74.00239},
		{40.702114, -74.009671},
		{40.701203, -74.015164}}

	var i int
	var j int
	var odd_nodes bool
	var number_of_points int

	number_of_points = len(rt.points)
	j = number_of_points - 1
	odd_nodes = false

	for i = 0; i < number_of_points; i++ {
		if rt.points[i][1] < latitude && rt.points[j][1] >= latitude || rt.points[j][1] < latitude && rt.points[i][1] >= latitude {
			if rt.points[i][0]+(latitude-rt.points[i][1])/(rt.points[j][1]-rt.points[i][1])*(rt.points[j][0]-rt.points[i][0]) < longitude {
				odd_nodes = !odd_nodes
			}
		}
		j = i
	}

	return odd_nodes
}

func (rt *RealtimeAnalyzer) minElement(col int) float64 {
	min := math.MaxFloat64

	for i := 0; i < len(rt.points); i++ {
		if rt.points[i][col] < min {
			min = rt.points[i][col]
		}
	}

	return min
}

func (rt *RealtimeAnalyzer) maxElement(col int) float64 {
	max := math.SmallestNonzeroFloat64

	for i := 0; i < len(rt.points); i++ {
		if rt.points[i][col] > max {
			max = rt.points[i][col]
		}
	}

	return max
}

func (rt *RealtimeAnalyzer) generateRandGeoLoc() (float64, float64) {
	var longitude float64
	var latitude float64

	for {
		longitude = random(rt.minElement(0), rt.maxElement(0))
		latitude = random(rt.minElement(1), rt.maxElement(1))

		retval := rt.isPointInPolygon(longitude, latitude)
		if retval {
			break
		}
	}

	return longitude, latitude
}

func (rt *RealtimeAnalyzer) generateGeoData() {
	for {
		longitude, latitude := rt.generateRandGeoLoc()

		str := "0" + ", " +
			floatToString(longitude) + ", " +
			floatToString(latitude) + ", " +
			"tweet"

		rt.strChan <- str

		time.Sleep(200 * time.Millisecond)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func (rt *RealtimeAnalyzer) broadcastData() {
	var Message = websocket.Message
	var err error
	tweet_count := 0

	for {
		select {
		case str := <-rt.strChan:
			for ip, _ := range rt.activeClients {
				if err = Message.Send(rt.activeClients[ip].websocket, str); err != nil {
					// we could not send the message to a peer
					log.Println("Could not send message to ", ip, err.Error())

					// work-around: https://code.google.com/p/go/issues/detail?id=3117
					var tmp = rt.activeClients[ip]
					tmp.errorCount += 1
					rt.activeClients[ip] = tmp

					if rt.activeClients[ip].errorCount >= errorCounterMax {
						log.Println("Client disconnected:", ip)
						delete(rt.activeClients, ip)
					}
				}
			}

		default:
			continue

			tweet_count += 1
			fmt.Println(tweet_count)
		}
	}
}

func (rt *RealtimeAnalyzer) formatTwitterData(tweet *twitterstream.Tweet) (string, error) {
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

	comment = "0" + ", " +
		floatToString(float64(tweet.Coordinates.Lat)) + ", " +
		floatToString(float64(tweet.Coordinates.Long)) + ", " +
		tweet.User.ScreenName + ": " +
		"<br>" + comment +
		"<br>" + link
	return comment, nil
}

func (rt *RealtimeAnalyzer) decode(conn *twitterstream.Connection) {
	for {
		if tweet, err := conn.Next(); err == nil {
			comment, err := rt.formatTwitterData(tweet)

			if err == nil {
				rt.strChan <- comment
			}
		} else {
			fmt.Printf("Failed decoding tweet: %s", err)
			return
		}
	}
}

func (rt *RealtimeAnalyzer) twitterStream() {
	var wait = 1
	var maxWait = 600 // Seconds

	client := twitterstream.NewClient(rt.config.TwitterConfig.ConsumerKey, rt.config.TwitterConfig.ConsumerSecret, rt.config.TwitterConfig.AccessToken, rt.config.TwitterConfig.AccessSecret)
	client.Timeout = 0

	for {
		// latitude/longitude of New York City
		conn, err := client.Locations(twitterstream.Point{twitterstream.Latitude(rt.config.TwitterConfig.Location[0].Lat), twitterstream.Longitude(rt.config.TwitterConfig.Location[0].Long)},
			twitterstream.Point{twitterstream.Latitude(rt.config.TwitterConfig.Location[1].Lat), twitterstream.Longitude(rt.config.TwitterConfig.Location[1].Long)})

		if err != nil {
			wait = wait << 1 // exponential backoff
			log.Printf(err.Error())
			log.Printf("waiting for %d seconds before reconnect", min(wait, maxWait))
			time.Sleep(time.Duration(min(wait, maxWait)) * time.Second)
			continue
		} else {
			wait = 1
		}

		rt.decode(conn)
	}
}

// reference: https://github.com/Niessy/websocket-golang-chat
// WebSocket server to handle clients
func (rt *RealtimeAnalyzer) WebSocketServer(ws *websocket.Conn) {
	var err error

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	client := ws.Request().RemoteAddr
	log.Println("New client connected:", client)
	rt.activeClients[client] = Client{0, ws, client}

	// for loop so the websocket stays open otherwise it'll close
	for {
		time.Sleep(1 * time.Second)
	}
}

// http://instagram.com/developer/clients/manage/?edited=RealtimeDataAnalysis
func (rt *RealtimeAnalyzer) InstagramStream() {
	rt.instagramClient = instagram.NewClient(nil)
	rt.instagramClient.ClientID = rt.config.InstagramConfig.ClientID
	rt.instagramClient.ClientSecret = rt.config.InstagramConfig.ClientSecret
	rt.instagramClient.AccessToken = rt.config.InstagramConfig.AccessToken

	// delete all existing subscriptions
	res, err := rt.instagramClient.Realtime.DeleteAllSubscriptions()
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
		res, err = rt.instagramClient.Realtime.SubscribeToGeography(floatToString(location.Lat), floatToString(location.Long), "5000", "http://"+rt.config.IPAddress+":"+rt.config.Port+rt.config.InstagramConfig.CallbackURL)
		if err != nil {
			fmt.Println("client.Realtime.SubscribeToGeography returned error: ", err)
			return
		}
		rt.subscriptionId = append(rt.subscriptionId, res.ObjectID)

		time.Sleep(1 * time.Second)
	}

	time.Sleep(2 * time.Second)

	rt.muStart.Lock()
	rt.start = true
	rt.muStart.Unlock()
}

func (rt *RealtimeAnalyzer) formatInstagramData(media instagram.Media) string {
	var comment string

	comment = "1" + ", " +
		floatToString(media.Location.Latitude) + ", " +
		floatToString(media.Location.Longitude) + ", " +
		media.User.Username + ": "

	if media.Caption != nil {
		comment += "<br>" + media.Caption.Text + " " +
			"<br><a href=\"" + media.Link + "\">" + media.Link + "</a>"
	}

	return comment
}

func (rt *RealtimeAnalyzer) isDuplicate(mediaId string) bool {
	rt.muDupl.Lock()
	defer rt.muDupl.Unlock()

	if _, ok := rt.dictInstagram[mediaId]; ok {
		return true
	} else {
		rt.dictInstagram[mediaId] = true
		return false
	}
}

func (rt *RealtimeAnalyzer) getRecentMedia(subscriptionId int64) {
	opt := &instagram.Parameters{
		Lat:      rt.config.InstagramConfig.Location[subscriptionId].Lat,
		Lng:      rt.config.InstagramConfig.Location[subscriptionId].Long,
		Distance: 5000,
	}

	rt.muMedia.Lock()
	media, _, err := rt.instagramClient.Media.Search(opt)
	rt.muMedia.Unlock()

	if err != nil {
		log.Println("Error: ", instagram.ErrorResponse(*rt.instagramClient.Response))
		return
	}

	if len(media) > 0 {
		if rt.isDuplicate(media[0].ID) {
			return
		}

		comment := rt.formatInstagramData(media[0])
		rt.strChan <- comment
	}
}

func (rt *RealtimeAnalyzer) instagramHandler(w http.ResponseWriter, r *http.Request) {
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
		defer r.Body.Close()

		var m = []instagram.RealtimeResponse{}
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			log.Println("Error: " + err.Error())
			return
		}

		rt.muStart.Lock()
		defer rt.muStart.Unlock()

		if rt.start && len(m) > 0 {
			found := false

			for i, location := range rt.subscriptionId {
				if location == m[0].ObjectID {
					found = true
					go rt.getRecentMedia(int64(i))
				}
			}

			if !found {
				fmt.Println("Error")
			}
		}

		w.WriteHeader(200)
	}
}

func (rt *RealtimeAnalyzer) startHTTPServer() {
	http.HandleFunc("/instagram", func(w http.ResponseWriter, r *http.Request) {
		rt.instagramHandler(w, r)
	})
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))
	http.Handle("/", http.HandlerFunc(HomeHandler))
	http.Handle("/sock", websocket.Handler(rt.WebSocketServer))

	err := http.ListenAndServe(":"+rt.config.Port, nil)
	rt.errChan <- err
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rt := new(RealtimeAnalyzer)
	rt.dictInstagram = make(map[string]bool)
	rt.strChan = make(chan string, 1000) // buffered channel with 1000 entries
	rt.errChan = make(chan error)        // unbuffered channel
	rt.activeClients = make(map[string]Client)
	rt.start = false

	// read configuration file
	err := rt.readConfig("config.xml")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// replace the IP Address with the HTML file
	err = rt.changeIPAddress("home.html", rt.config.IPAddress+":"+rt.config.Port)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	go rt.startHTTPServer()
	go rt.broadcastData()
	//go rt.generateGeoData()
	go rt.twitterStream()
	go rt.InstagramStream()

	err = <-rt.errChan

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
