package main

import(
    "net/http"
    "log"
    "os"
    "bufio"
    "fmt"
    "time"
    "strconv"
    "math/rand"
    "math"
    "runtime"
    "strings"
    "io/ioutil"
    "regexp"
    "encoding/json"
    "sync"
    "encoding/xml"
    "github.com/darkhelmet/twitterstream"
    "code.google.com/p/go.net/websocket"
    "github.com/carbocation/go-instagram/instagram"
)

const errorCounterMax = 3

// Client connection consists of the websocket and the client ip
type Client struct {
  errorCount int
  websocket *websocket.Conn
  clientIP string
}

type RealtimeAnalyzer struct {
  mu sync.Mutex
  start bool
  ActiveClients map[string]Client
  instagramClient* instagram.Client
  subscriptionIdUptown string
  subscriptionIdDowntown string
  channel chan string
  points [][]float64
}

type Config struct {
  IPAddress string                `xml:"ip"`
  Port string                     `xml:"port"`
  TwitterConfig TwitterConfig     `xml:"twitter"`
  InstagramConfig InstagramConfig `xml:"instagram"`
}

type TwitterConfig struct {
  ConsumerKey string    `xml:"consumerKey"`
  ConsumerSecret string `xml:"consumerSecret"`
  AccessToken string    `xml:"accessToken"`
  AccessSecret string   `xml:"accessSecret"`
}

type InstagramConfig struct {
  ClientID string     `xml:"clientID"`
  ClientSecret string `xml:"clientSecret"`
  AccessToken string  `xml:"accessToken"`
  CallbackURL string  `xml:"callbackURL"`
}

func parseXML(data []byte) (Config, error) {
    config := Config{}
    err := xml.Unmarshal(data, &config)
    if err != nil {
        return Config{}, err
    }

    return config, nil
}

func readConfig(filename string) (Config, error) {
  xmlFile, err := os.Open(filename)
  if err != nil {
    return Config{}, err
  }
  defer xmlFile.Close()
  
  reader := bufio.NewReader(xmlFile)
  contents, _ := ioutil.ReadAll(reader)
  config, _ := parseXML(contents)

  return config, nil
}

func changeIPAddress(filename string, newStr string) (error) {
    if len(filename) == 0 {
        return nil
    }

    b, err := ioutil.ReadFile(filename)
    if err != nil { 
      return err 
    }

    s := string(b)

    reg, err := regexp.Compile(`[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}:[0-9]{2,5}/sock";`)
    if err != nil {
      log.Fatal(err)
    }

    ips := reg.FindAllString(s, -1)

    newStr += "/sock\";"
    s = strings.Replace(s, ips[0], newStr, -1)

    err = ioutil.WriteFile(filename, []byte(s), 0644)
    if err != nil { 
      return err
    }

    return nil
}

// handler for the main page
func HomeHandler(response http.ResponseWriter, request *http.Request) {
  response.Header().Set("Content-type", "text/html")
  webpage, err := ioutil.ReadFile("home_websocket.html")
  if err != nil { 
    http.Error(response, fmt.Sprintf("home_websocket.html file error %v", err), 500)
  }
  fmt.Fprint(response, string(webpage));
}

