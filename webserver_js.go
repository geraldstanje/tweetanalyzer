package main

import(
    "net/http"
    "log"
    "os"
    "fmt"
    "time"
    "strconv"
    "math/rand"
    "math"
    "runtime"
    "io/ioutil"
    "github.com/darkhelmet/twitterstream"
)

// handler for the main page
func HomeHandler(response http.ResponseWriter, request *http.Request) {
  response.Header().Set("Content-type", "text/html")
  webpage, err := ioutil.ReadFile("home_js.html")
  if err != nil { 
    http.Error(response, fmt.Sprintf("home_js.html file error %v", err), 500)
  }
  fmt.Fprint(response, string(webpage));
}

type Slice struct {
  channel chan string
  points [][]float64
}

// handler to cater AJAX requests
func (s *Slice) myhandler(w http.ResponseWriter, r *http.Request) {
  select {
  case str := <-s.channel:
    fmt.Fprint(w, str)
  default: return
  }
}

func floatToString(input_num float64) string {
  // to convert a float number to a string
  return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func random(min, max float64) float64 {
  return rand.Float64() * (max - min) + min
}

// reference: http://alienryderflex.com/polygon/
func (s *Slice) isPointInPolygon(longitude float64, latitude float64) bool {
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

func (s *Slice) minElement(col int) float64 {
  min := math.MaxFloat64

  for i := 0; i < len(s.points); i++ {
    if s.points[i][col] < min {
      min = s.points[i][col]
    }
  }

  return min
}

func (s *Slice) maxElement(col int) float64 {
  max := math.SmallestNonzeroFloat64

  for i := 0; i < len(s.points); i++ {
    if s.points[i][col] > max {
      max = s.points[i][col]
    }
  }

  return max
}

func (s *Slice) generateRandGeoLoc() (float64, float64) {
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

func (s *Slice) generateGeoData() {
  tweet_count := 0

  for {
    longitude, latitude := s.generateRandGeoLoc()

    str := floatToString(longitude) + ", " + floatToString(latitude) + ", " + "tweet"
    s.channel <- str

    tweet_count += 1
    fmt.Println(tweet_count)

    time.Sleep(10 * time.Millisecond)
  }
}

func (s *Slice) min(a, b int) int {
  if a < b {
  return a
  }

  return b
}

func (s *Slice) decode(conn *twitterstream.Connection) {
  tweet_count := 0

  for {
    if tweet, err := conn.Next(); err == nil {
      var str string
      coord := tweet.Coordinates

      if coord != nil {
        str = floatToString(float64(coord.Lat)) + ", " + floatToString(float64(coord.Long)) + ", " + "@" + tweet.User.ScreenName + ": " + tweet.Text

        s.channel <- str

        tweet_count += 1
        fmt.Println(tweet_count)
      }
    } else {
      fmt.Printf("Failed decoding tweet: %s", err)
      return
    }
  }
}

func (s *Slice) twitterStream() {
  var wait = 1
  var maxWait = 600 // Seconds

  client := twitterstream.NewClient("l76vc0wSlg9UBGx6Pt2KuEdkY", 
                                    "0SUxkYDe4opkkoz1Hj72DNYRObQcmiAMHHE5VUjJRmwDk55RUs", 
                                    "957672396-4rvqhNjhM9nncGDyxcjYXnoUvSYrenKFGMtTDMBZ", 
                                    "Xp5c2fojBo2DlEm0ScXwtW9WbF2dYznvstEG75CrZs9fQ")
  client.Timeout = 0

  for {
      // latitude/longitude of NY
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

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())

  s := new(Slice)
  s.channel = make(chan string, 1000) // buffered channel with 1000 entries 

  //go s.twitterStream()
  go s.generateGeoData()

  http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))
  http.Handle("/", http.HandlerFunc(HomeHandler))
  http.HandleFunc("/getlatlongtext", func(w http.ResponseWriter, r *http.Request) {
    s.myhandler(w, r)
  })

  err := http.ListenAndServe(":8080", nil)

  if err != nil {
    log.Println(err)
    os.Exit(1)
  }
}