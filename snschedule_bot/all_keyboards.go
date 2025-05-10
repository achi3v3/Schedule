/*
Файл содержит функции для создания и управления клавиатурами Telegram бота.
Основные функции:
  - Создание клавиатур для навигации по меню
  - Управление выбором курса, группы и дня
  - Формирование клавиатур для различных панелей управления
  - Обработка пагинации и навигации

(используется для создания интерактивных элементов интерфейса)
*/
package functions

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

// Константы для текстов кнопок
const (
	// Общие кнопки
	btnBack           = "Назад"
	btnAdd            = "✏️ Добавить"
	btnClear          = "❌ Очистить записи"
	btnViewRecords    = "Просмотреть записи"
	btnUploadSchedule = "Загрузить расписание"
	btnInfoSheet      = "Информационный лист"
	btnSupport        = "🧾 Поддержка"
	btnPinGroup       = "📌 Закрепить группу"
	btnUnpinGroup     = "🔓 Открепить группу"
	btnMyGroup        = "🔒 Моя группа"
	btnSchedule       = "🎓 Расписание"
	btnControlPanel   = "⚙️ Панель управления"
	btnEditorPanel    = "🪪 Панель редактора"
	btnExcelFile      = "📂 Эксель-файл"
	btnInfo           = "📃 Информация"
	btnMyRecords      = "Мои записи"

	// Кнопки управления пользователями
	btnUsers        = "Пользователи"
	btnEditors      = "Редакторы"
	btnAdmins       = "Администраторы"
	btnMakeEditor   = "Сделать редактором"
	btnRemoveEditor = "Убрать редактора"
	btnMakeAdmin    = "Назначить админом"
	btnRemoveAdmin  = "Убрать админку"
)

// Константы для callback-данных
const (
	cbHome           = "home"
	cbBack           = "back"
	cbAdd            = "Добавить"
	cbClearRecords   = "Очистить записи"
	cbViewRecords    = "Просмотреть записи"
	cbSchedule       = "Расписание"
	cbPinGroup       = "Закрепить группу"
	cbUnpinGroup     = "Открепить группу"
	cbMyGroup        = "Моя группа"
	cbControlPanel   = "Панель управления"
	cbEditorPanel    = "Уполномоченным"
	cbSendFile       = "Отправить файл"
	cbInfo           = "Информация"
	cbUsers          = "Пользователи"
	cbEditors        = "Редакторы"
	cbAdmins         = "Администраторы"
	cbUploadSchedule = "Загрузить расписание"
	cbInfoSheet      = "Информационный лист"
	cbMyRecords      = "Мои записи"
	cbMakeEditor     = "Сделать редактором"
	cbRemoveEditor   = "Убрать редактора"
	cbMakeAdmin      = "Сделать админом"
	cbRemoveAdmin    = "Убрать админку"
)

// URL для кнопки поддержки
const supportURL = "https://t.me/sn_mira"

// Дни недели
var weekDays = []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}

// Вспомогательные функции для создания кнопок
func createButton(text, callbackData string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         text,
		CallbackData: callbackData,
	}
}

func createURLButton(text, url string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text: text,
		URL:  url,
	}
}

func createBackButton(callbackData string) models.InlineKeyboardButton {
	return createButton(btnBack, callbackData)
}

// Функция для создания ряда кнопок с группировкой
func createButtonRow(buttons ...models.InlineKeyboardButton) []models.InlineKeyboardButton {
	return buttons
}

// Создает клавиатуру для выбора курса
// Возвращает InlineKeyboardMarkup с кнопками курсов, сгруппированными по 3 в ряд
// Последняя строка содержит кнопку "Назад"
func CourseSelection() *models.InlineKeyboardMarkup {
	courses, _ := GetAllSheets()
	kb := New()
	var row []models.InlineKeyboardButton

	for i, course := range courses {
		if course == "Расписание" {
			continue
		}

		row = append(row, createButton(course, course))

		if (i)%3 == 0 || course == "4 курс" {
			kb.AddRow(row...)
			row = []models.InlineKeyboardButton{}
		}
	}

	if len(row) > 0 {
		kb.AddRow(row...)
	}

	kb.AddRow(createBackButton(cbHome))
	return kb.Build()
}

// Создает клавиатуру для выбора группы в рамках выбранного курса
// Принимает название курса для получения списка доступных групп
// Возвращает InlineKeyboardMarkup с кнопками групп, сгруппированными по 3 в ряд
// Последняя строка содержит кнопку "Назад"
func GroupSelection(course string) *models.InlineKeyboardMarkup {
	groups, _ := GetGroupsByCourseRu(course)
	kb := New()
	var row []models.InlineKeyboardButton

	for i, group := range groups {
		row = append(row, createButton(group, group))

		if (i+1)%3 == 0 || i == len(groups)-1 {
			kb.AddRow(row...)
			row = []models.InlineKeyboardButton{}
		}
	}

	kb.AddRow(createBackButton(cbBack))
	return kb.Build()
}

// Создает клавиатуру для выбора дня недели
// Возвращает InlineKeyboardMarkup с кнопками дней недели, сгруппированными по 3 в ряд
// Дополнительно содержит кнопки для закрепления группы и возврата назад
func DaySelection() *models.InlineKeyboardMarkup {
	kb := New()
	var row []models.InlineKeyboardButton

	for i, day := range weekDays {
		row = append(row, createButton(day, day))

		if (i+1)%3 == 0 || i == len(weekDays)-1 {
			kb.AddRow(row...)
			row = []models.InlineKeyboardButton{}
		}
	}

	kb.AddRow(createButton(btnPinGroup, cbPinGroup))
	kb.AddRow(createBackButton(cbBack))
	return kb.Build()
}

