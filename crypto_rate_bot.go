// Crypto Currency Rate Telegram Bot

package main

import (
	"errors"
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

// GetCryptoCurerncyData from cryptonator API
func GetCryptoCurerncyData(cryptoCode, currencyCode string) ([]byte, error) {
	url := fmt.Sprintf("https://api.cryptonator.com/api/full/%s-%s", cryptoCode, currencyCode)
	resp, err := http.Get(url)
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

// GetCryptoCurerncyRate and return CurrencyRate
func GetCryptoCurerncyRate(cryptoCode string) (CurrencyRate, error) {
	usdBody, _ := GetCryptoCurerncyData(cryptoCode, "usd")
	usdData := []byte(usdBody)
	eurBody, _ := GetCryptoCurerncyData(cryptoCode, "eur")
	eurData := []byte(eurBody)
	rubBody, _ := GetCryptoCurerncyData(cryptoCode, "rub")
	rubData := []byte(rubBody)

	usdValue, _, _, _ := jsonparser.Get(usdData, "ticker", "price")
	eurValue, _, _, _ := jsonparser.Get(eurData, "ticker", "price")
	rubValue, _, _, _ := jsonparser.Get(rubData, "ticker", "price")

	intUsdValue, _ := strconv.ParseFloat(string(usdValue), 10)
	intEurValue, _ := strconv.ParseFloat(string(eurValue), 10)
	intRubValue, _ := strconv.ParseFloat(string(rubValue), 10)

	if int(intUsdValue) == 0 && int(intEurValue) == 0 && int(intRubValue) == 0 {
		return CurrencyRate{0, 0, 0}, errors.New("Undefined crypto currency code")
	}

	return CurrencyRate{
		usdValue: int(intUsdValue),
		eurValue: int(intEurValue),
		rubValue: int(intRubValue),
	}, nil
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
		if update.InlineQuery != nil {
			if len(update.InlineQuery.Query) > 0 {
				fmt.Println(update.InlineQuery.Query)
				rate, err := GetCryptoCurerncyRate(update.InlineQuery.Query)

				if err != nil {
					continue
				}

				text := "*" + update.InlineQuery.Query + "*\n\n" + "Дата: "
				text += GetHumanDate() + "\nВремя: " + GetHumanTime() + "\n\n"
				text += fmt.Sprintf("*USD*: %d $\n", rate.usdValue)
				text += fmt.Sprintf("*EUR*: %d €\n", rate.eurValue)
				text += fmt.Sprintf("*RUB*: %d ₽\n", rate.rubValue)

				msgContent := tgbotapi.InputTextMessageContent{Text: text, ParseMode: "markdown"}
				msg := tgbotapi.InlineQueryResultArticle{
					Type:                "article",
					ID:                  "1",
					Title:               update.InlineQuery.Query,
					Description:         "Узнать курс",
					InputMessageContent: msgContent,
				}
				var results []interface{}
				results = append(results, msg)
				config := tgbotapi.InlineConfig{InlineQueryID: update.InlineQuery.ID, Results: results}
				bot.AnswerInlineQuery(config)
				continue
			}
		}

		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" {
			text := "Выберите код криптовалюты или введите свой"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

			var keyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("BTC"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("ETH"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("XRP"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("BCH"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("LTC"),
				),
			)

			keyboard.OneTimeKeyboard = false
			keyboard.ResizeKeyboard = true

			msg.ParseMode = "markdown"
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			continue
		}

		// Don't process codes longer than 5 characters
		if len(update.Message.Text) > 5 {
			continue
		}

		rate, err := GetCryptoCurerncyRate(update.Message.Text)
		if err != nil {
			text := fmt.Sprintf("Криптовалюты с кодом *%s* не существует", update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			msg.ParseMode = "markdown"
			bot.Send(msg)
			continue
		}
		text := "*" + update.Message.Text + "*\n\n" + "Дата: "
		text += GetHumanDate() + "\nВремя: " + GetHumanTime() + "\n\n"
		text += fmt.Sprintf("*USD*: %d $\n", rate.usdValue)
		text += fmt.Sprintf("*EUR*: %d €\n", rate.eurValue)
		text += fmt.Sprintf("*RUB*: %d ₽\n", rate.rubValue)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ParseMode = "markdown"
		bot.Send(msg)
	}
}
