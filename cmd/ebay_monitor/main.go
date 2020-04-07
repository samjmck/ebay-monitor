package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/pkg/errors"
)


func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "could not load config.toml")
	}

	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	err = viper.MergeInConfig()
	if err != nil {
		return errors.Wrap(err, "could not load .env")
	}

	return nil
}

func getSearchUrls() ([]string, error) {
	var searches []struct{
		Url string
	}

	err := viper.UnmarshalKey("searches", &searches)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal searches key of config")
	}

	searchUrls := []string{}
	for _, search := range searches {
		searchUrls = append(searchUrls, search.Url)
	}

	return searchUrls, nil
}

func startWebServer(pullListings *[]*Listing) error {
	http.HandleFunc("/pull_listings", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(*pullListings)
		if err != nil {
			fmt.Printf("Could not encode json for /pull_listings: %v", err)
			return
		}

		*pullListings = []*Listing{}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return errors.Wrap(err, "could not start web server on :8080")
	}

	return nil
}

func startScraping(searchUrls []string, pullListings *[]*Listing, usingWebServer bool) {
	scrapedf, err := os.OpenFile("scraped.json", os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not open scraped.json: %v\n", err)
	}
	defer scrapedf.Close()

	var scraped map[string]map[string]bool

	err = json.NewDecoder(scrapedf).Decode(&scraped)
	if err != nil {
		log.Fatalf("Could not decode scraped.json: %v\n", err)
	}

	encoder := json.NewEncoder(scrapedf)

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

				if usingWebServer {
					*pullListings = append(*pullListings, listing)
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

func main() {
	err := loadConfig()
	if err != nil {
		log.Fatalf("Could not load configs: %v\n", err)
	}

	searchUrls, err := getSearchUrls()
	if err != nil {
		log.Fatalf("Could not get search URLs: %v\n", err)
	}

	// TODO: use channel to stop race condition
	pullListings := []*Listing{}

	useWebServer := viper.GetBool("web-server")
	if useWebServer {
		go startWebServer(&pullListings)
	}

	startScraping(searchUrls, &pullListings, useWebServer)
}