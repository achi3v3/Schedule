/*
Файл содержит функции для управления состояниями пользователей в Telegram боте.
Основные функции:
  - Управление состояниями пользователей (курс, группа, день)
  - Хранение и обработка ID сообщений
  - Работа с текстом сообщений и извлечение ID пользователей

(используется RWMutex для потокобезопасности)
*/
package functions

import (
	"errors"
	"strconv"
	"sync"
)

// Хранение текущего состояния пользователя в диалоге с ботом
type UserState struct {
	Course string // Идентификатор выбранного курса
	Group  string // Идентификатор выбранной группы
	Day    string // Идентификатор выбранного дня
}

// Потокобезопасная структура для хранения состояний пользователей
var userStates = struct {
	sync.RWMutex
	data map[int64]map[string]string
}{data: make(map[int64]map[string]string)}

// Возвращает текущее состояние пользователя по его chatID
func getUserState(chatID int64) map[string]string {
	userStates.RLock()
	defer userStates.RUnlock()

	if state, exists := userStates.data[chatID]; exists {
		return state
	}
	return nil
}

// Устанавливает новое значение состояния для пользователя
func setUserState(chatID int64, key string, value string) {
	userStates.Lock()
	defer userStates.Unlock()

	if _, exists := userStates.data[chatID]; !exists {
		userStates.data[chatID] = make(map[string]string)
	}

	userStates.data[chatID][key] = value
}

// Удаляет указанное состояние пользователя
func deleteUserState(chatID int64, key string) {
	userStates.Lock()
	defer userStates.Unlock()

	if state, exists := userStates.data[chatID]; exists {
		delete(state, key)
		if len(state) == 0 {
			delete(userStates.data, chatID)
		}
	}
}

// Полностью очищает все состояния пользователя
func resetUserState(chatID int64) {
	userStates.Lock()
	defer userStates.Unlock()
	delete(userStates.data, chatID)
}

// Потокобезопасная структура для хранения ID сообщений
var userMessages = struct {
	sync.RWMutex
	data map[int64]int
}{data: make(map[int64]int)}

// Сохраняет ID сообщения для пользователя
func setUserMessageID(chatID int64, messageID int) {
	userMessages.Lock()
	defer userMessages.Unlock()
	userMessages.data[chatID] = messageID
}

// Возвращает сохраненный ID сообщения пользователя
func getUserMessageID(chatID int64) (int, bool) {
	userMessages.RLock()
	defer userMessages.RUnlock()
	messageID, exists := userMessages.data[chatID]
	return messageID, exists
}

// Удаляет сохраненный ID сообщения пользователя
func deleteUserMessageID(chatID int64) {
	userMessages.Lock()
	defer userMessages.Unlock()
	delete(userMessages.data, chatID)
}

// Структура для хранения текста сообщений и их ID
var userMessagesText = struct {
	sync.RWMutex
	data         map[int64]int
	ownerMessage struct {
		text   string // Текст сообщения с ID пользователя
		exists bool   // Флаг наличия сообщения
	}
}{data: make(map[int64]int)}

// Сохраняет текст сообщения с ID пользователя
func setOwnerMessageText(messageText string) {
	userMessagesText.Lock()
	defer userMessagesText.Unlock()
	userMessagesText.ownerMessage.text = messageText
	userMessagesText.ownerMessage.exists = true
}

// Извлекает ID пользователя из сохраненного сообщения
func extractUserIDFromMessage() (int64, error) {
	userMessagesText.RLock()
	defer userMessagesText.RUnlock()

	if !userMessagesText.ownerMessage.exists {
		return 0, errors.New("сообщение не найдено")
	}

	userID, err := strconv.ParseInt(userMessagesText.ownerMessage.text, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
