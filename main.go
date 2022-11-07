package main

import (
	"TeleBot/Duty"
	"TeleBot/Telegram"
	"log"

	"github.com/joho/godotenv"
)

var dej = Duty.Dejurnie{}

func main() {
	// t := time.Now().Local().Format("January")
	// filename, _ := filepath.Glob("./schedules/" + s.ToLower(t) + "*.xlsx")

	// _, err1 := os.Stat(filename[1])
	// if err1 != nil {
	// 	if os.IsNotExist(err1) {
	// 		fmt.Println("file does not exist") // это_true
	// 	} else {
	// 		fmt.Println("another error")
	// 	}
	// } else {
	// 	fmt.Println("file exist")
	// 	fmt.Println(filename)
	// }

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dej.FindXLSX()

	//Вызываем бота
	Telegram.TeleBot(&dej)
}
