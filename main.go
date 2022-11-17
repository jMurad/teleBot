package main

import (
	dbase "TeleBot/Database"
	"TeleBot/Duty"
	"TeleBot/Telegram"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

var dej = Duty.Dejurnie{}
var db = dbase.TGDB{}
var tb = Telegram.TeleBot{}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//host := os.Getenv("HOST")

	//Слушаем Telegram
	go func() {
		cert := os.Getenv("CERT")
		key := os.Getenv("KEY")
		err := http.ListenAndServeTLS(":8443", cert, key, nil)
		if err != nil {
			log.Panic(err)
		}
	}()

	//Инициализация БД и Бота
	db.DBinit()
	tb.TBInit()
	dej.DutyInit()

	newdej := make(chan Duty.Dejurnie)
	go dej.CronXLSX(newdej)

	//Запускаем бота
	tb.RunBot(&dej, newdej, &db)
}