func floatToString(input_num float64) string {
  // to convert a float number to a string
  return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func random(min, max float64) float64 {
  return rand.Float64() * (max - min) + min
}

// reference: http://alienryderflex.com/polygon/
func (s *RealtimeAnalyzer) isPointInPolygon(longitude float64, latitude float64) bool {
  s.points = [][]float64{{40.703286, -74.017739}, 
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

  number_of_points = len(s.points)
  j = number_of_points-1
  odd_nodes = false

  for i = 0; i < number_of_points; i++ {
    if (s.points[i][1]<latitude && s.points[j][1]>=latitude || s.points[j][1]<latitude && s.points[i][1]>=latitude) {
      if (s.points[i][0]+(latitude-s.points[i][1])/(s.points[j][1]-s.points[i][1])*(s.points[j][0]-s.points[i][0])<longitude) {
        odd_nodes = !odd_nodes 
      }
    }
    j = i 
  }

  return odd_nodes
}

func (s *RealtimeAnalyzer) minElement(col int) float64 {
  min := math.MaxFloat64

  for i := 0; i < len(s.points); i++ {
    if s.points[i][col] < min {
      min = s.points[i][col]
    }
  }

  return min
}

func (s *RealtimeAnalyzer) maxElement(col int) float64 {
  max := math.SmallestNonzeroFloat64

  for i := 0; i < len(s.points); i++ {
    if s.points[i][col] > max {
      max = s.points[i][col]
    }
  }

  return max
}

func (s *RealtimeAnalyzer) generateRandGeoLoc() (float64, float64) {
  var longitude float64
  var latitude float64

  for {
    longitude = random(s.minElement(0), s.maxElement(0))
    latitude = random(s.minElement(1), s.maxElement(1))

    retval := s.isPointInPolygon(longitude, latitude)
    if retval == true {
      break
    }
  }

  return longitude, latitude
}

func (s *RealtimeAnalyzer) generateGeoData() {

  for {
    longitude, latitude := s.generateRandGeoLoc()

    str := floatToString(longitude) + ", " + floatToString(latitude) + ", " + "tweet" + ", 0"
    s.channel <- str

    time.Sleep(50 * time.Millisecond)
  }
}

func (s *RealtimeAnalyzer) min(a, b int) int {
  if a < b {
  return a
  }

  return b
}

func (s *RealtimeAnalyzer) broadcastData() {
  var Message  = websocket.Message 
  var err error
  tweet_count := 0

  for {
    select {
      case str := <-s.channel:
        for ip, _ := range s.ActiveClients {
          if err = Message.Send(s.ActiveClients[ip].websocket, str); err != nil {
            // we could not send the message to a peer
            log.Println("Could not send message to ", ip, err.Error())

            // work-around: https://code.google.com/p/go/issues/detail?id=3117
            var tmp = s.ActiveClients[ip]
            tmp.errorCount += 1 
            s.ActiveClients[ip] = tmp 

            if s.ActiveClients[ip].errorCount >= errorCounterMax {
              log.Println("Client disconnected:", ip)
              delete(s.ActiveClients, ip)
            }
          }
        }

      default: continue
  
      tweet_count += 1
      fmt.Println(tweet_count)
    }
  }
}

func (s *RealtimeAnalyzer) decode(conn *twitterstream.Connection) {
  for {
    if tweet, err := conn.Next(); err == nil {
      var str string
      coord := tweet.Coordinates

      if coord != nil {
        str = floatToString(float64(coord.Lat)) + ", " + floatToString(float64(coord.Long)) + ", " + tweet.User.ScreenName + ": " + tweet.Text + ", 0"

        s.channel <- str
      }
    } else {
      fmt.Printf("Failed decoding tweet: %s", err)
      return
    }
  }
}

func (s *RealtimeAnalyzer) twitterStream(config TwitterConfig) {
  var wait = 1
  var maxWait = 600 // Seconds

  client := twitterstream.NewClient(config.ConsumerKey, config.ConsumerSecret, config.AccessToken, config.AccessSecret)
  client.Timeout = 0

  for {
      // latitude/longitude of New York City
      conn, err := client.Locations(twitterstream.Point{40, -74}, twitterstream.Point{41, -73})

      if err != nil {
        log.Println(err)
        wait = wait << 1 // exponential backoff
        log.Printf("waiting for %d seconds before reconnect", s.min(wait, maxWait))
        time.Sleep(time.Duration(s.min(wait, maxWait)) * time.Second)
        continue
      } else {
        wait = 1
      }

      s.decode(conn)
  }
}

// reference: https://github.com/Niessy/websocket-golang-chat
// WebSocket server to handle chat between clients
func (s *RealtimeAnalyzer) WebSocketServer(ws *websocket.Conn) {
  var err error

  // cleanup on server side
  defer func() {
    if err = ws.Close(); err != nil {
      log.Println("Websocket could not be closed", err.Error())
    }
  }()

  client := ws.Request().RemoteAddr
  log.Println("New client connected:", client)
  s.ActiveClients[client] = Client{0, ws, client}

  // for loop so the websocket stays open otherwise it'll close
  for {
    time.Sleep(1 * time.Second)
  }
}

// http://instagram.com/developer/clients/manage/?edited=RealtimeDataAnalysis
func (s *RealtimeAnalyzer) InstagramStream(conf Config) {
  time.Sleep(1 * time.Second)
 
  s.instagramClient = instagram.NewClient(nil)
  s.instagramClient.ClientID = conf.InstagramConfig.ClientID
  s.instagramClient.ClientSecret = conf.InstagramConfig.ClientSecret
  s.instagramClient.AccessToken = conf.InstagramConfig.AccessToken

  res, err := s.instagramClient.Realtime.DeleteAllSubscriptions()
  if err != nil {
    fmt.Println("client.Realtime.DeleteAllSubscriptions returned error: ", err)
    return
  }

  time.Sleep(1 * time.Second)

  // subscribe to Manhattan uptown area
  res, err = s.instagramClient.Realtime.SubscribeToGeography("40.790716", "-73.955841", "5000", "http://" + conf.IPAddress + ":" + conf.Port + conf.InstagramConfig.CallbackURL)
  if err != nil {
    fmt.Println("client.Realtime.SubscribeToGeography returned error: ", err)
    return
  }

  s.subscriptionIdUptown = res.ObjectID

  // subscribe to Manhattan downtown area
  res, err = s.instagramClient.Realtime.SubscribeToGeography("40.711446", "-74.007968", "5000", "http://" + conf.IPAddress + ":" + conf.Port + conf.InstagramConfig.CallbackURL)
  if err != nil {
    fmt.Println("client.Realtime.SubscribeToGeography returned error: ", err)
    return
  }

  s.subscriptionIdDowntown = res.ObjectID


  time.Sleep(5 * time.Second)
  s.start = true
}

func (s *RealtimeAnalyzer) getRecentMediaUptown(Time int64) {
  opt := &instagram.Parameters {
    Lat: 40.790716,
    Lng: -73.955841,
    Distance: 5000,
  }

  s.mu.Lock()
  media, _, err := s.instagramClient.Media.Search(opt)
  s.mu.Unlock()

  if err != nil {
      log.Println("Error: " + err.Error())
      return
  }

  if len(media) > 0 {
    var str string

    str = media[0].User.Username
    if media[0].Caption != nil {
      str = str + ": " + media[0].Caption.Text + " " + media[0].Link
    }

    str = floatToString(media[0].Location.Latitude) + ", " + floatToString(media[0].Location.Longitude) + ", " + str + ", 1"
    s.channel <- str
  }
}

func (s *RealtimeAnalyzer) getRecentMediaDowntown(Time int64) {
  opt := &instagram.Parameters {
    Lat: 40.711446,
    Lng: -74.007968,
    Distance: 5000,
  }

  s.mu.Lock()
  media, _, err := s.instagramClient.Media.Search(opt)
  s.mu.Unlock()

  if err != nil {
      log.Println("Error: " + err.Error())
      return
  }

  if len(media) > 0 {
    var str string

    str = media[0].User.Username
    if media[0].Caption != nil {
      str = str + ": " + media[0].Caption.Text
    }

    str = floatToString(media[0].Location.Latitude) + ", " + floatToString(media[0].Location.Longitude) + ", " + str + ", 1"
    s.channel <- str
  }
}

// http://instagram.com/developer/clients/manage/?edited=RealtimeDataAnalysis
func (s *RealtimeAnalyzer) instagramHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "GET" && r.FormValue("hub.mode") == "subscribe" && r.FormValue("hub.challenge") != "" {
    r.ParseForm()
    hub_challenge := r.FormValue("hub.challenge")
    fmt.Fprintf(w, hub_challenge)
  } else {
    var m = []instagram.RealtimeResponse{}
    err := json.NewDecoder(r.Body).Decode(&m)
    if err != nil {
        log.Println("Error: " + err.Error())
        return
    }
  
    if s.start == true && len(m) > 0 {
      if s.subscriptionIdUptown == m[0].ObjectID {
        go s.getRecentMediaUptown(m[0].Time)
      } else if s.subscriptionIdDowntown == m[0].ObjectID {
        go s.getRecentMediaDowntown(m[0].Time)
      } else {
        fmt.Println(m[0].ObjectID)
        fmt.Println(m[0].SubscriptionID)
        fmt.Println("Error")
      }
    }

    w.WriteHeader(200)
    w.Write([]byte("Thanks\n"))
  }
}

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())

  // read configuration file
  config, err := readConfig("config.xml")
  if err != nil {
    log.Println(err)
    os.Exit(1)
  }
  // replace the IP in the HTML file
  err = changeIPAddress("home_websocket.html", config.IPAddress + ":" + config.Port)
  if err != nil {
    log.Println(err)
    os.Exit(1)
  }

  s := new(RealtimeAnalyzer)
  s.channel = make(chan string, 1000) // buffered channel with 1000 entries 
  s.ActiveClients = make(map[string]Client)
  s.start = false

  go s.InstagramStream(config)
  //go s.generateGeoData()
  go s.twitterStream(config.TwitterConfig)
  go s.broadcastData()

  http.HandleFunc("/instagram", func(w http.ResponseWriter, r *http.Request) {
    s.instagramHandler(w, r)
  })
  http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))
  http.Handle("/", http.HandlerFunc(HomeHandler))
  http.Handle("/sock", websocket.Handler(s.WebSocketServer))

  err = http.ListenAndServe(":" + config.Port, nil)

  if err != nil {
    log.Println(err)
    os.Exit(1)
  }
}