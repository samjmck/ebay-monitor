# ebay-monitor
Scrapes eBay search pages so you don't have to revisit them manually

## Quick install for Raspberry Pi
Change directory to where you would like to install the program and then run the following commands.

```shell script
curl -OL https://github.com/samjmckenzie/ebay-monitor/releases/latest/download/ebay-monitor-linux-armv6l.zip
curl -OL https://github.com/samjmckenzie/ebay-monitor/releases/latest/download/config-files.zip
unzip ebay-monitor-linux-armv6l.zip
unzip config-files.zip
chmod +x ebay-monitor telegram-chat-id
```

## Installation
1. Download the latest release for your system architecture as well as the config files [here](https://github.com/samjmckenzie/ebay-monitor/releases).
2. Unzip the binary and config zip files.
3. Add execute permissions to the binary files by running `chmod +x ebay-monitor` and `chmod +x telegram-chat-id`.

## Setting up
The main purpose of the program is to notify you whenever a new listing appears on an eBay search page. We will be using the Telegram messaging app to send these notifications as the app easily allows you to create bots. 
You can either download the app [on desktop](https://desktop.telegram.org/) or on mobile. I'd suggest using the desktop for the setup as you will need to copy an API key.

Once you've downloaded the app and have setup an account, you can create the bot by sending a message to BotFather. You can do this by either opening [this](https://t.me/BotFather) link or by looking up the name "BotFather" when trying to send a new message.
After you select its name, click on the start button at the bottom of your screen. Send the message "/newbot" to it and it will guide you through the bot creation process.
When you finish the creation process, you should receive a message that contains a token to access the HTTP API token. Copy this token and paste it into the `TELEGRAM_TOKEN` field of your `.env` so it looks like this:

```.env
TELEGRAM_TOKEN="1222533313:AAFwNd_HsPtpxBy35vEaZoFzUUB74v5mBpW"
```

Now you need to get the chat ID which the bot will send messages to. You can get this by sending your bot a message and then running the `telegram-chat-id` binary, after which you will see a message like this in your terminal:
```
From YOUR-TELEGRAM
667630712
``` 

The last line of the message is the Telegram chat ID. Your `.env` file should automatically be updated after running the binary. If it isn't, paste it into the `TELEGRAM_CHAT_ID` field.

## Configuration
You can make the program search for new item listings by simply searching for that item like you normally would and then sorting by newly listed.
Now copy the URL and add the following to `config.toml`:
```toml
[[searches]]
url = "copied URL"
```
You can also use filters as they are visible in the URL.

If, for example, I wanted the program to search for new listings of "duct tape" and "macbook pro", the bottom of `config.toml` would look like this:
```toml
[[searches]]
url = "https://www.ebay.co.uk/sch/i.html?_from=R40&_nkw=duct+tape&_sacat=0&_sop=10"

[[searches]]
url = "https://www.ebay.co.uk/sch/i.html?_from=R40&_nkw=macbook+pro&_sacat=0&_sop=10"
```

The other keys in `config.toml` will be explained here:
- `interval` is the interval in seconds between each loop of scraping
- `message` is the format of the Telegram message
- `web-server` indicates whether the web server will be ran from which you can pull new listings. This can act as an alternative to the Telegram messages.
- `track-scraped-urls` will add scraped URLs to scraped.json to save the last position, but it is not essential to running the program

## Running
To run the program, change directory to the folder with all the program files and run `ebay-monitor`. 