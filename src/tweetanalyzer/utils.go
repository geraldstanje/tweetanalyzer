package tweetanalyzer

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	TwitterID = iota
	InstagramID
	FlickrID
)

var points = [][]float64{{40.703286, -74.017739},
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

// to convert a float number to a string
func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func StringToFloat(str string) float64 {
	floatVal, _ := strconv.ParseFloat(str, 64)
	return floatVal
}

func IntToString(value int) string {
	return strconv.FormatInt(int64(value), 10)
}

func StringToInt(value string) int {
	result, _ := strconv.ParseInt(value, 10, 64)
	return int(result)
}

func Random(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func RandInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// reference: http://alienryderflex.com/polygon/
func IsPointInPolygon(longitude float64, latitude float64) bool {
	var i int
	var j int
	var odd_nodes bool
	var number_of_points int

	number_of_points = len(points)
	j = number_of_points - 1
	odd_nodes = false

	for i = 0; i < number_of_points; i++ {
		if points[i][1] < latitude && points[j][1] >= latitude || points[j][1] < latitude && points[i][1] >= latitude {
			if points[i][0]+(latitude-points[i][1])/(points[j][1]-points[i][1])*(points[j][0]-points[i][0]) < longitude {
				odd_nodes = !odd_nodes
			}
		}
		j = i
	}

	return odd_nodes
}

func MinElement(col int) float64 {
	min := math.MaxFloat64

	for i := 0; i < len(points); i++ {
		if points[i][col] < min {
			min = points[i][col]
		}
	}

	return min
}

func MaxElement(col int) float64 {
	max := math.SmallestNonzeroFloat64

	for i := 0; i < len(points); i++ {
		if points[i][col] > max {
			max = points[i][col]
		}
	}

	return max
}

func GenerateRandGeoLoc() (float64, float64) {
	var longitude float64
	var latitude float64

	for {
		longitude = Random(MinElement(0), MaxElement(0))
		latitude = Random(MinElement(1), MaxElement(1))

		retval := IsPointInPolygon(longitude, latitude)
		if retval {
			break
		}
	}

	return longitude, latitude
}

func CreateString(s string, num int) string {
	var str string

	for i := 0; i < num; i++ {
		str += s
	}

	return str
}
