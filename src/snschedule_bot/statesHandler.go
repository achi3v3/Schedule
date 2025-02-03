package functions

import (
	"errors"
	"strconv"
	"sync"
)

// ========================================================USER_STATES===========================================================================
type UserState struct {
	Course string
	Group  string
	Day    string
}

var userStates = struct {
	sync.RWMutex
	data map[int64]map[string]string
}{data: make(map[int64]map[string]string)}

func getUserState(chatID int64) map[string]string {
	userStates.RLock()
	defer userStates.RUnlock()

	if state, exists := userStates.data[chatID]; exists {
		return state
	}
	return nil
}
func setUserState(chatID int64, key string, value string) {
	userStates.Lock()
	defer userStates.Unlock()

	if _, exists := userStates.data[chatID]; !exists {
		userStates.data[chatID] = make(map[string]string)
	}

	userStates.data[chatID][key] = value
}
func deleteUserState(chatID int64, key string) {
	userStates.Lock()
	defer userStates.Unlock()

	if state, exists := userStates.data[chatID]; exists {
		delete(state, key)
		if len(state) == 0 { // Если мапа пустая, удаляем весь объект
			delete(userStates.data, chatID)
		}
	}
}
func resetUserState(chatID int64) {
	userStates.Lock()
	defer userStates.Unlock()
	delete(userStates.data, chatID)
}

// ========================================================USER_MESSAGES========================================================================
var userMessages = struct {
	sync.RWMutex
	data map[int64]int // chatID -> messageID
}{data: make(map[int64]int)}

func setUserMessageID(chatID int64, messageID int) {
	userMessages.Lock()
	defer userMessages.Unlock()
	userMessages.data[chatID] = messageID
}
func getUserMessageID(chatID int64) (int, bool) {
	userMessages.RLock()
	defer userMessages.RUnlock()
	messageID, exists := userMessages.data[chatID]
	return messageID, exists
}
func deleteUserMessageID(chatID int64) {
	userMessages.Lock()
	defer userMessages.Unlock()
	delete(userMessages.data, chatID)
}

var userMessagesText = struct {
	sync.RWMutex
	data         map[int64]int // chatID -> messageID
	ownerMessage struct {
		text   string // Хранение текста сообщения (содержит ID пользователя)
		exists bool
	}
}{data: make(map[int64]int)}

func setOwnerMessageText(messageText string) {
	userMessagesText.Lock()
	defer userMessagesText.Unlock()
	userMessagesText.ownerMessage.text = messageText
	userMessagesText.ownerMessage.exists = true
}

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

// func getOwnerMessageText() (string, bool) {
// 	userMessagesText.RLock()
// 	defer userMessagesText.RUnlock()
// 	return userMessagesText.ownerMessage.text, userMessagesText.ownerMessage.exists
// }
