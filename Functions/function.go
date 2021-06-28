package Functions

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func LastDayOfMonth(date time.Time) int {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, 0).Add(time.Nanosecond * -1).Format("2")
	lastDayInt, _ := strconv.Atoi(lastDay)
	return lastDayInt
}

func IfStrDay(strDay string) bool {
	if strDay[0] == '-' && strDay[len(strDay)-1] == '-' {
		num, _ := strconv.Atoi(strings.Trim(strDay, "-"))
		for i := 1; i <= LastDayOfMonth(time.Now().Local()); i++ {
			if i == num {
				return true
			}
		}
	}
	return false
}

func StrInArray(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetPathImg(name string) string {
	var imgfile string
	switch name {
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
	return imgfile
}

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