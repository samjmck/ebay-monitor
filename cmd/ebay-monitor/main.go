package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

type SearchItem struct {
	Url string
	Currency string
}

func loadConfig() error {
	// Load config.toml
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "could not load config.toml")
	}

	// Load .env
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	err = viper.MergeInConfig()
	if err != nil {
		return errors.Wrap(err, "could not load .env")
	}

	return nil
}

func getSearchItems() ([]SearchItem, error) {
	var searchItems []SearchItem

	// Load SearchItem items from config.toml
	err := viper.UnmarshalKey("searches", &searchItems)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal searchItems key of config")
	}

	return searchItems, nil
}

// This func will probably be ran in a new goroutine which means we should not
// use a pointer to []*Listing, we should use a channel. With the channel, we can
// stop race conditions. This isn't such a big problem though as the HTTP pulling
// is not a core function of the program
func startWebServer(pullListings *[]*Listing) error {
	http.HandleFunc("/pull_listings", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(*pullListings)
		if err != nil {
			fmt.Printf("Could not encode json for /pull_listings: %v\n", err)
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

func startScraping(searchItems []SearchItem, trackScrapedUrls bool, newListing func(searchUrl string, listing *Listing)) error {
	var scraped map[string]map[string]bool
	var encoder *json.Encoder
	var scrapedf *os.File
	decoded := false
	if trackScrapedUrls {
		// scraped.json will be for storing the already scraped URLs
		scrapedf, err := os.OpenFile("scraped.json", os.O_RDWR | os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "could not open scraped.json")
		}
		defer scrapedf.Close()

		err = json.NewDecoder(scrapedf).Decode(&scraped)
		decoded = err == nil
		if err != nil {
			fmt.Printf("Not using data in scraped.json: %v\n", err)
		}

		encoder = json.NewEncoder(scrapedf)
	}
	if !decoded {
		scraped = map[string]map[string]bool{}
	}

	for {
		for _, searchItem := range searchItems {
			searchUrl := searchItem.Url

			fmt.Println("Searching with", searchUrl)
			doc, err := Get(searchUrl)
			if err != nil {
				fmt.Printf("Could not make request to SearchItem page: %v", err)
				continue
			}

			// Returning false within this loop will break out of the loop
			doc.Find("a.s-item__link").EachWithBreak(func(i int, sel *goquery.Selection) bool {
				url, exists := sel.Attr("href")
				if !exists {
					return true
				}

				// Initialise map for this searchUrl if it doesn't already exist
				if len(scraped[searchUrl]) == 0 {
					scraped[searchUrl] = map[string]bool{}
				}

				if scraped[searchUrl][url] {
					if i == 0 {
						fmt.Println("Found nothing new")
					}
					return false
				}

				// Set the value of this url to true to indicate that it has already been scraped
				scraped[searchUrl][url] = true

				if trackScrapedUrls {
					// Update scraped.json with new contents
					// Reset cursor so the file contents will be overwritten
					scrapedf.Seek(0, io.SeekStart)
					err = encoder.Encode(scraped)
					if err != nil {
						log.Fatalf("Could not encode to scraped.json: %s\n", err)
					}
				}

				fmt.Println("\nVisiting new item page", url)
				doc, err := Get(url)
				if err != nil {
					fmt.Printf("Could not load item page: %v", err)
					return true
				}

				listing, err := GetListing(url, searchItem.Currency, doc)
				if err != nil {
					fmt.Printf("Could not get listing details: %v", err)
					return true
				}

				fmt.Println("Got listing details")

				newListing(searchUrl, listing)

				if len(scraped[searchUrl]) == 1 {
					// This was the first time scraping this searchUrl. As we only want to check for new listings,
					// we won't scrape all the next listings and we will just wait for new ones. This is why we
					// will break out of the loop.
					return false
				}

				return true
			})
		}
		time.Sleep(time.Duration(viper.GetInt("delay")) * time.Second)
	}
}

func main() {
	err := loadConfig()
	if err != nil {
		log.Fatalf("Could not load configs: %v\n", err)
	}

	searchItems, err := getSearchItems()
	if err != nil {
		log.Fatalf("Could not get SearchItem URLs: %v\n", err)
	}

	// TODO: use channel to stop race condition
	pullListings := []*Listing{}

	useWebServer := viper.GetBool("web-server")
	if useWebServer {
		go func() {
			err = startWebServer(&pullListings)
			if err != nil {
				fmt.Printf("Could not start web server: %v\n", err)
			}
		}()
	}

	tpl, err := template.New("message").Parse(viper.GetString("message"))
	if err != nil {
		fmt.Printf("Could not parse message template: %v\n", err)
	}

	err = startScraping(searchItems, viper.GetBool("track-scraped-urls"), func(_ string, listing *Listing) {
		pullListings = append(pullListings, listing)

		buf := &bytes.Buffer{}
		err = tpl.Execute(buf, *listing)
		var msg string
		if err != nil {
			fmt.Printf("Could not execute template: %v\n", err)
			msg = listing.Url
		} else {
			msg = buf.String()
		}

		err = SendTelegramMessage(
			viper.GetString("TELEGRAM_TOKEN"),
			viper.GetString("TELEGRAM_CHAT_ID"),
			msg,
			)
		if err != nil {
			fmt.Printf("Could not send Telegram message: %v\n", err)
		}
	})
	if err != nil {
		log.Fatalf("Could not start scraping: %v\n", err)
	}
}