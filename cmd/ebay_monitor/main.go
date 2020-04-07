package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"time"
)

func StartWebServer(pullListings *[]*Listing) {
	http.HandleFunc("/pull_listings", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(*pullListings)
		if err != nil {
			fmt.Printf("Could not encode json for /pull_listings: %v", err)
			return
		}

		*pullListings = []*Listing{}
	})

	http.ListenAndServe(":8080", nil)
}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config.toml: %s \n", err))
	}

	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	err = viper.MergeInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error .env: %s \n", err))
	}

	p, _ := json.MarshalIndent(viper.Get("searches"), "", "  ")
	fmt.Println(string(p))

	var searches []struct{
		Url string
	}
	viper.UnmarshalKey("searches", &searches)
	searchUrls := []string{}
	for _, search := range searches {
		searchUrls = append(searchUrls, search.Url)
	}

	fmt.Println(searchUrls)

	scrapedf, err := os.OpenFile("scraped.json", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("Could not open scraped.json: %s \n", err))
	}
	defer scrapedf.Close()
	var scraped map[string]map[string]bool

	err = json.NewDecoder(scrapedf).Decode(&scraped)
	if err != nil {
		panic(fmt.Errorf("Could not decode scraped.json: %s \n", err))
	}

	encoder := json.NewEncoder(scrapedf)

	// TODO: use channel to stop race condition
	pullListings := []*Listing{}

	useWebServer := viper.GetBool("web-server")
	if useWebServer {
		go StartWebServer(&pullListings)
	}

	for {
		for _, searchUrl := range searchUrls {
			fmt.Println("Searching with", searchUrl)
			doc, err := Get(searchUrl)
			if err != nil {
				fmt.Printf("Could not make request to search page: %v", err)
				continue
			}

			doc.Find("a.s-item__link").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
				url, exists := sel.Attr("href")
				if !exists {
					return true
				}

				if len(scraped[searchUrl]) == 0 {
					scraped[searchUrl] = map[string]bool{}
				}

				if scraped[searchUrl][url] {
					fmt.Println("Already visited these links... will try again next time")
					return false
				}

				scraped[searchUrl][url] = true

				fmt.Println("\nVisiting new item page", url)
				doc, err := Get(url)
				if err != nil {
					fmt.Printf("Could not load item page: %v", err)
					return true
				}

				listing, err := GetListing(url, doc)
				if err != nil {
					fmt.Printf("Could not get listing details: %v", err)
					return true
				}

				fmt.Println("Got listing details")

				// Reset cursor so the file contents will be overwritten
				scrapedf.Seek(0, io.SeekStart)
				err = encoder.Encode(scraped)
				if err != nil {
					panic(fmt.Errorf("Could not encode to scraped.json: %s \n", err))
				}

				if useWebServer {
					pullListings = append(pullListings, listing)
				}

				if len(scraped[searchUrl]) == 1 {
					// This was the first time scraping this searchUrl so we will just check the first listing
					return false
				}

				return true
			})
		}
		time.Sleep(time.Duration(viper.GetInt("interval")) * time.Second)
	}
}