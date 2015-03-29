package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"time"
	"tweetanalyzer"
)

const debug = false
const errorCounterMax = 3

// Client connection consists of the websocket and the client ip
type Client struct {
	errorCount int
	websocket  *websocket.Conn
	clientIP   string
}

type RealtimeAnalyzer struct {
	twitterstream   *tweetanalyzer.TwitterStream
	instagramstream *tweetanalyzer.InstagramStream
	flickrstream    *tweetanalyzer.FlickrStream
	config          tweetanalyzer.Config
	activeClients   map[string]Client
	strChan         chan string
	errChan         chan error
}

func (rt *RealtimeAnalyzer) changeIPAddressInFile(filename string, newStr string) error {
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

func (rt *RealtimeAnalyzer) generateGeoData() {
	for {
		longitude, latitude := tweetanalyzer.GenerateRandGeoLoc()

		str := tweetanalyzer.IntToString(tweetanalyzer.RandInt(0, 3)) + ", " +
			tweetanalyzer.FloatToString(longitude) + ", " +
			tweetanalyzer.FloatToString(latitude) + ", " +
			"tweet"

		rt.strChan <- str

		time.Sleep(200 * time.Millisecond)
	}
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

			tweet_count += 1
			fmt.Println(tweet_count)
		}
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

	// wait for errChan, so the websocket stays open otherwise it'll close
	err = <-rt.errChan
}

func (rt *RealtimeAnalyzer) startHTTPServer() {
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))
	http.Handle("/", http.HandlerFunc(HomeHandler))
	http.Handle("/sock", websocket.Handler(rt.WebSocketServer))

	err := http.ListenAndServe(":"+rt.config.Port, nil)
	rt.errChan <- err
}

func (rt *RealtimeAnalyzer) getExternalIP() string {
	resp, _ := http.Get("http://myexternalip.com/raw")
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)
	ip := string(contents)
	return ip[:len(ip)-1]
}

func NewRealtimeAnalyzer() *RealtimeAnalyzer {
	rt := RealtimeAnalyzer{}
	rt.strChan = make(chan string, 1000) // buffered channel with 1000 entries
	rt.errChan = make(chan error)        // unbuffered channel
	rt.activeClients = make(map[string]Client)
	return &rt
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	rt := NewRealtimeAnalyzer()

	// read configuration file
	var err error
	rt.config, err = tweetanalyzer.ReadConfig("config.xml")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// get the exernal IP address
	rt.config.IPAddress = rt.getExternalIP()

	// create TwitterStream, InstagramStream
	rt.twitterstream = tweetanalyzer.NewTwitterStream(rt.strChan, rt.errChan, rt.config)
	rt.instagramstream = tweetanalyzer.NewInstagramStream(rt.strChan, rt.errChan, rt.config)
	rt.instagramstream.SetRedirectIP(rt.config.IPAddress)
	rt.instagramstream.Create()
	rt.flickrstream = tweetanalyzer.NewFlickrStream(rt.strChan, rt.errChan, rt.config)
	rt.flickrstream.Create()

	// replace the IP Address within the HTML file
	err = rt.changeIPAddressInFile("home.html", rt.config.IPAddress+":"+rt.config.Port)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	go rt.startHTTPServer()
	go rt.broadcastData()
	//go rt.generateGeoData()
	go rt.instagramstream.InstagramStream()
	go rt.twitterstream.TwitterStream()
	go rt.flickrstream.FlickrStream()

	err = <-rt.errChan

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
