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
    "github.com/darkhelmet/twitterstream"
)

const resp = `
<!DOCTYPE html>
<html>
  <head>
    <title>Realtime Map</title>
    <meta name="viewport" content="initial-scale=1.0, user-scalable=no">
    <meta charset="utf-8">
    <style>
      html, body, #map-canvas {
            height: 100%;
            margin: 0px;
            padding: 0px
    }
    </style>

    <script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.3.2/jquery.min.js"></script>
    <script type="text/javascript" src="https://maps.googleapis.com/maps/api/js?key=AIzaSyBy9hZSnE48aRbGD8zQAHRrtlOsk_H27BU&sensor=false"></script>
    <script>

var map;
var URL="http://194.96.77.18:8080/getlatlongtext";

$(document).ready(function () {
  setInterval("delayedPost()", 50);
});

function initialize() {
  var NY = new google.maps.LatLng(40.760837, -73.981847); // (40.784148,-73.966140);
  var mapOptions = {
    zoom: 13,
    center: NY//,
    //mapTypeId: google.maps.MapTypeId.TERRAIN
  }
  map = new google.maps.Map(document.getElementById('map-canvas'),mapOptions);
 
  //makePolygon();
}

function delayedPost() {
  $.post(URL, "", function(data, status) {
    if(data.length > 0) {
      var string_split = data.split(",");
      var myLatlng = new google.maps.LatLng(parseFloat(string_split[0]), parseFloat(string_split[1]));

      //drawSimppleMarker(myLatlng, string_split[2]);

      //drawCustomMarker(myLatlng, string_split[2]);

      drawCircle(myLatlng, string_split[2]);
    }
  });
}

function drawSimppleMarker(location, text) {
  var infowindow = new google.maps.InfoWindow({
    content: text,
    maxWidth: 200
  });

  var marker = new google.maps.Marker({
    position: location,
    map: map,
    title: 'Some location'
  });

  google.maps.event.addListener(marker, 'click', function() {
    infowindow.open(map,marker);
  });
}

function drawCustomMarker(location, text) {
  var infowindow = new google.maps.InfoWindow({
    content: text,
    maxWidth: 200
  });

  // Marker sizes are expressed as a Size of X,Y
  // where the origin of the image (0,0) is located
  // in the top left of the image.

  // Origins, anchor positions and coordinates of the marker
  // increase in the X direction to the right and in
  // the Y direction down.
  var image = {
    url: '/images/beachflag.png',
    // This marker is 20 pixels wide by 32 pixels tall.
    size: new google.maps.Size(20, 32),
    // The origin for this image is 0,0.
    origin: new google.maps.Point(0,0),
    // The anchor for this image is the base of the flagpole at 0,32.
    anchor: new google.maps.Point(0, 32)
  };
  // Shapes define the clickable region of the icon.
  // The type defines an HTML &lt;area&gt; element 'poly' which
  // traces out a polygon as a series of X,Y points. The final
  // coordinate closes the poly by connecting to the first
  // coordinate.
  var shape = {
    coords: [1, 1, 1, 20, 18, 20, 18 , 1],
    type: 'poly'
  };
  
  var marker = new google.maps.Marker({
    position: location,
    map: map,
    icon: image,
    shape: shape,
    title: 'Some location',
    zIndex: 0
  });

  google.maps.event.addListener(marker, 'click', function() {
    infowindow.open(map,marker);
  });
}

function HandleInfoWindow(latLng, content) {
  var infoWindow = new google.maps.InfoWindow({
    maxWidth: 420
  });

  infoWindow.setContent(content);
  infoWindow.setPosition(latLng);
  infoWindow.open(map);
}

function drawCircle(location, text) {
  // Add a Circle overlay to the map.
  var circle = new google.maps.Circle({
    center: location,
    radius: 100,
    strokeColor: "#FF0000",
    strokeOpacity: 0.8,
    strokeWeight: 2,
    fillColor: "#FF0000",
    fillOpacity: 0.35,
    map: map,
    title: 'Some location'
  });

  google.maps.event.addListener(circle, 'click', function() {
    HandleInfoWindow(location, text)
  });
} 

function makePolygon() {
  polygonCoords = [
    new google.maps.LatLng('40.703286','-74.017739'),
    new google.maps.LatLng('40.735551','-74.010487'),
    new google.maps.LatLng('40.752979','-74.007397'),
    new google.maps.LatLng('40.815891', '-73.960540'),
    new google.maps.LatLng('40.800966', '-73.929169'),
    new google.maps.LatLng('40.783921','-73.94145'),
    new google.maps.LatLng('40.776122','-73.941965'),
    new google.maps.LatLng('40.739974','-73.972864'),
    new google.maps.LatLng('40.729308','-73.971663'),
    new google.maps.LatLng('40.711614','-73.978014'),
    new google.maps.LatLng('40.706148','-74.00239'),
    new google.maps.LatLng('40.702114','-74.009671'),
    new google.maps.LatLng('40.701203','-74.015164')    
  ]

  var polygon = new google.maps.Polygon({
    paths: polygonCoords,
    strokeColor: "#FF0000",
    strokeOpacity: 0.8,
    strokeWeight: 2,
    fillColor: "#FF0000",
    fillOpacity: 0.35 
  });
 
  polygon.setMap(map); 
}

google.maps.event.addDomListener(window, 'load', initialize);

    </script>

  </head>
  <body>
    <div id="map-canvas"></div>
  </body>
</html>`

// handler for the main page
func handler(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte(resp))
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

  client := twitterstream.NewClient("xxx", 
                                    "xxx", 
                                    "xxx", 
                                    "xxx")
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
  http.HandleFunc("/", handler)
  http.HandleFunc("/getlatlongtext", func(w http.ResponseWriter, r *http.Request) {
    s.myhandler(w, r)
  })

  err := http.ListenAndServe(":8080", nil)

  if err != nil {
    log.Println(err)
    os.Exit(1)
  }
}