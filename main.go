package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	kbrd "TeleBot/Keyboards"
	"math/rand"

	//"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type dejurniy struct {
	name	string
	drasp	[31]struct {
		begin	time.Time
		end		time.Time
	}
}

const timeTempl = "2 1 2006 15:04 (MST)"

func readXLSX(schedule string) []dejurniy {
	f, err := excelize.OpenFile(schedule)
	if err != nil {
		fmt.Println(err)
		return []dejurniy{}
	}

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("TDSheet")
	if err != nil {
		fmt.Println(err)
		return []dejurniy{}
	}

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
					dej.drasp[j-4].begin, _ = time.Parse(timeTempl, strconv.Itoa(j-3)+" "+month+" "+year+" "+colCell+" (MSK)")
				} else {
					if colCell == "24:00" {
						colCell =  "23:59"
					}
					dej.drasp[j-4].end, _ = time.Parse(timeTempl, strconv.Itoa(j-3)+" "+month+" "+year+" "+colCell+" (MSK)")
				}
			}
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
	var smens [31]string
	for _, d := range dejurs {
		if d.name == name {
			for i, rsp := range d.drasp {
				if rsp.begin.IsZero() != true {
					switch rsp.begin.Format("15:04") {
					case "08:00":
						smens[i] = "Day"
					case "20:00":
						smens[i] = "Night"
					default:
						smens[i] = "No"
					}
				}
			}
		}
	}
	return smens
}

func randomRustam() string {
	//rand.Seed(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())
	rnd := rand.Intn(10)
	var msg = [10]string{
		"Не понял!",
		"Кто здесь!?",
		"А если я встану?",
		"А глаз че говорит!?",
		"А шея!?",
		"А шляпа с большими полями?!",
		"Чтооооо!?",
		"Как ты мне надоел!!!",
		"Вахтанг!?",
		"Скунс!?",
	}
	return msg[rnd]
}

func containsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsInt(e string) bool {
	if e[0] == '-' && e[len(e)-1] == '-'{
		num, _ := strconv.Atoi(strings.Trim(e, "-"))
		for i := 1; i <= 31; i++ {
			if i == num {
				return true
			}
		}
	}
	return false
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
			//Проверяем что от пользователья пришло именно текстовое сообщение
			if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
				switch update.Message.Text {
				case "/start":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ассалам алейкум! Я скажу тебе кто сейчас на смене!")
					msg.ReplyMarkup = kbrd.MenuLevel_1
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "Кто сейчас на смене?":
					today := time.Now().Local()
					fmt.Println(today)
					nameduty1 := whoDuty(today, dej1)
					nameduty2 := whoDuty(today, dej2)
					imgfile := ""

					switch nameduty1 {
					case "Велиханов А.В.":
						imgfile = "photo/VAV.jpg"
					case "Нурмагомедов Р.М.":
						imgfile = "photo/NRM.jpg"
					case "Сулейманов И.А.":
						imgfile = "photo/SIA.jpg"
					case "Сулейманов Ш.А.":
						imgfile = "photo/SSA.jpg"
					case "Яхьяев М.Л.":
						imgfile = "photo/YML.jpg"
					case "Магомедрасулов М.Б":
						imgfile = "photo/MMB.jpg"
					}
					pht1 := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
					pht1.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameduty1, dej1))
					pht1.Caption = nameduty1 + " - Дежурный ООЭ АСУ"
					if _, err := bot.Send(pht1); err != nil {
						log.Panic(err)
					}

					switch nameduty2 {
					case "Абдуллаев М.М.":
						imgfile = "photo/AMM.jpg"
					case "Газиев Г.М.":
						imgfile = "photo/GGM.jpg"
					case "Идрисов М.А.":
						imgfile = "photo/IMA.jpg"
					case "Кузнецов Д.В.":
						imgfile = "photo/KDV.jpg"
					case "Шихвеледов Р.Ш.":
						imgfile = "photo/SRS.jpg"
					}
					pht2 := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
					pht2.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameduty2, dej2))
					pht2.Caption = nameduty2 + " - Дежурный ОИХО"
					if _, err := bot.Send(pht2); err != nil {
						log.Panic(err)
					}
				case "Дежурные":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Дежурные")
					msg.ReplyMarkup = kbrd.MenuLevel_1_2
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
					msg.ReplyMarkup = kbrd.MenuLevel_1_2
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "<  Назад":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
					msg.ReplyMarkup = kbrd.MenuLevel_1
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				default:
					if containsStr(getListDuty(dej1), update.Message.Text) {
						imgfile := ""
						switch update.Message.Text {
						case "Велиханов А.В.":
							imgfile = "photo/VAV.jpg"
						case "Нурмагомедов Р.М.":
							imgfile = "photo/NRM.jpg"
						case "Сулейманов И.А.":
							imgfile = "photo/SIA.jpg"
						case "Сулейманов Ш.А.":
							imgfile = "photo/SSA.jpg"
						case "Яхьяев М.Л.":
							imgfile = "photo/YML.jpg"
						case "Магомедрасулов М.Б":
							imgfile = "photo/MMB.jpg"
						}
						pht := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
						pht.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(update.Message.Text, dej1))
						pht.Caption = update.Message.Text + " - Дежурный ООЭ АСУ"
						if _, err := bot.Send(pht); err != nil {
							log.Panic(err)
						}
					} else
					if containsStr(getListDuty(dej2), update.Message.Text) {
						imgfile := ""
						switch update.Message.Text {
						case "Абдуллаев М.М.":
							imgfile = "photo/AMM.jpg"
						case "Газиев Г.М.":
							imgfile = "photo/GGM.jpg"
						case "Идрисов М.А.":
							imgfile = "photo/IMA.jpg"
						case "Кузнецов Д.В.":
							imgfile = "photo/KDV.jpg"
						case "Шихвеледов Р.Ш.":
							imgfile = "photo/SRS.jpg"
						}
						pht := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
						pht.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(update.Message.Text, dej2))
						pht.Caption = update.Message.Text + " - Дежурный ОИХО"
						if _, err := bot.Send(pht); err != nil {
							log.Panic(err)
						}
					} else
					if containsInt(update.Message.Text) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "День/Ночь")
						msg.ReplyMarkup = kbrd.GetMenuDayNight(strings.Trim(update.Message.Text, "-"))
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					} else
					if str := update.Message.Text; (len(update.Message.Text) >= 19) && (str[len(str)-19:len(str)] == "Дневная 🌝") {
						selDay := strings.Trim(update.Message.Text, " Дневная 🌝")
						//fmt.Println(selDay)
						strDate := selDay + time.Now().Local().Format(" 1 2006 ") + "15:00"
						//fmt.Println(strDate)
						calDate, _ := time.Parse(timeTempl, strDate+" (MSK)")
						fmt.Println(calDate)
						nameduty1 := whoDuty(calDate, dej1)
						nameduty2 := whoDuty(calDate, dej2)
						imgfile := ""

						switch nameduty1 {
						case "Велиханов А.В.":
							imgfile = "photo/VAV.jpg"
						case "Нурмагомедов Р.М.":
							imgfile = "photo/NRM.jpg"
						case "Сулейманов И.А.":
							imgfile = "photo/SIA.jpg"
						case "Сулейманов Ш.А.":
							imgfile = "photo/SSA.jpg"
						case "Яхьяев М.Л.":
							imgfile = "photo/YML.jpg"
						case "Магомедрасулов М.Б":
							imgfile = "photo/MMB.jpg"
						}
						pht1 := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
						pht1.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameduty1, dej1))
						pht1.Caption = nameduty1 + " - Дежурный ООЭ АСУ"
						if _, err := bot.Send(pht1); err != nil {
							log.Panic(err)
						}

						switch nameduty2 {
						case "Абдуллаев М.М.":
							imgfile = "photo/AMM.jpg"
						case "Газиев Г.М.":
							imgfile = "photo/GGM.jpg"
						case "Идрисов М.А.":
							imgfile = "photo/IMA.jpg"
						case "Кузнецов Д.В.":
							imgfile = "photo/KDV.jpg"
						case "Шихвеледов Р.Ш.":
							imgfile = "photo/SRS.jpg"
						}
						pht2 := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
						pht2.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameduty2, dej2))
						pht2.Caption = nameduty2 + " - Дежурный ОИХО"
						if _, err := bot.Send(pht2); err != nil {
							log.Panic(err)
						}
					} else
					if str := update.Message.Text; (len(update.Message.Text) >= 19) && (str[len(str)-17:len(str)] == "Ночная 🌚") {
						selDay := strings.Trim(update.Message.Text, " Ночная 🌚")
						fmt.Println(selDay)
						strDate := selDay + time.Now().Local().Format(" 1 2006 ") + "22:00"
						fmt.Println(strDate)
						calDate, _ := time.Parse(timeTempl, strDate+" (MSK)")
						fmt.Println(calDate)
						nameduty1 := whoDuty(calDate, dej1)
						nameduty2 := whoDuty(calDate, dej2)
						imgfile := ""

						switch nameduty1 {
						case "Велиханов А.В.":
							imgfile = "photo/VAV.jpg"
						case "Нурмагомедов Р.М.":
							imgfile = "photo/NRM.jpg"
						case "Сулейманов И.А.":
							imgfile = "photo/SIA.jpg"
						case "Сулейманов Ш.А.":
							imgfile = "photo/SSA.jpg"
						case "Яхьяев М.Л.":
							imgfile = "photo/YML.jpg"
						case "Магомедрасулов М.Б":
							imgfile = "photo/MMB.jpg"
						}
						pht1 := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
						pht1.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameduty1, dej1))
						pht1.Caption = nameduty1 + " - Дежурный ООЭ АСУ"
						if _, err := bot.Send(pht1); err != nil {
							log.Panic(err)
						}

						switch nameduty2 {
						case "Абдуллаев М.М.":
							imgfile = "photo/AMM.jpg"
						case "Газиев Г.М.":
							imgfile = "photo/GGM.jpg"
						case "Идрисов М.А.":
							imgfile = "photo/IMA.jpg"
						case "Кузнецов Д.В.":
							imgfile = "photo/KDV.jpg"
						case "Шихвеледов Р.Ш.":
							imgfile = "photo/SRS.jpg"
						}
						pht2 := tgbotapi.NewPhoto(update.Message.Chat.ID, imgfile)
						pht2.ReplyMarkup = kbrd.InlineKeyboardMaker(allSchedule(nameduty2, dej2))
						pht2.Caption = nameduty2 + " - Дежурный ОИХО"
						if _, err := bot.Send(pht2); err != nil {
							log.Panic(err)
						}
					} else
					{
						randomRustam()
						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, randomRustam())
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					}
				}
			} else {

				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, randomRustam())
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
