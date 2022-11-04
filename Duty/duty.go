package Duty

import (
	fncs "TeleBot/Functions"
	// "encoding/json"
	// "fmt"
	"path/filepath"
	"strconv"
	s "strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

type Dejurnie struct {
	DeptNames []string
	Depts     []department
}

type department struct {
	DutyNames []string
	Drasp     []rasp
}

type rasp [31]struct {
	Begin time.Time
	End   time.Time
}

var RunningParse bool = false

const timeTempl = "2 1 2006 15:04 (MST)"

func (d *Dejurnie) FindXLSX() {
	// Ищем все XLSX файлы с названием текущего месяца
	t := time.Now().Local().Format("January")
	files, err := filepath.Glob(s.ToLower(t) + "*.xlsx")
	if err != nil {
		panic(err)
	}

	// Проходимся по всем файлам XLSX
	for num, fpath := range files {
		d.ParseXLSX(fpath, num)
	}

	// b, err := json.Marshal(d)
	// if err != nil {
	// 	fmt.Printf("Error: %s", err)
	// 	return
	// }
	// fmt.Println("|" + string(b) + "|")
}

func (d *Dejurnie) ParseXLSX(fpath string, num int) {
	var month, year string

	d.DeptNames = append(d.DeptNames, "")
	d.Depts = append(d.Depts, department{})

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

	beginDate, endDate := "", ""
	month = time.Now().Local().Format("1")
	year = time.Now().Local().Format("2006")
	d.DeptNames[num] = fncs.TripDept(rows[4][5])
	for ii := 12; ii < len(rows)-2; ii += 4 {
		d.Depts[num].DutyNames = append(d.Depts[num].DutyNames, rows[ii][1])
		for jj := 4; jj <= 34; jj++ {
			if s.Contains(rows[ii][jj], ":") {
				beginDate = strconv.Itoa(jj-3) + " " + month + " " + year + " " + rows[ii][jj] + " (MSK)"
				raspDuty[jj-4].Begin, _ = time.Parse(timeTempl, beginDate)
			} else if s.Contains(rows[ii+2][jj], ":") {
				beginDate = strconv.Itoa(jj-3) + " " + month + " " + year + " " + rows[ii+2][jj] + " (MSK)"
				raspDuty[jj-4].Begin, _ = time.Parse(timeTempl, beginDate)
			}
			if s.Contains(rows[ii+1][jj], ":") {
				endDate = strconv.Itoa(jj-3) + " " + month + " " + year + " " + rows[ii+1][jj] + " (MSK)"
				raspDuty[jj-4].End, _ = time.Parse(timeTempl, endDate)
			} else if s.Contains(rows[ii+3][jj], ":") {
				endDate = strconv.Itoa(jj-3) + " " + month + " " + year + " " + rows[ii+3][jj] + " (MSK)"
				raspDuty[jj-4].End, _ = time.Parse(timeTempl, endDate)
			}
		}

		d.Depts[num].Drasp = append(d.Depts[num].Drasp, raspDuty)
		raspDuty = rasp{}

	}

	// Проходимся по всем строкам страницы TDSheet
	// for i, row := range rows {
	// 	// Проходимся по всем столбцам строки
	// 	for j, colCell := range row {
	// 		// Извлекаем название Месяца
	// 		if i == 4 && j == 17 {
	// 			month = MONTH[colCell]
	// 		}

	// 		// Извлекаем Год
	// 		if i == 4 && j == 21 {
	// 			year = colCell
	// 		}

	// 		// Извлекаем название Отдела
	// 		if i == 4 && j == 5 {
	// 			d.DeptNames[num] = fncs.TripDept(colCell)
	// 		}

	// 		// Извлекаем Имя дежурного
	// 		if i >= 12 && i%4 == 0 && j == 1 && i <= len(rows)-2 {
	// 			d.Depts[num].DutyNames = append(d.Depts[num].DutyNames, colCell)
	// 		}

	// 		// Извлекаем Время начала и конца смены
	// 		if j >= 4 && i >= 12 && i <= len(rows)-2 && s.Contains(colCell, ":") {
	// 			if i%2 == 0 {
	// 				beginDate := strconv.Itoa(j-3) + " " + month + " " + year + " " + colCell + " (MSK)"
	// 				raspDuty[j-4].Begin, _ = time.Parse(timeTempl, beginDate)
	// 			} else {
	// 				if colCell == "24:00" {
	// 					colCell = "23:59"
	// 				}
	// 				endDate := strconv.Itoa(j-3) + " " + month + " " + year + " " + colCell + " (MSK)"
	// 				raspDuty[j-4].End, _ = time.Parse(timeTempl, endDate)
	// 			}
	// 		}

	// 		// Добавляем дежурного в список
	// 		if i >= 12 && (i+1)%4 == 0 && j == len(rows[12])-1 {
	// 			d.Depts[num].Drasp = append(d.Depts[num].Drasp, raspDuty)
	// 			raspDuty = rasp{}
	// 		}
	// 	}
	// }
}

func (d *Dejurnie) CronXLSX(flag chan Dejurnie) {
	c := time.Tick(60 * time.Minute)
	for range c {
		RunningParse = true
		// Инициализация переменных пустыми значениями
		d.DeptNames = []string{}
		d.Depts = []department{}
		d.FindXLSX()
		RunningParse = false
		flag <- *d
	}
}

func (d *Dejurnie) GetSchedule(dutyName string) [31]string {
	var schedules [31]string
	for _, dept := range d.Depts {
		for i, dn := range dept.DutyNames {
			if dn == dutyName {
				for j, rsp := range dept.Drasp[i] {
					if !rsp.Begin.IsZero() {
						switch rsp.Begin.Format("15:04") {
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
				return schedules
			}
		}
	}
	return schedules
}

func (d *Dejurnie) whoDuty(date time.Time, deptName string) string {
	num := fncs.StrInArray(d.DeptNames, deptName)
	if num < 0 {
		return ""
	}
	for i, dr := range d.Depts[num].Drasp {
		for _, rsp := range dr {
			if (date.After(rsp.Begin) || date.Equal(rsp.Begin)) && (date.Before(rsp.End) || date.Equal(rsp.End)) {
				return d.Depts[num].DutyNames[i]
			}
		}
	}
	return ""
}

func (d *Dejurnie) GetListDept() []string {
	var listDept []string
	listDept = append(listDept, d.DeptNames...)
	return listDept
}

func (d *Dejurnie) GetListDuty(deptName string) []string {
	var listDuty []string
	num := fncs.StrInArray(d.DeptNames, deptName)

	if num == -1 {
		return listDuty
	}

	listDuty = append(listDuty, d.Depts[num].DutyNames...)
	return listDuty
}

func (d *Dejurnie) GetListDutyAll() []string {
	var listDuty []string

	for _, deptName := range d.DeptNames {
		listDuty = append(listDuty, d.GetListDuty(deptName)...)
	}
	return listDuty
}

func (d *Dejurnie) WhoDutyAll(date time.Time) []string {
	var listDuty []string

	for _, dept := range d.DeptNames {
		listDuty = append(listDuty, d.whoDuty(date, dept))
	}
	return listDuty
}

func (d *Dejurnie) DutyToDept(dutyName string) string {
	for i, dept := range d.Depts {
		for _, dn := range dept.DutyNames {
			if dn == dutyName {
				return d.DeptNames[i]
			}
		}
	}
	return ""
}
