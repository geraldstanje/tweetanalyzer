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
  instagramClient *instagram.Client
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
        return fmt.Errorf("Error: invalid len of file")
    }

    b, err := ioutil.ReadFile(filename)
    if err != nil { 
      return err 
    }

    s := string(b)

    reg, err := regexp.Compile(`[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}:[0-9]{2,5}/sock";`)
    if err != nil {
      return err
    }

    ips := reg.FindAllString(s, -1)

    for _, ip := range ips {
      newStr += "/sock\";"
      s = strings.Replace(s, ip, newStr, -1)
    }

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

// to convert a float number to a string
func floatToString(input_num float64) string {
  return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func random(min, max float64) float64 {
  return rand.Float64() * (max - min) + min
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
  j = number_of_points-1
  odd_nodes = false

  for i = 0; i < number_of_points; i++ {
    if (rt.points[i][1]<latitude && rt.points[j][1]>=latitude || rt.points[j][1]<latitude && rt.points[i][1]>=latitude) {
      if (rt.points[i][0]+(latitude-rt.points[i][1])/(rt.points[j][1]-rt.points[i][1])*(rt.points[j][0]-rt.points[i][0])<longitude) {
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
    if retval == true {
      break
    }
  }

  return longitude, latitude
}

func (rt *RealtimeAnalyzer) generateGeoData() {
  for {
    longitude, latitude := rt.generateRandGeoLoc()

    str := floatToString(longitude) + ", " + 
           floatToString(latitude) + ", " + 
           "tweet" + ", " +
           "0"

    rt.channel <- str

    time.Sleep(50 * time.Millisecond)
  }
}

func (rt *RealtimeAnalyzer) min(a, b int) int {
  if a < b {
    return a
  }

  return b
}

func (rt *RealtimeAnalyzer) broadcastData() {
  var Message  = websocket.Message 
  var err error
  tweet_count := 0

  for {
    select {
      case str := <-rt.channel:
        for ip, _ := range rt.ActiveClients {
          if err = Message.Send(rt.ActiveClients[ip].websocket, str); err != nil {
            // we could not send the message to a peer
            log.Println("Could not send message to ", ip, err.Error())

            // work-around: https://code.google.com/p/go/issues/detail?id=3117
            var tmp = rt.ActiveClients[ip]
            tmp.errorCount += 1 
            rt.ActiveClients[ip] = tmp 

            if rt.ActiveClients[ip].errorCount >= errorCounterMax {
              log.Println("Client disconnected:", ip)
              delete(rt.ActiveClients, ip)
            }
          }
        }

      default: continue
  
      tweet_count += 1
      fmt.Println(tweet_count)
    }
  }
}

func (rt *RealtimeAnalyzer) formatTwitterData(tweet *twitterstream.Tweet) (string, error) {
  var comment string

  if tweet.Coordinates == nil {
    return "", fmt.Errorf("Tweet without geolocation set")
  }

  reg, err := regexp.Compile(`https?://t\.co/[a-zA-Z0-9]{0,10}`)
  if err != nil {
    return "", fmt.Errorf("Regex definition failed: %s", err)
  }

  reg2, err := regexp.Compile(`pic.twitter.com/[a-zA-Z0-9]{0,10}`)
  if err != nil {
    return "", fmt.Errorf("Regex definition failed: %s", err)
  }

  comment = tweet.Text

  link1 := reg.FindAllString(comment, -1)
  if link1 != nil {
    comment = strings.Replace(comment, link1[0], "<br><a href=\"" + link1[0] + "\">" + link1[0] + "</a>", -1)
  }
        
  if link1 == nil {
    link2 := reg2.FindAllString(comment, -1)

    if link2 != nil {
      comment = strings.Replace(comment, link2[0], "<br><a href=\"" + link2[0] + "\">" + link2[0] + "</a>", -1)
    }
  }

  comment = floatToString(float64(tweet.Coordinates.Lat)) + ", " + 
            floatToString(float64(tweet.Coordinates.Long)) + ", " + 
            tweet.User.ScreenName + ": " + 
            "<br>" + comment + ", " +
            "0"
  return comment, nil
}

func (rt *RealtimeAnalyzer) decode(conn *twitterstream.Connection) {
  for {
    if tweet, err := conn.Next(); err == nil {
      comment, err := rt.formatTwitterData(tweet)

      if err == nil {
        rt.channel <- comment
      }
    } else {
      fmt.Printf("Failed decoding tweet: %s", err)
      return
    }
  }
}

func (rt *RealtimeAnalyzer) twitterStream(config Config) {
  var wait = 1
  var maxWait = 600 // Seconds

  client := twitterstream.NewClient(config.TwitterConfig.ConsumerKey, config.TwitterConfig.ConsumerSecret, config.TwitterConfig.AccessToken, config.TwitterConfig.AccessSecret)
  client.Timeout = 0

  for {
      // latitude/longitude of New York City
      conn, err := client.Locations(twitterstream.Point{40, -74}, twitterstream.Point{41, -73})

      if err != nil {
        wait = wait << 1 // exponential backoff
        log.Printf(err.Error())
        log.Printf("waiting for %d seconds before reconnect", rt.min(wait, maxWait))
        time.Sleep(time.Duration(rt.min(wait, maxWait)) * time.Second)
        continue
      } else {
        wait = 1
      }

      rt.decode(conn)
  }
}

// reference: https://github.com/Niessy/websocket-golang-chat
// WebSocket server to handle chat between clients
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
  rt.ActiveClients[client] = Client{0, ws, client}

  // for loop so the websocket stays open otherwise it'll close
  for {
    time.Sleep(1 * time.Second) // TODO: do we need a delay?
  }
}

// http://instagram.com/developer/clients/manage/?edited=RealtimeDataAnalysis
func (rt *RealtimeAnalyzer) InstagramStream(conf Config) {
  time.Sleep(1 * time.Second) // TODO: do we need a delay?
 
  rt.instagramClient = instagram.NewClient(nil)
  rt.instagramClient.ClientID = conf.InstagramConfig.ClientID
  rt.instagramClient.ClientSecret = conf.InstagramConfig.ClientSecret
  rt.instagramClient.AccessToken = conf.InstagramConfig.AccessToken

  res, err := rt.instagramClient.Realtime.DeleteAllSubscriptions()
  if err != nil {
    fmt.Println("client.Realtime.DeleteAllSubscriptions returned error: ", err)
    return
  }

  time.Sleep(1 * time.Second) // TODO: do we need a delay?

  // check radius with: http://www.freemaptools.com/radius-around-point.htm
  // subscribe to Manhattan uptown area
  res, err = rt.instagramClient.Realtime.SubscribeToGeography("40.790716", "-73.955841", "5000", "http://" + conf.IPAddress + ":" + conf.Port + conf.InstagramConfig.CallbackURL)
  if err != nil {
    fmt.Println("client.Realtime.SubscribeToGeography returned error: ", err)
    return
  }

  rt.subscriptionIdUptown = res.ObjectID

  // subscribe to Manhattan downtown area
  res, err = rt.instagramClient.Realtime.SubscribeToGeography("40.711446", "-74.007968", "5000", "http://" + conf.IPAddress + ":" + conf.Port + conf.InstagramConfig.CallbackURL)
  if err != nil {
    fmt.Println("client.Realtime.SubscribeToGeography returned error: ", err)
    return
  }

  rt.subscriptionIdDowntown = res.ObjectID

  time.Sleep(5 * time.Second) // TODO: do we need a delay?
  rt.start = true // TODO: do we need to syncronize start?
}

func (rt *RealtimeAnalyzer) formatInstagramData(media instagram.Media) string {
  var comment string

  comment = floatToString(media.Location.Latitude) + ", " + 
            floatToString(media.Location.Longitude) + ", " +
            media.User.Username + ": "

  if media.Caption != nil {
    comment += "<br>" + media.Caption.Text + " " + 
               "<br><a href=\"" + media.Link + "\">" + media.Link + "</a>"
  }

  comment +=  ", " + "1"

  return comment
}

func (rt *RealtimeAnalyzer) getRecentMediaUptown(Time int64) {
  opt := &instagram.Parameters {
    Lat: 40.790716,
    Lng: -73.955841,
    Distance: 5000,
  }

  rt.mu.Lock() // TODO: we need the lock?
  media, _, err := rt.instagramClient.Media.Search(opt)
  rt.mu.Unlock()

  if err != nil {
      log.Println("Error: " + err.Error())
      return
  }

  if len(media) > 0 {
    comment := rt.formatInstagramData(media[0])
    rt.channel <- comment
  }
}

func (rt *RealtimeAnalyzer) getRecentMediaDowntown(Time int64) {
  opt := &instagram.Parameters {
    Lat: 40.711446,
    Lng: -74.007968,
    Distance: 5000,
  }

  rt.mu.Lock()
  media, _, err := rt.instagramClient.Media.Search(opt)
  rt.mu.Unlock()

  if err != nil {
      log.Println("Error: " + err.Error())
      return
  }

  if len(media) > 0 {
    comment := rt.formatInstagramData(media[0])
    rt.channel <- comment
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
  } else if r.Method == "POST" {
    defer r.Body.Close() // TODO: is ok?

    var m = []instagram.RealtimeResponse{}
    err := json.NewDecoder(r.Body).Decode(&m)
    if err != nil {
        log.Println("Error: " + err.Error())
        return
    }
  
    if rt.start == true && len(m) > 0 {
      if rt.subscriptionIdUptown == m[0].ObjectID {
        go rt.getRecentMediaUptown(m[0].Time)
      } else if rt.subscriptionIdDowntown == m[0].ObjectID {
        go rt.getRecentMediaDowntown(m[0].Time)
      } else {
        fmt.Println(m[0].ObjectID)
        fmt.Println(m[0].SubscriptionID)
        fmt.Println("Error")
      }
    }

    w.WriteHeader(200)
    w.Write([]byte("Thanks\n")) // TODO: is ok?
  }
  // TODO: do we need to reply in the else case?
}

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())

  // read configuration file
  config, err := readConfig("config.xml")
  if err != nil {
    log.Println(err)
    os.Exit(1)
  }

  // replace the IP Address with the HTML file
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
  go s.twitterStream(config)
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