package main

import (
	fncs "TeleBot/Functions"
	kbrd "TeleBot/Keyboards"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)
const timeTempl = "2 1 2006 15:04 (MST)"

type dejurniy struct {
	name	string
	drasp	[31]struct {
		begin	time.Time
		end		time.Time
	}
}

func readXLSX(schedule string) []dejurniy {
	var (
		dejurs []dejurniy
		dej dejurniy
		month,year string
		MONTH = map[string]string{
			"Январь": "1",
			"Февраль": "2",
			"Март": "3",
			"Апрель": "4",
			"Май": "5",
			"Июнь": "6",
			"Июль": "7",
			"Август": "8",
			"Сентябрь": "9",
			"Октябрь": "10",
			"Ноябрь": "11",
			"Декабрь": "12",
		}
	)

	f, err := excelize.OpenFile(schedule)
	if err != nil {
		fmt.Println(err)
		return []dejurniy{}
	}

	rows, err := f.GetRows("TDSheet")
	if err != nil {
		fmt.Println(err)
		return []dejurniy{}
	}

	for i, row := range rows {
		for j, colCell := range row {
			// Extraction Month
			if i == 4 && j == 17 {
				month = MONTH[colCell]
			}

			// Extraction Year
			if i == 4 && j == 21 {
				year = colCell
			}

			// Extraction Name Dejurniy
			if i >= 12 && i % 4 == 0 && j == 1 && i <= len(rows)-2 {
				dej.name = colCell
			}

			// Extraction Time Duty
			if j >= 4 && i >= 12 && i <= len(rows)-2 && strings.Contains(colCell, ":") {
				if i%2 == 0 {
					beginDate := strconv.Itoa(j-3)+" "+month+" "+year+" "+colCell+" (MSK)"
					dej.drasp[j-4].begin, _ = time.Parse(timeTempl, beginDate)
				} else {
					if colCell == "24:00" {
						colCell =  "23:59"
					}
					endDate := strconv.Itoa(j-3)+" "+month+" "+year+" "+colCell+" (MSK)"
					dej.drasp[j-4].end, _ = time.Parse(timeTempl, endDate)
				}
			}

			// Add dej to array dejurs
			if i >= 12 && (i+1) % 4 == 0 && j == len(rows[12])-1 {
				dejurs = append(dejurs, dej)
				dej = dejurniy{}
			}
		}
	}
	return dejurs
}

func getListDuty(dejurs []dejurniy) []string {
	var listDuty []string
	for _, d := range dejurs {
		listDuty = append(listDuty, d.name)
	}
	return listDuty
}

func whoDuty(dat time.Time, dejurs []dejurniy) string {
	var name string

	for _, d := range dejurs {
		for _, rsp := range d.drasp {
			if (dat.After(rsp.begin) || dat.Equal(rsp.begin)) && (dat.Before(rsp.end) || dat.Equal(rsp.end)) {
				name = d.name
			}
		}
	}
	return name
}

func allSchedule(name string, dejurs[]dejurniy) [31]string {
	var schedules [31]string
	for _, d := range dejurs {
		if d.name == name {
			for i, rsp := range d.drasp {
				if rsp.begin.IsZero() != true {
					switch rsp.begin.Format("15:04") {
					case "08:00":
						schedules[i] = "Day"
					case "20:00":
						schedules[i] = "Night"
					case "00:00":
						schedules[i] = "Morning"
					default:
						schedules[i] = "No"
					}
				}
			}
		}
	}
	return schedules
}

