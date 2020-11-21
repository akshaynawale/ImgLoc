package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/rwcarlsen/goexif/exif"
	"googlemaps.github.io/maps"
)

// base URL for google maps search
var gMapURL string = "https://www.google.com/maps/search/?api=1"
var dPath string // stores Dir path

// getAllImgNames takes DirPath and returns all jpg files from it
func getAllImgNames(fPath string) []os.FileInfo {
	files, err := ioutil.ReadDir(fPath)
	if err != nil {
		glog.Errorf("failed to read folder: %s Error: %v", fPath, err)
		os.Exit(1)
	}

	var imgs []os.FileInfo
	for _, f := range files {
		// if Dir ignore
		if f.IsDir() {
			continue
		}
		// if jpeg then only consider
		m, err := regexp.MatchString(`.*\.(jpg)$`, strings.ToLower(f.Name()))
		if err != nil {
			glog.Errorf("failed to match aginst regexp for file: %s Error: %v", f.Name(), err)
			continue
		}
		if m {
			imgs = append(imgs, f)
		}
	}
	return imgs
}

// getImageLocations takes files as input and returns Map of file names and there respective google map urls
func getImageLocations(imgs []os.FileInfo) map[string]string {
	gMapURLs := map[string]string{}
	for _, f := range imgs {
		if glog.V(2) {
			glog.Infof("decoding exif data for %s", f.Name())
		}
		imgf, err := os.Open(path.Join(dPath, f.Name()))
		defer imgf.Close()
		if err != nil {
			e := fmt.Sprintf("failed to read file: %s", f.Name())
			glog.Error(e)
			gMapURLs[f.Name()] = e
			continue
		}

		e, err := exif.Decode(imgf)
		if err != nil {
			e := fmt.Sprintf("failed to decode exif data from file: %s, Error: %v", f.Name(), err)
			glog.Error(e)
			gMapURLs[f.Name()] = e
			continue
		}
		// get location information from the exif data
		var loc maps.LatLng
		loc.Lat, loc.Lng, err = e.LatLong()
		if err != nil {
			glog.Errorf("failed to get latitude longitude data for file: %s Error: %v", f.Name(), err)
		}
		if loc.Lat == 0 && loc.Lng == 0 {
			gMapURLs[f.Name()] = ""
			continue
		}
		gMapURLs[f.Name()] = fmt.Sprintf("%s", getGoogleMapURL(loc))
	}
	return gMapURLs

}

// getGoogleMapURL takes longitude and latitude info and returns google map url
func getGoogleMapURL(loc maps.LatLng) *url.URL {
	// Create google maps URL
	u, err := url.Parse(gMapURL)
	if err != nil {
		glog.Errorf("failed to parse url somthing is wrong: %v", err)
		os.Exit(1)
	}
	v := u.Query()
	v.Add("query", fmt.Sprintf("%f,%f", loc.Lat, loc.Lng))
	u.RawQuery = v.Encode()
	return u
}

func main() {

	var fmtJSON bool
	flag.StringVar(&dPath, "path", ".", "filepath to the folder for photos")
	flag.BoolVar(&fmtJSON, "json", false, "output in json format")
	flag.Parse()

	if glog.V(2) {
		glog.Infof("looking for photos at path: %s", dPath)
	}
	imgs := getAllImgNames(dPath)

	gMapURLs := getImageLocations(imgs)

	if glog.V(2) {
		glog.Info("Images And There location:")
	}
	// display output
	if !fmtJSON {
		for fName, u := range gMapURLs {
			fmt.Printf("\n%s: %s\n", fName, u)
		}
	} else {
		data, err := json.Marshal(gMapURLs)
		if err != nil {
			glog.Errorf("failed to convert to JSON. Error: %v", err)
			os.Exit(1)
		}
		fmt.Printf("%s", data)
	}
}
