package functions

// ================================================FIND FUNCS==============================================================================
func findAndAdd(slice [][]string, value string, newValue string) ([][]string, bool) {
	for i, subSlice := range slice {
		if contains(subSlice, value) {
			slice[i] = append(slice[i], newValue)
			return slice, true
		}
	}
	return slice, false
}
func findIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// ================================================CONTAIN FUNCS==============================================================================
func containsInNested(slice [][]string, value string) bool {
	for _, subSlice := range slice {
		if contains(subSlice, value) {
			return true
		}
	}
	return false
}

// ================================================GET FUNCS==============================================================================
func getAdjacentDays(currentDay string) (string, string) {

	daysOfWeek := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}
	currentIndex := -1
	for i, day := range daysOfWeek {
		if day == currentDay {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return "", ""
	}
	prevIndex := (currentIndex - 1 + len(daysOfWeek)) % len(daysOfWeek)
	nextIndex := (currentIndex + 1) % len(daysOfWeek)

	return daysOfWeek[prevIndex], daysOfWeek[nextIndex]
} // Лево Право День недели
