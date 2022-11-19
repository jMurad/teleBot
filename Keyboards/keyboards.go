package Keyboards

import (
	fncs "TeleBot/Functions"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"time"
)

var MainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Кто сейчас на смене?"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Дежурные"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Календарь"),
	),
)

func GetListDept(dept []string) tgbotapi.ReplyKeyboardMarkup {

	var keyboard [][]tgbotapi.KeyboardButton
	for _, ld := range dept {
		keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ld)))
	}
	keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Назад")))
	return tgbotapi.NewReplyKeyboard(keyboard...)
}

func GetMenuDN(day string) tgbotapi.ReplyKeyboardMarkup {
	menuDayNight := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(day+" Дневная 🌝"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(day+" Ночная 🌚"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	return menuDayNight
}

func InKeyMkr(schedule [31]string) tgbotapi.InlineKeyboardMarkup {
	var row []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	i := 0
	for day, smn := range schedule {
		if smn == "Day" {
			i += 1
			text := strconv.Itoa(day+1) + " 🌝"
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(text, "🌝"))
		} else if smn == "Night" {
			i += 1
			text := strconv.Itoa(day+1) + " 🌚"
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(text, "🌚"))
		}
		if i > 2 {
			i = 0
			keyboard = append(keyboard, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	keyboard = append(keyboard, row)
	row = []tgbotapi.InlineKeyboardButton{}
	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

func CdrKeyMkr() tgbotapi.ReplyKeyboardMarkup {
	var row []tgbotapi.KeyboardButton
	var keyboard [][]tgbotapi.KeyboardButton
	i := 0
	t := time.Now()
	firstday := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	lastday := firstday.AddDate(0, 1, 0).Add(time.Nanosecond * -1).Format("2")
	ld, _ := strconv.Atoi(lastday)
	fncs.LastDayOfMonth(time.Now())
	for day := 1; day <= ld; day++ {
		i += 1
		text := strconv.Itoa(day)
		row = append(row, tgbotapi.NewKeyboardButton(text))
		if i > 7 {
			i = 0
			keyboard = append(keyboard, row)
			row = []tgbotapi.KeyboardButton{}
		}
	}
	keyboard = append(keyboard, row)
	row = []tgbotapi.KeyboardButton{}
	keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Назад")))
	return tgbotapi.NewReplyKeyboard(keyboard...)
}

func GetListDuty(listDuty []string) tgbotapi.ReplyKeyboardMarkup {
	var row []tgbotapi.KeyboardButton
	var keyboard [][]tgbotapi.KeyboardButton
	for i, ld := range listDuty {
		row = append(row, tgbotapi.NewKeyboardButton(ld))
		if (i+1)%2 == 0 {
			keyboard = append(keyboard, row)
			row = []tgbotapi.KeyboardButton{}
		}
		//keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ld)))
	}
	keyboard = append(keyboard, row)
	keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Назад")))
	return tgbotapi.NewReplyKeyboard(keyboard...)
}
