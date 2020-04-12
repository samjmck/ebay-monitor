#!/bin/zsh

# Raspberry pi builds
env GOOS=linux GOARCH=arm GOARM=6 go build ./cmd/ebay-monitor/
env GOOS=linux GOARCH=arm GOARM=6 go build ./cmd/telegram-chat-id/
zip -r ebay-monitor-linux-armv6l.zip ebay-monitor telegram-chat-id

rm ebay-monitor
rm telegram-chat-id

# Darwin/macOS builds
go build ./cmd/ebay-monitor/
go build ./cmd/telegram-chat-id/
zip -r ebay-monitor-darwin-amd64.zip ebay-monitor telegram-chat-id

rm ebay-monitor
rm telegram-chat-id

# Config files
zip -r config-files.zip .env config.toml