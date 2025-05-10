package main

import (
	"fmt"
	"log"
	functions "schedule/snschedule_bot"
	"strings"
	"sync"
	"time"
)

func main() {
	log.Println("Starting Schedule Bot...")

	var wg sync.WaitGroup
	wg.Add(1) // Добавляем счетчик для бота

	// Если нужно инициализировать БД, делаем это в основной горутине
	if shouldInitializeDatabase() {
		initializeDatabase()
	}

	// Запускаем бота в отдельной горутине
	go func() {
		defer wg.Done()
		startBot()
	}()

	// Ждем завершения работы бота
	wg.Wait()
}

// shouldInitializeDatabase запрашивает у пользователя необходимость инициализации БД
func shouldInitializeDatabase() bool {
	var answer string
	fmt.Println("ЗАПУСКАТЬ ПОЛНУЮ ЗАГРУЗКУ БД? (БАЗА ДАННЫХ ДОЛЖНА БЫТЬ ПУСТАЯ){КОНЕЧНО/СКИП}")
	fmt.Scan(&answer)
	return strings.ToLower(answer) == "конечно"
}

// initializeDatabase выполняет полную инициализацию базы данных
func initializeDatabase() {
	log.Println("ПОЛНАЯ ЗАГРУЗКА БД (ПЕРЕД ЗАПУСКОМ ОЖИДАНИЕ 10 СЕКУНД)")
	time.Sleep(10 * time.Second)

	functions.FunctionDbWriter()

	log.Println("База данных успешно инициализирована")
}

// startBot запускает телеграм бота
func startBot() {
	log.Println("Запуск телеграм бота...")

	functions.NewBot()

	log.Println("Бот успешно запущен")
}
