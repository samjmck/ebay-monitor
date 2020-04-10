package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

func Get(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// eBay gives a page that's formatted differently if we don't use a desktop User Agent
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func SendTelegramMessage(token string, chatId string, message string) error {
	data := fmt.Sprintf("{\"chat_id\": \"%s\", \"text\": \"%s\"}", chatId, message)
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", token), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not make Telegram request")
	}
	defer resp.Body.Close()

	if resp == nil {
		return errors.New("no response from Telegram request")
	}

	if resp.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("Telegram responded with status code %v", resp.StatusCode))
	}

	return nil
}
