package main

import (
	"fmt"
	functions "schedule/snschedule_bot"
	"strings"
	"time"
)

func main() {
	fmt.Println("func: main.go")

	var answer string
	fmt.Println("ЗАПУСКАТЬ ПОЛНУЮ ЗАГРУЗКУ БД? (БАЗА ДАННЫХ ДОЛЖНА БЫТЬ ПУСТАЯ){КОНЕЧНО/СКИП}")
	fmt.Scan(&answer)
	if strings.ToLower(answer) == "конечно" {
		fmt.Println("ПОЛНАЯ ЗАГРУЗКА БД (ПЕРЕД ЗАПУСКОМ ОЖИДАНИЕ 10 СЕКУНД)")
		time.Sleep(10 * time.Second)
		functions.FunctionDbWriter()
	}
	functions.NewBot()
}
