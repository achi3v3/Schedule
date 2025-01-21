package funcExcel

import "sync"

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
