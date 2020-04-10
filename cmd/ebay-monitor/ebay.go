package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"strconv"
	"strings"
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
	Price			[2]int	`json:"price"`
	Postage			int		`json:"postage"`
	CanMakeOffer	bool	`json:"canMakeOffer"`
	Returns			string	`json:"returns"`
}

func GetListing(url string, doc *goquery.Document) (*Listing, error) {
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

	format := BuyItNow
	if len(doc.Find("a#bidBtn_btn").Nodes) > 0 {
		format = Auction
	}
	listing.Format = format

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
