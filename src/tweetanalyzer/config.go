package tweetanalyzer

import (
	"bufio"
	"encoding/xml"
	"os"
)

type Config struct {
	IPAddress       string          `xml:"ip"`
	Port            string          `xml:"port"`
	TwitterConfig   TwitterConfig   `xml:"twitter"`
	InstagramConfig InstagramConfig `xml:"instagram"`
	FlickrConfig    FlickrConfig    `xml:"flickr"`
}

type TwitterConfig struct {
	ConsumerKey    string        `xml:"consumerKey"`
	ConsumerSecret string        `xml:"consumerSecret"`
	AccessToken    string        `xml:"accessToken"`
	AccessSecret   string        `xml:"accessSecret"`
	Location       []GeoLocation `xml:"location"`
}

type InstagramConfig struct {
	Username     string        `xml:"username"`
	Password     string        `xml:"password"`
	ClientID     string        `xml:"clientID"`
	ClientSecret string        `xml:"clientSecret"`
	AccessToken  string        `xml:"accessToken"`
	CallbackURL  string        `xml:"callbackURL"`
	Location     []GeoLocation `xml:"location"`
}

type FlickrConfig struct {
	ApiKey      string      `xml:"apiKey"`
	Secret      string      `xml:"secret"`
	Radius      string      `xml:"radius"`
	RadiusUnits string      `xml:"radiusUnits"`
	Location    GeoLocation `xml:"location"`
}

type GeoLocation struct {
	Lat  float64 `xml:"lat,attr"`
	Long float64 `xml:"long,attr"`
}

func parseXML(reader *bufio.Reader) (Config, error) {
	config := Config{}
	err := xml.NewDecoder(reader).Decode(&config)
	return config, err
}

func ReadConfig(filename string) (Config, error) {
	xmlFile, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer xmlFile.Close()

	reader := bufio.NewReader(xmlFile)
	config, err := parseXML(reader)

	return config, err
}
