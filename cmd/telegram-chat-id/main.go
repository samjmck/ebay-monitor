package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
)

type GetUpdatesResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"result"`
}

func main() {
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error .env: %s \n", err))
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", viper.GetString("TELEGRAM_TOKEN")), nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var body GetUpdatesResponse
	json.NewDecoder(resp.Body).Decode(&body)

	for _, result := range body.Result {
		fmt.Println("From", result.Message.From.Username)
		fmt.Println(result.Message.Chat.ID)
		viper.Set("TELEGRAM_CHAT_ID", result.Message.Chat.ID)
	}

	viper.WriteConfig()
}
