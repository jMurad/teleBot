package main

import (
	dbase "TeleBot/Database"
	"TeleBot/Duty"
	"TeleBot/Telegram"
	"log"

	"github.com/joho/godotenv"
)

var dej = Duty.Dejurnie{}
var db = dbase.TGDB{}
var tb = Telegram.TeleBot{}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//Инициализация БД и Бота
	db.DBinit()
	tb.TBInit()

	dej.FindXLSX()
	newdej := make(chan Duty.Dejurnie)
	go dej.CronXLSX(newdej)

	//Запускаем бота
	tb.RunBot(&dej, newdej, &db)
}
