package Functions

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// LastDayOfMonth Число последнего дня месяца или сколько в месяце дней
func LastDayOfMonth(date time.Time) int {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, 0).Add(time.Nanosecond * -1).Format("2")
	lastDayInt, _ := strconv.Atoi(lastDay)
	return lastDayInt
}

// IfStrDay Находится ли число в пределах месяца от 1 до 30 или 31
func IfStrDay(strDay string) bool {
	num, _ := strconv.Atoi(strings.Trim(strDay, "-"))
	if 1 <= num && num <= LastDayOfMonth(time.Now().Local()) {
		return true
	} else {
		return false
	}
}

// StrInArray Есть ли строка в массиве, если да то вывести порядковый номер
func StrInArray(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}

// RandomRustam Выдает случайные фразы Рустама
func RandomRustam() string {
	//rand.Seed(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())
	rnd := rand.Intn(10)
	var msg = [20]string{
		"Как ты мне надоел!!!",
		"Не понял!",
		"Кто здесь!?",
		"А если я встану?",
		"А глаз че говорит!?",
		"А шея!?",
		"Шляпа с большими полями?!",
		"Чтооооо!?",
		"Как ты мне надоел!!!",
		"Вахтанг!?",
		"Скунс!?",
		"Щииииб",
		"Смирно!",
		"Как ты мне надоел!!!",
		"Не позволю!",
		"Наливай тогда...",
		"Вооот же ооон...",
		"Товарищ Пэрденко?!",
		"Как ты мне надоел!!!",
		"Если я встану по другому же будешь разговаривать!",
	}
	return msg[rnd]
}

// TripDept Составляет абревиатуру из первых букв всех слов
func TripDept(nameDept string) string {
	strNew := ""
	for i, str := range strings.ToUpper(nameDept) {
		if i == 0 {
			strNew += string(str)
		} else if nameDept[i-1] == ' ' {
			strNew += string(str)
		}
	}
	return strNew
}
