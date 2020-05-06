package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"unicode"
)

type Format string
const (
	Auction		Format = "auction"
	BuyItNow	Format = "buy-it-now"
)
type Listing struct {
	Url					string	`json:"url"`
	ImageUrl			string	`json:"imageUrl"`
	EbayItemNumber		string	`json:"ebayItemNumber"`

	SellerName					string	`json:"sellerName"`
	SellerStars					int		`json:"sellerStars"`
	SellerFeedbackPercentage	float32	`json:"sellerFeedbackPercentage"`

	Format			Format	`json:"format"`
	Location		string	`json:"location"`
	Title			string	`json:"title"`
	Condition		string	`json:"condition"`
	Price			float32	`json:"price"`
	Currency		string
	Postage			int		`json:"postage"`
	CanMakeOffer	bool	`json:"canMakeOffer"`
	Returns			string	`json:"returns"`
}

func GetPrice(price string) (float32, error) {
	// some countries use commas instead of dots so replace commas with dots
	if strings.Index(price, ",") != -1 {
		price = strings.ReplaceAll(price, ".", "")
		price = strings.ReplaceAll(price, ",", ".")
	}
	foundStartingIndex := false
	var startingIndex, endingIndex int
	for i, rune := range price {
		if unicode.IsDigit(rune) || rune == '.' {
			if !foundStartingIndex {
				startingIndex = i
				foundStartingIndex = true
			}
			continue
		}
		if foundStartingIndex {
			endingIndex = i - 1
			break
		}
	}

	if endingIndex == 0 {
		endingIndex = len(price)
	}

	float, err := strconv.ParseFloat(price[startingIndex:endingIndex], 32)
	if err != nil {
		return 0, errors.Wrap(err, "could not convert price to float")
	}
	return float32(float), nil
}

func GetListing(url string, currency string, doc *goquery.Document) (*Listing, error) {
	listing := &Listing{}

	listing.Url = url

	imageUrl, exists := doc.Find("img#icImg").Attr("src")
	if !exists {
		return nil, errors.New("src attribute does not exist on img#icImg")
	}
	listing.ImageUrl = imageUrl

	listing.EbayItemNumber = doc.Find("div#descItemNumber").Text()

	listing.SellerName = doc.Find("span.mbg-nw").Text()

	sellerStars, err := strconv.Atoi(doc.Find("span.mbg-l").Children().Eq(0).Text())
	if err != nil {
		return nil, errors.Wrap(err, "could convert seller stars text to int")
	}
	listing.SellerStars = sellerStars

	sellerFeedbackPercentageText := strings.TrimSpace(doc.Find("div#si-fb").Text())
	listing.SellerFeedbackPercentage = -1
	if sellerFeedbackPercentageText != "" {
		sellerFeedbackPercentage, err := strconv.ParseFloat(sellerFeedbackPercentageText[0:strings.Index(sellerFeedbackPercentageText, "%")], 32)
		if err != nil {
			return nil, errors.New("could not determine SellerFeedbackPercentage")
		}
		listing.SellerFeedbackPercentage = float32(sellerFeedbackPercentage)
	}

	price, err := GetPrice(doc.Find("span#prcIsum").Text())
	if err != nil {
		return nil, errors.Wrap(err, "could not determine price")
	}
	listing.Price = price

	format := BuyItNow
	if len(doc.Find("a#bidBtn_btn").Nodes) > 0 {
		format = Auction
	}
	listing.Format = format

	listing.Currency = currency

	listing.Location = doc.Find("span[itemprop=availableAtOrFrom]").Text()

	itemTitleSel := doc.Find("h1#itemTitle")
	itemTitleSel.Find("span.g-hdn").Remove()
	listing.Title = itemTitleSel.Text()

	listing.Condition = doc.Find("div#vi-itm-cond").Text()

	listing.CanMakeOffer = false
	if len(doc.Find("a#boBtn_btn").Nodes) > 0 {
		listing.CanMakeOffer = true
	}

	return listing, nil
}