// Создает клавиатуру для навигации по расписанию
// Принимает названия предыдущего и следующего дня для создания кнопок навигации
// Возвращает InlineKeyboardMarkup с кнопками добавления записи, навигации по дням и возврата
func ScheduleKeyboard(prevDay, nextDay string) *models.InlineKeyboardMarkup {
	kb := New()

	kb.AddRow(createButton(btnAdd, cbAdd))
	kb.AddRow(
		createButton(prevDay, prevDay),
		createBackButton(cbBack),
		createButton(nextDay, nextDay),
	)

	return kb.Build()
}

// Создает основную клавиатуру бота (главное меню)
// Возвращает InlineKeyboardMarkup с основными функциями:
// - Просмотр расписания
// - Управление закрепленной группой
// - Доступ к панели редактора
// - Работа с Excel-файлами
// - Информация и поддержка
func StartKeyboard() *models.InlineKeyboardMarkup {
	kb := New()

	kb.AddRow(createButton(btnSchedule, cbSchedule))
	kb.AddRow(
		createButton(btnMyGroup, cbMyGroup),
		createButton(btnUnpinGroup, cbUnpinGroup),
	)
	kb.AddRow(createButton(btnEditorPanel, cbEditorPanel))
	kb.AddRow(
		createButton(btnExcelFile, cbSendFile),
		createButton(btnInfo, cbInfo),
	)
	kb.AddRow(createButton(btnControlPanel, cbControlPanel))
	kb.AddRow(createURLButton(btnSupport, supportURL))

	return kb.Build()
}

// Создает клавиатуру панели управления
// Возвращает InlineKeyboardMarkup с функциями управления:
// - Управление пользователями
// - Управление редакторами и администраторами
// - Загрузка расписания
// - Управление информационным листом
func ControlPanelKeyboard() *models.InlineKeyboardMarkup {
	kb := New()

	kb.AddRow(createButton(btnUsers, cbUsers))
	kb.AddRow(
		createButton(btnEditors, cbEditors),
		createButton(btnAdmins, cbAdmins),
	)
	kb.AddRow(createButton(btnUploadSchedule, cbUploadSchedule))
	kb.AddRow(createButton(btnInfoSheet, cbInfoSheet))
	kb.AddRow(createBackButton(cbHome))

	return kb.Build()
}

// Создает клавиатуру панели редактора
// Возвращает InlineKeyboardMarkup с функциями для редакторов:
// - Просмотр своих записей
// - Управление записями
func RedactorPanelKeyboard() *models.InlineKeyboardMarkup {
	kb := New()

	kb.AddRow(createButton(btnMyRecords, cbMyRecords))
	kb.AddRow(createBackButton(cbHome))

	return kb.Build()
}

// Создает клавиатуру для навигации по уведомлениям
// Принимает текущую страницу, общее количество уведомлений и флаг принадлежности уведомлений
// Возвращает InlineKeyboardMarkup с кнопками навигации и управления уведомлениями
func NoticesKeyboard(page, totalNotices int, isMyNotices bool) *models.InlineKeyboardMarkup {
	kb := New()

	if page > 0 || page*3 < totalNotices {
		var navButtons []models.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, createButton("Предыдущий", fmt.Sprintf("Мои записи:%d", page-1)))
		}
		if page*3 < totalNotices {
			navButtons = append(navButtons, createButton("Следующий", fmt.Sprintf("Мои записи:%d", page+1)))
		}
		if len(navButtons) > 0 {
			kb.AddRow(navButtons...)
		}
	}

	if totalNotices > 0 {
		kb.AddRow(createButton(btnClear, cbClearRecords))
	} else {
		kb.AddRow(createButton(btnAdd, "Группа"))
	}

	kb.AddRow(createBackButton(cbEditorPanel))
	return kb.Build()
}

// Создает клавиатуру для управления правами пользователей
// Принимает флаги наличия прав редактора и администратора
// Возвращает InlineKeyboardMarkup с соответствующими кнопками управления правами
func UserPermissionsKeyboard(isRedactor, isAdmin bool) *models.InlineKeyboardMarkup {
	kb := New()

	if !isRedactor {
		kb.AddRow(createButton(btnMakeEditor, cbMakeEditor))
	} else {
		kb.AddRow(createButton(btnRemoveEditor, cbRemoveEditor))
	}

	if !isAdmin {
		kb.AddRow(createButton(btnMakeAdmin, cbMakeAdmin))
	} else {
		kb.AddRow(createButton(btnRemoveAdmin, cbRemoveAdmin))
	}

	kb.AddRow(createButton(btnViewRecords, cbViewRecords))
	kb.AddRow(createBackButton(cbControlPanel))

	return kb.Build()
}

// Создает клавиатуру с кнопкой "Назад"
// Принимает callbackData для кнопки возврата
// Возвращает InlineKeyboardMarkup с одной кнопкой возврата
func BackKeyboard(callbackData string) *models.InlineKeyboardMarkup {
	kb := New()
	kb.AddRow(createBackButton(callbackData))
	return kb.Build()
}
