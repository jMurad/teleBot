package main

import (
	"TeleBot/Duty"
	"TeleBot/Telegram"
	"github.com/joho/godotenv"
	"log"
)

var dej = Duty.Dejurnie{}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dej.FindXLSX()

	//Вызываем бота
	Telegram.TeleBot(&dej)
}
