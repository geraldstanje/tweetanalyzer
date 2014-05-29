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
    "math"
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
  var NY = new google.maps.LatLng(40.760837, -73.981847); // (40.784148,-73.966140);
  var mapOptions = {
    zoom: 13,
    center: NY//,
    //mapTypeId: google.maps.MapTypeId.TERRAIN
  }
  map = new google.maps.Map(document.getElementById('map-canvas'),mapOptions);

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
 
  //makePolygon(polygonCoords);
}

function delayedPost() {
  $.post("http://212.197.174.235:8080/getgeolocation", "", function(data, status) {
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

  /* 
* add markers to the given lat lng
*
*/
function makePolygon(polygonCoords) {
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
  mu sync.Mutex
  geolocation []string 
  points [][]float64
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

func floatToString(input_num float64) string {
  // to convert a float number to a string
  return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func random(min, max float64) float64 {
  return rand.Float64() * (max - min) + min
}

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
  for {
    longitude, latitude := s.generateRandGeoLoc()
    str := floatToString(longitude) + ", " + floatToString(latitude)

    s.mu.Lock()
    s.geolocation = append(s.geolocation, str)
    s.mu.Unlock()

    time.Sleep(500 * time.Millisecond)
  }
}

func main() {
  s := new(Slice)

  go s.generateGeoData()

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