package main

import(
    "net/http"
    "log"
    "os"
    "fmt"
    "time"
    "strconv"
    "sync"
    "math/rand"
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

$(document).ready(function () {
  setInterval("delayedPost()", 1000);
});

function initialize() {
  var NY = new google.maps.LatLng(40.784148,-73.966140);
  var mapOptions = {
    zoom: 12,
    center: NY//,
    //mapTypeId: google.maps.MapTypeId.TERRAIN
  }
  map = new google.maps.Map(document.getElementById('map-canvas'),mapOptions);
}

function delayedPost() {
  $.post("http://91.115.5.50:8080/getgeolocation", "", function(data, status) {
    var location = data.split(",");
    var myLatlng = new google.maps.LatLng(parseFloat(location[0]), parseFloat(location[1]));

    //drawSimppleMarker(myLatlng);

    //drawCustomMarker(myLatlng);

    drawCircle(myLatlng);
  });
}

function drawSimppleMarker(location) {
  var marker = new google.maps.Marker({
    position: location,
    map: map,
    title: 'Some location'
  });
}

function drawCustomMarker(location) {
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
        title: 'Some location'//,
        //zIndex: 0
    });
  }

  function drawCircle(location) {
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
  mu sync.Mutex
  geolocation []string 
}

// handler to cater AJAX requests
func (s *Slice) myhandler(w http.ResponseWriter, r *http.Request) {
  var str string // e.g. "40.765498, -73.980732"

  s.mu.Lock()
  if len(s.geolocation) > 0 {
    str = s.geolocation[0]
    s.geolocation = s.geolocation[1:len(s.geolocation)]
  }
  s.mu.Unlock()

  fmt.Fprint(w, str)
}

func FloatToString(input_num float64) string {
  // to convert a float number to a string
  return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func (s *Slice) generatedata() {
  // links:               right:    delta:   
  // latitude = -74.00,   -73.95    0.5
  // up:                  down:
  // longitude = 40.82,   40.70     0.12

  for {
    var longitude = (rand.Float64() * 0.12) + 40.70
    var latitude = (rand.Float64() * 0.05) - 74.00

    str := FloatToString(longitude) + ", " + FloatToString(latitude)

    s.mu.Lock()
    s.geolocation = append(s.geolocation, str)
    s.mu.Unlock()

    time.Sleep(500 * time.Millisecond)
  }
}

func main() {
  s := new(Slice)

  go s.generatedata()

  http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))
  http.HandleFunc("/", handler)
  http.HandleFunc("/getgeolocation", func(w http.ResponseWriter, r *http.Request) {
    s.myhandler(w, r)
  })

  err := http.ListenAndServe(":8080", nil)

  if err != nil {
    log.Println(err)
    os.Exit(1)
  }
}