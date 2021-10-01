package Duty

import (
	fncs "TeleBot/Functions"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"path/filepath"
	"strconv"
	s "strings"
	"time"
)

type Dejurnie struct {
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

var RunningParse bool = false

const timeTempl = "2 1 2006 15:04 (MST)"

func (d *Dejurnie) whoDuty(date time.Time, deptName string) string {
	num := fncs.StrInArray(d.deptName, deptName)

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

func (d *Dejurnie) addToLog(operation string) {

}

func (d *Dejurnie) FindXLSX() {
	// Ищем все XLSX файлы с названием текущего месяца
	t := time.Now().Local().Format("January")
	files, err := filepath.Glob(s.ToLower(t)+"*.xlsx")
	if err != nil {
		panic(err)
	}

	// Проходимся по всем файлам XLSX
	for num, fpath := range files {
		d.ParseXLSX(fpath, num)
	}
}

func (d *Dejurnie) ParseXLSX(fpath string, num int) {
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

func (d *Dejurnie) GetListDept() []string {
	var listDept []string
	for _, dn := range d.deptName {
		listDept = append(listDept, dn)
	}
	return listDept
}

func (d *Dejurnie) GetListDuty(deptName string) []string {
	var listDuty []string
	num := fncs.StrInArray(d.deptName, deptName)

	for _, dn := range d.dept[num].dutyName {
		listDuty = append(listDuty, dn)
	}
	return listDuty
}

func (d *Dejurnie) GetListDutyAll() []string {
	var listDuty []string

	for _, deptName := range d.deptName {
		listDuty = append(listDuty, d.GetListDuty(deptName)...)
	}
	return listDuty
}

func (d *Dejurnie) WhoDutyAll(date time.Time) []string {
	var listDuty []string

	for _, dept := range d.deptName {
		listDuty = append(listDuty, d.whoDuty(date, dept))
	}
	return listDuty
}

func (d *Dejurnie) DutyToDept(dutyName string) string {
	for i, dept := range d.dept {
		for _, dn := range dept.dutyName {
			if dn == dutyName {
				return d.deptName[i]
			}
		}
	}
	panic("I don't know")
}

func (d *Dejurnie) GetSchedule(dutyName string) [31]string {
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

func (d *Dejurnie) CronXLSX(flag chan Dejurnie)  {
	c := time.Tick(60 * time.Minute)
	for range c {
		RunningParse = true
		// Инициализация переменных пустыми значениями
		d.deptName = []string{}
		d.dept = []department{}
		d.FindXLSX()
		RunningParse = false
		flag <- *d
	}
}