package main

import (
	dbase "TeleBot/Database"
	"TeleBot/Duty"
	"TeleBot/Telegram"
	"github.com/joho/godotenv"
	"log"
	"net"
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
	cert := os.Getenv("CERT")
	key := os.Getenv("KEY")
	//Слушаем Telegram
	go func() {
		l, err := net.Listen("tcp4", ":8443")
		if err != nil {
			log.Panic(err)
		}
		err = http.ServeTLS(l, nil, cert, key)
		//err := http.ListenAndServeTLS("0.0.0.0:8443", tb.cert, tb.key, nil)
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