func telegramBot(dej1, dej2 []dejurniy) {
	//Создаем бота
	bot, err := tgbotapi.NewBotAPI("524283381:AAEAawm4tlOjjWgR_hLx2W4fnsqFvX11XhY")
	if err != nil {
		panic(err)
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			text := update.Message.Text
			//Проверяем что от пользователья пришло именно текстовое сообщение
			if reflect.TypeOf(text).Kind() == reflect.String && text != "" {
				switch text {
				case "/start":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ассалам алейкум! Я скажу тебе кто сейчас на смене!")
					msg.ReplyMarkup = kbrd.MenuLevel1
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "Кто сейчас на смене?":
					today := time.Now().Local()
					nameDuty1 := whoDuty(today, dej1)
					pht1 := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty1))
					pht1.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameDuty1, dej1))
					pht1.Caption = nameDuty1 + " - Дежурный ООЭ АСУ"
					if _, err := bot.Send(pht1); err != nil {
						log.Panic(err)
					}
					nameDuty2 := whoDuty(today, dej2)
					pht2 := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty2))
					pht2.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameDuty2, dej2))
					pht2.Caption = nameDuty2 + " - Дежурный ОИХО"
					if _, err := bot.Send(pht2); err != nil {
						log.Panic(err)
					}
				case "Дежурные":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Дежурные")
					msg.ReplyMarkup = kbrd.MenuLevel12
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "ООЭ АСУ":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ООЭ АСУ")
					msg.ReplyMarkup = kbrd.GetListDuty(getListDuty(dej1))
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "ОИХО":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "ОИХО")
					msg.ReplyMarkup = kbrd.GetListDuty(getListDuty(dej2))
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "Календарь":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Календарь")
					msg.ReplyMarkup = kbrd.CalendarKeyboardMaker()
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "<- Назад":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Календарь")
					msg.ReplyMarkup = kbrd.CalendarKeyboardMaker()
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "< Назад":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Дежурные")
					msg.ReplyMarkup = kbrd.MenuLevel12
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "<  Назад":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
					msg.ReplyMarkup = kbrd.MenuLevel1
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				default:
					if fncs.StrInArray(getListDuty(dej1), text) {
						pht := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(text))
						pht.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(text, dej1))
						pht.Caption = text + " - Дежурный ООЭ АСУ"
						if _, err := bot.Send(pht); err != nil {
							log.Panic(err)
						}
					} else if fncs.StrInArray(getListDuty(dej2), text) {
						pht := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(text))
						pht.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(text, dej2))
						pht.Caption = text + " - Дежурный ОИХО"
						if _, err := bot.Send(pht); err != nil {
							log.Panic(err)
						}
					} else if fncs.IfStrDay(text) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "День/Ночь")
						msg.ReplyMarkup = kbrd.GetMenuDayNight(strings.Trim(text, "-"))
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					} else if str := text; (len(text) >= 19) && (str[len(str)-19:len(str)] == "Дневная 🌝") {
						selDay := strings.Trim(text, " Дневная 🌝")
						strDate := selDay + time.Now().Local().Format(" 1 2006 ") + "15:00"
						calDate, _ := time.Parse(timeTempl, strDate+" (MSK)")
						nameDuty1 := whoDuty(calDate, dej1)
						pht1 := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty1))
						pht1.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameDuty1, dej1))
						pht1.Caption = nameDuty1 + " - Дежурный ООЭ АСУ"
						if _, err := bot.Send(pht1); err != nil {
							log.Panic(err)
						}
						nameDuty2 := whoDuty(calDate, dej2)
						pht2 := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty2))
						pht2.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameDuty2, dej2))
						pht2.Caption = nameDuty2 + " - Дежурный ОИХО"
						if _, err := bot.Send(pht2); err != nil {
							log.Panic(err)
						}
					} else
					if str := text; (len(text) >= 19) && (str[len(str)-17:len(str)] == "Ночная 🌚") {
						selDay := strings.Trim(text, " Ночная 🌚")
						fmt.Println(selDay)
						strDate := selDay + time.Now().Local().Format(" 1 2006 ") + "22:00"
						fmt.Println(strDate)
						calDate, _ := time.Parse(timeTempl, strDate+" (MSK)")
						nameDuty1 := whoDuty(calDate, dej1)
						pht1 := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty1))
						pht1.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameDuty1, dej1))
						pht1.Caption = nameDuty1 + " - Дежурный ООЭ АСУ"
						if _, err := bot.Send(pht1); err != nil {
							log.Panic(err)
						}
						nameDuty2 := whoDuty(calDate, dej2)
						pht2 := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty2))
						pht2.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameDuty2, dej2))
						pht2.Caption = nameDuty2 + " - Дежурный ОИХО"
						if _, err := bot.Send(pht2); err != nil {
							log.Panic(err)
						}
					} else
					{
						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fncs.RandomRustam())
						msg.ReplyMarkup = kbrd.MenuLevel1
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					}
				}
			} else {
				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fncs.RandomRustam())
				msg.ReplyMarkup = kbrd.MenuLevel1
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}
		}
	}
}

func main() {
	dej1 := readXLSX("june.xlsx")
	dej2 := readXLSX("june2.xlsx")

	//Вызываем бота
	telegramBot(dej1, dej2)
}
