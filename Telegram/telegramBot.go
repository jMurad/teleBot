package Telegram

import (
	dbase "TeleBot/Database"
	"TeleBot/Duty"
	fncs "TeleBot/Functions"
	kbrd "TeleBot/Keyboards"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TeleBot struct {
	Bot     *tgb.BotAPI
	Updates tgb.UpdatesChannel
}

func (tb *TeleBot) TBInit() {
	//Создаем бота
	token := os.Getenv("TOKEN")
	var err error
	tb.Bot, err = tgb.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	tb.Bot.Debug = true

	log.Printf("Authorized on account %s", tb.Bot.Self.UserName)

	host := os.Getenv("HOST")
	cert := os.Getenv("CERT")

	wh, _ := tgb.NewWebhookWithCert(host+tb.Bot.Token, cert)
	_, err = tb.Bot.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := tb.Bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	//Получаем обновления от бота
	tb.Updates = tb.Bot.ListenForWebhook("/" + tb.Bot.Token)
}

func (tb *TeleBot) sendMsg(md bool, id int64, text string, kb interface{}) {
	msg := tgb.NewMessage(id, text)
	msg.ReplyMarkup = kb
	if md {
		msg.ParseMode = "Markdown"
	}
	if _, err := tb.Bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (tb *TeleBot) sendPht(name string, id int64, text string, kb interface{}) {
	pht := tgb.NewPhoto(id, os.Getenv(name))
	pht.ReplyMarkup = kb
	pht.Caption = text
	if _, err := tb.Bot.Send(pht); err != nil {
		log.Panic(err)
	}
}

func (tb *TeleBot) RunBot(dej *Duty.Dejurnie, newdej chan Duty.Dejurnie, db *dbase.TGDB) {
	//tb.startServer()

	reNumber := regexp.MustCompile(`^\d\d?`)
	reCalendarDay := regexp.MustCompile(`^\d\d? Дневная \x{1F31D}`)
	reCalendarNight := regexp.MustCompile(`^\d\d? Ночная \x{1F31A}`)
	reCalendar := regexp.MustCompile(`^\d\d? Дневная|Ночная [\x{1F31A}\x{1F31D}]`)

	var (
		currentMenu     = "Меню"
		currentCalendar = false
	)

	for update := range tb.Updates {
		if !Duty.RunningParse {
			listDept := dej.GetListDept()
			if update.Message != nil {
				text := update.Message.Text
				id := update.Message.Chat.ID
				userId := strconv.Itoa(int(id))
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
								tb.sendMsg(true, id, report[i:], kbrd.MainMenu)
							} else {
								tb.sendMsg(true, id, report[i:i+lenReport], kbrd.MainMenu)
							}
							i += lenReport
						}
					case text == "/start":
						tempText := "Ассалам алейкум! Я скажу тебе кто сейчас на смене!"
						tb.sendMsg(false, id, tempText, kbrd.MainMenu)
						currentCalendar = false
					case text == "Кто сейчас на смене?":
						today := time.Now().Local()
						for _, nameDuty := range dej.WhoDutyAll(today) {
							tempText := nameDuty + " - Дежурный " + dej.DutyToDept(nameDuty)
							tb.sendPht(nameDuty, id, tempText, kbrd.InKeyMkr(dej.GetSched(nameDuty)))
						}
					case text == "Дежурные":
						tb.sendMsg(false, id, "Дежурные", kbrd.GetListDept(listDept))
						currentMenu = "Меню"
					case text == "Календарь":
						tb.sendMsg(false, id, "Календарь", kbrd.CdrKeyMkr())
						currentMenu = "Меню"
						currentCalendar = true
					case text == "Назад":
						switch currentMenu {
						case "Календарь":
							tb.sendMsg(false, id, "Календарь", kbrd.CdrKeyMkr())
							currentMenu = "Меню"
						case "Дежурные":
							tb.sendMsg(false, id, "Дежурные", kbrd.GetListDept(listDept))
							currentMenu = "Меню"
						case "Меню":
							tb.sendMsg(false, id, "Меню", kbrd.MainMenu)
							currentCalendar = false
						}
					case reCalendar.MatchString(text):
						var selDate time.Time
						selDay, _ := strconv.Atoi(reNumber.FindString(text))
						year := time.Now().Local().Year()
						month := time.Now().Local().Month()
						if reCalendarDay.MatchString(text) {
							selDate = time.Date(year, month, selDay, 15, 00, 0, 0, time.Local)
						} else if reCalendarNight.MatchString(text) {
							selDate = time.Date(year, month, selDay, 22, 00, 0, 0, time.Local)
						}
						for _, nameDuty := range dej.WhoDutyAll(selDate) {
							tempText := nameDuty + " - Дежурный " + dej.DutyToDept(nameDuty)
							tb.sendPht(nameDuty, id, tempText, kbrd.InKeyMkr(dej.GetSched(nameDuty)))
						}
					case fncs.StrInArray(dej.GetListDutyAll(), text) != -1:
						tempText := text + " - Дежурный " + dej.DutyToDept(text)
						tb.sendPht(text, id, tempText, kbrd.InKeyMkr(dej.GetSched(text)))
					case fncs.StrInArray(listDept, text) != -1:
						tb.sendMsg(false, id, text, kbrd.GetListDuty(dej.GetListDuty(text)))
						currentMenu = "Дежурные"
					case fncs.IfStrDay(text) && currentCalendar:
						tb.sendMsg(false, id, "День/Ночь", kbrd.GetMenuDN(strings.Trim(text, "-")))
						currentMenu = "Календарь"
					default:
						//Отправлем сообщение
						tb.sendMsg(false, id, fncs.RandomRustam(), kbrd.MainMenu)
						currentCalendar = false
					}
					db.AddToLog(userId, firstName, text)
				} else {
					//Отправляем сообщение
					tb.sendMsg(false, id, fncs.RandomRustam(), kbrd.MainMenu)
					currentCalendar = false
				}
			} else if update.CallbackQuery != nil {
				callback := tgb.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
				if _, err := tb.Bot.Request(callback); err != nil {
					panic(err)
				}
			}
		} else {
			*dej = <-newdej
		}
	}
}
