package Telegram

import (
	dbpg "TeleBot/Database"
	"TeleBot/Duty"
	fncs "TeleBot/Functions"
	kbrd "TeleBot/Keyboards"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var db dbpg.DatabasePG

func TeleBot(dej *Duty.Dejurnie) {
	//Инициализация БД
	db.DBinit()

	//Создаем бота
	token := os.Getenv("TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	updates := bot.ListenForWebhook("/" + bot.Token)
	// updates := bot.GetUpdatesChan(u)

	newdej := make(chan Duty.Dejurnie)
	go dej.CronXLSX(newdej)

	//Слушаем Telegram
	go http.ListenAndServeTLS("176.124.209.183:8443", "SELF_SIGN_CERT.pem", "SELF_SIGN_CERT.key", nil)

	reNumber := regexp.MustCompile(`^\d\d?`)
	reCalendarDay := regexp.MustCompile(`^\d\d? Дневная \x{1F31D}`)
	reCalendarNight := regexp.MustCompile(`^\d\d? Ночная \x{1F31A}`)
	reCalendar := regexp.MustCompile(`^\d\d? Дневная|Ночная [\x{1F31A}\x{1F31D}]`)

	var (
		currentMenu     = "Меню"
		currentCalendar = false
	)

	for update := range updates {
		if !Duty.RunningParse {
			listDept := dej.GetListDept()
			if update.Message != nil {
				text := update.Message.Text
				userId := strconv.Itoa(int(update.Message.Chat.ID))
				firstName := update.Message.Chat.FirstName
				//Проверяем что от пользователья пришло именно текстовое сообщение
				if reflect.TypeOf(text).Kind() == reflect.String && text != "" {
					switch {
					case db.ParseAdminMsg(text) && userId == "422322499":
						report := db.PrettyLog()
						i := 0
						lenReport := 4095
						for {
							if i > len(report) {
								break
							}
							if i+lenReport >= len(report) {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, report[i:len(report)])
								msg.ParseMode = "Markdown"
								msg.ReplyMarkup = kbrd.MainMenu
								if _, err := bot.Send(msg); err != nil {
									log.Panic(err)
								}
							} else {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, report[i:i+lenReport])
								msg.ParseMode = "Markdown"
								msg.ReplyMarkup = kbrd.MainMenu
								if _, err := bot.Send(msg); err != nil {
									log.Panic(err)
								}
							}
							i += lenReport

						}
					case text == "/start":
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ассалам алейкум! Я скажу тебе кто сейчас на смене!")
						msg.ReplyMarkup = kbrd.MainMenu
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						currentCalendar = false
					case text == "Кто сейчас на смене?":
						today := time.Now().Local()

						for _, nameDuty := range dej.WhoDutyAll(today) {
							gpi := fncs.GetPathImg(nameDuty)
							pht := tgbotapi.NewPhoto(update.Message.Chat.ID, gpi)
							pht.ReplyMarkup = kbrd.InlineKeyboardMaker(dej.GetSchedule(nameDuty))
							pht.Caption = nameDuty + " - Дежурный " + dej.DutyToDept(nameDuty)
							if _, err := bot.Send(pht); err != nil {
								log.Panic(err)
							}
						}
					case text == "Дежурные":
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Дежурные")
						msg.ReplyMarkup = kbrd.GetListDept(listDept)
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						currentMenu = "Меню"
					case text == "Календарь":
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Календарь")
						msg.ReplyMarkup = kbrd.CalendarKeyboardMaker()
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						currentMenu = "Меню"
						currentCalendar = true
					case text == "Назад":
						switch currentMenu {
						case "Календарь":
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Календарь")
							msg.ReplyMarkup = kbrd.CalendarKeyboardMaker()
							if _, err := bot.Send(msg); err != nil {
								log.Panic(err)
							}
							currentMenu = "Меню"
						case "Дежурные":
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Дежурные")
							msg.ReplyMarkup = kbrd.GetListDept(listDept)
							if _, err := bot.Send(msg); err != nil {
								log.Panic(err)
							}
							currentMenu = "Меню"
						case "Меню":
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
							msg.ReplyMarkup = kbrd.MainMenu
							if _, err := bot.Send(msg); err != nil {
								log.Panic(err)
							}
							currentCalendar = false
						}
					case reCalendar.MatchString(text):
						var selDate time.Time
						selDay, _ := strconv.Atoi(reNumber.FindString(text))
						year := time.Now().Local().Format("2006")
						month := time.Now().Local().Format("1")
						if reCalendarDay.MatchString(text) {
							selDate = time.Date(year, month, selDay, 15, 00, 0, 0, time.Local)
						} else if reCalendarNight.MatchString(text) {
							selDate = time.Date(year, month, selDay, 22, 00, 0, 0, time.Local)
						}

						for _, nameDuty := range dej.WhoDutyAll(selDate) {
							pht := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty))
							pht.ReplyMarkup = kbrd.InlineKeyboardMaker(dej.GetSchedule(nameDuty))
							pht.Caption = nameDuty + " - Дежурный " + dej.DutyToDept(nameDuty)
							if _, err := bot.Send(pht); err != nil {
								log.Panic(err)
							}
						}
					case fncs.StrInArray(dej.GetListDutyAll(), text) != -1:
						pht := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(text))
						pht.ReplyMarkup = kbrd.InlineKeyboardMaker(dej.GetSchedule(text))
						pht.Caption = text + " - Дежурный " + dej.DutyToDept(text)
						if _, err := bot.Send(pht); err != nil {
							log.Panic(err)
						}
					case fncs.StrInArray(listDept, text) != -1:
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
						msg.ReplyMarkup = kbrd.GetListDuty(dej.GetListDuty(text))
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						currentMenu = "Дежурные"
					case fncs.IfStrDay(text) && currentCalendar:
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "День/Ночь")
						msg.ReplyMarkup = kbrd.GetMenuDayNight(strings.Trim(text, "-"))
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						currentMenu = "Календарь"
					default:
						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fncs.RandomRustam())
						msg.ReplyMarkup = kbrd.MainMenu
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						currentCalendar = false
					}
					db.AddToLog(userId, firstName, text)
				} else {
					//Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fncs.RandomRustam())
					msg.ReplyMarkup = kbrd.MainMenu
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					currentCalendar = false
				}
			} else if update.CallbackQuery != nil {
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
				if _, err := bot.Request(callback); err != nil {
					panic(err)
				}
			}
		} else {
			*dej = <-newdej
		}
	}
}
