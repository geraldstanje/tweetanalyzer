package main

import(
    "net/http"
    "log"
    "os"
    "fmt"
    "time"
    "strconv"
    "sync"
)

const resp = `<!DOCTYPE html>
<html>
<head>
<title>Marker Test</title>

<style>
html, body {
    height: 100%;
    margin: 0;
    padding: 0;
}

#map-canvas, #map_canvas {
    height: 100%;
}

#map-canvas, #map_canvas {
    height: 650px;
}
</style>

<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.3.2/jquery.min.js"></script>
<script type="text/javascript" src="https://maps.googleapis.com/maps/api/js?key=AIzaSyBy9hZSnE48aRbGD8zQAHRrtlOsk_H27BU&sensor=false"></script>

<script>
var map;
var markers = [];

$(document).ready(function () {
  setInterval("delayedPost()", 1000);
});

function initialize() {
    var NY = new google.maps.LatLng(40.784148,-73.966140);
    var mapOptions = {
        zoom: 13,
        center: NY//,
        //mapTypeId: google.maps.MapTypeId.TERRAIN
    }
    map = new google.maps.Map(document.getElementById('map-canvas'),mapOptions);

    //google.maps.event.addListener(map, 'click', function(event) {
    //            addMarker(event.latLng);
    //});

    //google.maps.event.addListener(map, 'rightclick', function(event) {
    //            marker.setMap(null);
    //});
}

function delayedPost() {
  //  window.alert(data);
  $.post("http://localhost:8080/gettime", "", function(data, status) {
    var location = data.split(",");
    var myLatlng = new google.maps.LatLng(parseFloat(location[0]), parseFloat(location[1]));
    addMarker(myLatlng);
    //$("#output").empty();
    //$("#output").append(data);
    //window.alert(data);
  });
}

function addMarker(location) {
        var marker = new google.maps.Marker({
        position: location,
        map: map
    });
    google.maps.event.addListener(marker, 'rightclick', function(event) {
        marker.setMap(null);
    });

    markers.push(marker);
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
  s []string 
}

func (s *Slice) myhandler(w http.ResponseWriter, r *http.Request) {
  var str string // "40.765498, -73.980732"

  s.mu.Lock()
  if len(s.s) > 0 {
      str = s.s[0]
      s.s = s.s[1:len(s.s)]
  }
  s.mu.Unlock()

  //println(x)
  fmt.Fprint(w, str)
}

// handler to cater AJAX requests
func handlerGetTime(w http.ResponseWriter, r *http.Request) {
    fmt.Println("handler called")

    fmt.Fprint(w, "40.765498, -73.980732")
}

func FloatToString(input_num float64) string {
    // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func (s *Slice) generatedata() {
    var longitude = 40.765498
    var latitude = -73.980732

    for {
      str := FloatToString(longitude) + ", " + FloatToString(latitude)

      s.mu.Lock()
      s.s = append(s.s, str)
      s.mu.Unlock()

      // increment
      longitude += 0.0001
      latitude += 0.0001

      time.Sleep(500 * time.Millisecond)
    }
}

func main() {
    s := new(Slice)

    go s.generatedata()

    http.HandleFunc("/", handler)
    http.HandleFunc("/gettime", func(w http.ResponseWriter, r *http.Request) {
      s.myhandler(w, r)
    })
    //http.HandleFunc("/gettime", handlerGetTime)
    err := http.ListenAndServe(":8080", nil)

    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
}