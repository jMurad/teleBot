package main

import (
	fncs "TeleBot/Functions"
	kbrd "TeleBot/Keyboards"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	s "strings"
	"time"
)

const timeTempl = "2 1 2006 15:04 (MST)"

type dejurnie struct {
	deptName	[]string
	dept		[]department
}

type department struct {
	dutyName	[]string
	drasp		[]rasp
}

type rasp [31]struct {
	begin	time.Time
	end		time.Time
}

func (d *dejurnie) readXLSX() {
	var (
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

	// Ищем все XLSX файлы с названием текущего месяца
	t := time.Now().Local().Format("January")
	files, err := filepath.Glob(s.ToLower(t)+"*.xlsx")
	if err != nil {
		panic(err)
	}

	// Проходимся по всем файлам XLSX
	for num, fpath := range files {
		d.deptName = append(d.deptName, "")
		d.dept = append(d.dept, department{})

		// Открываем файл XLSX
		f, err := excelize.OpenFile(fpath)
		if err != nil {
			panic(err)
		}

		// Извлекаем все строки страницы TDSheet
		rows, err := f.GetRows("TDSheet")
		if err != nil {
			panic(err)
		}

		// Временная переменная для хранения расписания дежурного
		raspDuty := rasp{}

		// Проходимся по всем строкам страницы TDSheet
		for i, row := range rows {
			for j, colCell := range row {
				// Извлекаем название Месяца
				if i == 4 && j == 17 {
					month = MONTH[colCell]
				}

				// Извлекаем Год
				if i == 4 && j == 21 {
					year = colCell
				}

				// Извлекаем название Отдела
				if i == 4 && j == 5 {
					d.deptName[num] = fncs.TripDept(colCell)
				}

				// Извлекаем Имя дежурного
				if i >= 12 && i % 4 == 0 && j == 1 && i <= len(rows)-2 {
					d.dept[num].dutyName = append(d.dept[num].dutyName, colCell)
				}

				// Извлекаем Время начала и конца смены
				if j >= 4 && i >= 12 && i <= len(rows)-2 && s.Contains(colCell, ":") {
					if i%2 == 0 {
						beginDate := strconv.Itoa(j-3)+" "+month+" "+year+" "+colCell+" (MSK)"
						raspDuty[j-4].begin, _ = time.Parse(timeTempl, beginDate)
					} else {
						if colCell == "24:00" {
							colCell =  "23:59"
						}
						endDate := strconv.Itoa(j-3)+" "+month+" "+year+" "+colCell+" (MSK)"
						raspDuty[j-4].end, _ = time.Parse(timeTempl, endDate)
					}
				}

				// Добавляем дежурного в список
				if i >= 12 && (i+1) % 4 == 0 && j == len(rows[12])-1 {
					d.dept[num].drasp = append(d.dept[num].drasp, raspDuty)
					raspDuty = rasp{}
				}
			}
		}
	}
}

func (d *dejurnie) getListDept() []string {
	var listDept []string
	for _, dn := range d.deptName {
		listDept = append(listDept, dn)
	}
	return listDept
}

func (d *dejurnie) getListDuty(deptName string) []string {
	var listDuty []string
	_, num := fncs.StrInArray(d.deptName, deptName)

	for _, dn := range d.dept[num].dutyName {
		listDuty = append(listDuty, dn)
	}
	return listDuty
}

func (d *dejurnie) getListDutyAll() []string {
	var listDuty []string

	for _, deptName := range d.deptName {
		listDuty = append(listDuty, d.getListDuty(deptName)...)
	}
	return listDuty
}

func (d *dejurnie) whoDuty(date time.Time, deptName string) string {
	_, num := fncs.StrInArray(d.deptName, deptName)

	for i, dr := range d.dept[num].drasp {
		for _, rsp := range dr {
			if (date.After(rsp.begin) || date.Equal(rsp.begin)) && (date.Before(rsp.end) || date.Equal(rsp.end)) {
				result := d.dept[num].dutyName[i]
				return result
			}
		}
	}
	panic("I don't know")
	//return result
}

func (d *dejurnie) whoDutyAll(date time.Time) []string {
	var listDuty []string

	for _, dept := range d.deptName {
		listDuty = append(listDuty, d.whoDuty(date, dept))
	}
	return listDuty
}

func (d *dejurnie) dutyToDept(dutyName string) string {
	for i, dept := range d.dept {
		for _, dn := range dept.dutyName {
			if dn == dutyName {
				return d.deptName[i]
			}
		}
	}
	panic("I don't know")
}

func (d *dejurnie) getSchedule(dutyName string) [31]string {
	var schedules [31]string
	for _, dept := range d.dept {
		for i, dn := range dept.dutyName {
			if dn == dutyName {
				for j, rsp := range dept.drasp[i] {
					if rsp.begin.IsZero() != true {
						switch rsp.begin.Format("15:04") {
						case "08:00":
							schedules[j] = "Day"
						case "20:00":
							schedules[j] = "Night"
						case "00:00":
							schedules[j] = "Morning"
						default:
							schedules[j] = "No"
						}
					}
				}
			}
		}
	}
	return schedules
}

func telegramBot(dej dejurnie, token string) {
	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	listDept := dej.getListDept()

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
					msg.ReplyMarkup = kbrd.MainMenu
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "Кто сейчас на смене?":
					today := time.Now().Local()

					for _, nameDuty := range dej.whoDutyAll(today) {
						pht := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty))
						pht.ReplyMarkup = kbrd.InlineKeyboardMaker(dej.getSchedule(nameDuty))
						pht.Caption = nameDuty + " - Дежурный " + dej.dutyToDept(nameDuty)
						if _, err := bot.Send(pht); err != nil {
							log.Panic(err)
						}
					}
				case "Дежурные":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Дежурные")
					msg.ReplyMarkup = kbrd.GetListDept(listDept)
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
					msg.ReplyMarkup = kbrd.GetListDept(listDept)
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				case "<  Назад":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Меню")
					msg.ReplyMarkup = kbrd.MainMenu
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
				default:
					if inDuty, _ := fncs.StrInArray(dej.getListDutyAll(), text); inDuty {
						pht := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(text))
						pht.ReplyMarkup = kbrd.InlineKeyboardMaker(dej.getSchedule(text))
						pht.Caption = text + " - Дежурный " + dej.dutyToDept(text)
						if _, err := bot.Send(pht); err != nil {
							log.Panic(err)
						}
					} else
					if inDept, _ := fncs.StrInArray(listDept, text); inDept {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
						msg.ReplyMarkup = kbrd.GetListDuty(dej.getListDuty(text))
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					} else
					if fncs.IfStrDay(text) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "День/Ночь")
						msg.ReplyMarkup = kbrd.GetMenuDayNight(s.Trim(text, "-"))
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					} else
					if s.HasSuffix(text, "Дневная 🌝") || s.HasSuffix(text, "Ночная 🌚") {
						selDate := time.Time{}
						if s.HasSuffix(text, "Дневная 🌝") {
							selDay := s.Trim(text, " Дневная 🌝")
							strDate := selDay + time.Now().Local().Format(" 1 2006 ") + "15:00"
							selDate, _ = time.Parse(timeTempl, strDate+" (MSK)")
						} else {
							selDay := s.Trim(text, " Ночная 🌚")
							strDate := selDay + time.Now().Local().Format(" 1 2006 ") + "22:00"
							selDate, _ = time.Parse(timeTempl, strDate+" (MSK)")
						}

						for _, nameDuty := range dej.whoDutyAll(selDate) {
							pht := tgbotapi.NewPhoto(update.Message.Chat.ID, fncs.GetPathImg(nameDuty))
							pht.ReplyMarkup = kbrd.InlineKeyboardMaker(dej.getSchedule(nameDuty))
							pht.Caption = nameDuty + " - Дежурный " + dej.dutyToDept(nameDuty)
							if _, err := bot.Send(pht); err != nil {
								log.Panic(err)
							}
						}
					} else
					{
						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fncs.RandomRustam())
						msg.ReplyMarkup = kbrd.MainMenu
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					}
				}
			} else {
				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fncs.RandomRustam())
				msg.ReplyMarkup = kbrd.MainMenu
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
	dej := dejurnie{}
	dej.readXLSX()
	token := fncs.GetAPIToken()
	//Вызываем бота
	telegramBot(dej, token)
}
