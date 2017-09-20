package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"log"
	"regexp"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

var regImages = regexp.MustCompile(`src\s*=\s*"(.*?)\.(jpeg|png|gif|jpg|webm)"`)
var regAnchors = regexp.MustCompile(`href\s*=\s*"(.*?)\.(jpeg|png|gif|jpg|webm)"`)
var client = http.Client{Timeout: time.Duration(1 * time.Hour)}

// Download sends get request to URL and writes response to file
func Download(url string) {

	url = Extract(url)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "image/*,video/webm")
	resp, err := client.Do(req)

	if err != nil {

		log.Println("FAILED: ", err)
		log.Println(url)
		return
	}

	defer resp.Body.Close()

	re, _ := ioutil.ReadAll(resp.Body)

	if ok := ValidMime(re); !ok {
		return
	}

	filetype := strings.Split(url, ".")

	file, err := os.Create(uuid.NewV4().String() + "." + filetype[len(filetype)-1])

	if err != nil {
		log.Println("Error creating File")
		return
	}
	defer file.Close()
	file.Write(re)
	log.Println("SUCCESS: ", url)

}

// ValidMime checks if the Mime-Types match image or video
func ValidMime(data []byte) bool {

	mime := http.DetectContentType(data)

	if strings.Contains(mime, "image") || strings.Contains(mime, "video") {
		return true
	}

	return false
}

// Extract makes static/unfit Urls useable
func Extract(url string) string {
	arr := strings.Split(url, "")
	if arr[0] == "/" {
		for i := 0; i < len(url); i++ {
			if arr[i] != "/" {
				return "https://" + strings.Join(arr[i:], "")
			}
		}
	}
	return url
}

// Filter sends get request to url and filters for image url's
func Filter(url string) {
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "image/*,video/webm")
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	byteResponse, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
	}

	final := string(byteResponse[:])

	foundAnchors := regAnchors.FindAllString(final, -1)

	foundImages := regImages.FindAllString(final, -1)

	if len(foundAnchors) == 0 || len(foundImages) == 0 {
		log.Println("Nothing Found")
		return
	}

	go Schedule(foundAnchors)
	go Schedule(foundImages)

}

// Schedule schedules the Downloads
func Schedule(found []string) {
	if len(found) == 0 {
		return
	}
	for item := range found {
		go Download(strings.Split(found[item], "\"")[1])
	}
}

func main() {

	var inp string
	fmt.Println("Enter Url or Type quit or Ctrl-C to Quit")
	for inp != "quit" {
		fmt.Scanln(&inp)
		Filter(inp)
	}

}
