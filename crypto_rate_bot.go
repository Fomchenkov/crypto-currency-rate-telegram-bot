// Crypto Currency Rate Telegram Bot

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// CurrencyRate USD and RUB
type CurrencyRate struct {
	usdValue, eurValue, rubValue int
}

// GetBTCData from API
func GetBTCData() ([]byte, error) {
	resp, err := http.Get("https://blockchain.info/ticker")
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

// GetBTCRate and return CurrencyRate
func GetBTCRate() CurrencyRate {
	body, _ := GetBTCData()
	data := []byte(body)
	usdValue, _, _, _ := jsonparser.Get(data, "USD", "last")
	eurValue, _, _, _ := jsonparser.Get(data, "EUR", "last")
	rubValue, _, _, _ := jsonparser.Get(data, "RUB", "last")

	intUsdValue, _ := strconv.ParseFloat(string(usdValue), 10)
	intEurValue, _ := strconv.ParseFloat(string(eurValue), 10)
	intRubValue, _ := strconv.ParseFloat(string(rubValue), 10)

	return CurrencyRate{
		usdValue: int(intUsdValue),
		eurValue: int(intEurValue),
		rubValue: int(intRubValue),
	}
}

// GetHumanDate returns human format of date
func GetHumanDate() string {
	start := time.Now()
	return fmt.Sprintf("%d.%d.%d", start.Day(), start.Month(), start.Year())
}

// GetHumanTime returns human format of time
func GetHumanTime() string {
	start := time.Now()
	return fmt.Sprintf("%d:%d:%d", start.Hour(), start.Minute(), start.Second())
}

func main() {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" {
			text := "Выберите валюту"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

			var keyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("BTC"),
				),
			)

			keyboard.OneTimeKeyboard = false
			keyboard.ResizeKeyboard = true

			msg.ParseMode = "markdown"
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}

		if update.Message.Text == "BTC" {
			rate := GetBTCRate()
			text := "*BTC*\n\n" + "Дата: " + GetHumanDate() + "\nВремя: " + GetHumanTime() + "\n\n"
			text += fmt.Sprintf("*USD*: %d $\n", rate.usdValue)
			text += fmt.Sprintf("*EUR*: %d €\n", rate.eurValue)
			text += fmt.Sprintf("*RUB*: %d ₽\n", rate.rubValue)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			msg.ParseMode = "markdown"
			bot.Send(msg)
		}
	}
}
