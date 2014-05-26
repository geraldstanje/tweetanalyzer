package main

import(
    "net/http"
    "log"
    "os"
    "fmt"
    "github.com/darkhelmet/twitterstream"
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
  setInterval("delayedPost()", 5000);
});

function initialize() {
    var NY = new google.maps.LatLng(40.784148,-73.966140);
    var mapOptions = {
        zoom: 13,
        center: NY//,
        //mapTypeId: google.maps.MapTypeId.TERRAIN
    }
    map = new google.maps.Map(document.getElementById('map-canvas'),mapOptions);

    google.maps.event.addListener(map, 'click', function(event) {
                addMarker(event.latLng);
    });

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

// handler to cater AJAX requests
func handlerGetTime(w http.ResponseWriter, r *http.Request) {
    fmt.Println("handler called")
    fmt.Fprint(w, "40.765498, -73.980732")
}

func decode(conn *twitterstream.Connection) {
    for {
        if tweet, err := conn.Next(); err == nil {

            //fmt.Println(tweet.Text)

            
            fmt.Println(tweet.User.ScreenName + " said: " + tweet.Text)
              //fmt.Println(tweet.Text)
            //}
        } else {
            fmt.Printf("Failed decoding tweet: %s", err)
            //continue
            return
        }
    }
}

func main() {
    var wait = 1
    var maxWait = 600 // Seconds

    client := twitterstream.NewClient("l76vc0wSlg9UBGx6Pt2KuEdkY", "0SUxkYDe4opkkoz1Hj72DNYRObQcmiAMHHE5VUjJRmwDk55RUs", "957672396-4rvqhNjhM9nncGDyxcjYXnoUvSYrenKFGMtTDMBZ", "Xp5c2fojBo2DlEm0ScXwtW9WbF2dYznvstEG75CrZs9fQ")
    
    for {
        conn, err := client.Locations(twitterstream.Point{40, -74}, twitterstream.Point{41, -73})

        if err != nil {
            log.Println(err)
            wait = wait << 1 // exponential backoff
            log.Printf("waiting for %d seconds before reconnect", min(wait, maxWait))
            time.Sleep(time.Duration(min(wait, maxWait)) * time.Second)
            continue
        } else {
          wait = 1
        }

        decode(conn)
    }
}

func main() {
    locations := []string{}

    http.HandleFunc("/", handler)
    http.HandleFunc("/gettime", handlerGetTime)
    err := http.ListenAndServe(":8080", nil)

    if err != nil {
        log.Println(err)
        os.Exit(1)
    }
}