package functions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	UpArrow   = "🔝"
	DownArrow = "🔻"
)

// Универсальный обработчик для всех типов сообщений // РАЗДЕЛИТЬ НА НЕСКОЛЬКО ФУНКЦИЙ
func UniversalHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	if update == nil {
		log.Println("Update is nil")
		return
	}
	// Проверяем тип чата (Chat не является указателем, поэтому nil-проверка не нужна)

	if update.Message != nil {
		if update.Message.Document != nil {
			if update.Message.From.ID == idOwner {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "✅ Администратор",
				})
				handleDocument(ctx, b, update)

			} else {
				sendNotPermisions(ctx, b, update.Message.Chat.ID)
			}
			return
		}
		if update.Message.Text != "" {
			if strings.Contains(strings.ToLower(update.Message.Text), "добавить") {
				deleteUserMessageID(update.Message.Chat.ID)
				userID := update.Message.From.ID
				chatID := update.Message.Chat.ID
				if IsRedactorsByUserID(userID) || idOwner == userID {
					state := getUserState(userID)
					course, group := getPermCourseGroupByUserID(userID)
					if state["course"] == course && state["group"] == group {
						if state["day"] == "" {
							deleteUserMessageID(update.Message.Chat.ID)
							b.SendMessage(ctx, &bot.SendMessageParams{
								ChatID: chatID,
								Text:   "⛔️ У вас не выбран день, выберите день",
							})
							sendDaySelection(ctx, b, chatID, state)
							return
						}
						student := update.Message.From.Username
						if student == "" {
							student = fmt.Sprintf("id: %d", userID)
						} else {
							student = fmt.Sprintf("@%s", student)
						}
						notice := parseSchedule(update.Message.Text)
						msg := PinsByStarost(state["course"], state["group"], state["day"], notice["time"], notice["event"], notice["auditory"], notice["teacher"], notice["weeks"], student, true, userID)
						if msg != "" {
							sendOnlyMessage(ctx, b, chatID, msg)
						} else {
							sendError(ctx, b, chatID)
						}
						deleteUserMessageID(update.Message.Chat.ID)
					} else if state["course"] != "" && state["group"] != "" {
						var keyboardRows [][]models.InlineKeyboardButton
						keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
							{Text: "Перейти к группе", CallbackData: "Группа"},
						})

						keyboard := &models.InlineKeyboardMarkup{
							InlineKeyboard: keyboardRows,
						}
						msg := fmt.Sprintf("<b>⛔️ У вас нет доступа добавлять записи в:</b><blockquote>%s / %s</blockquote>\n\n<i>У вас есть доступ только тут:</i>\n<blockquote>Уровень обучения: %s\nКурс: %s</blockquote>", state["course"], state["group"], course, group)

						sendEditMessage(ctx, b, chatID, msg, keyboard)
					} else {
						var keyboardRows [][]models.InlineKeyboardButton
						keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
							{Text: "Перейти к группе", CallbackData: "Группа"},
						})

						keyboard := &models.InlineKeyboardMarkup{
							InlineKeyboard: keyboardRows,
						}
						msg := fmt.Sprintf("<b>⛔️ У вас не выбраны курс и группа!</b>\n\n<i>Выберите вашу группу:</i>\n<blockquote>Уровень обучения: %s\nКурс: %s</blockquote>", course, group)
						sendEditMessage(ctx, b, chatID, msg, keyboard)
					}
				} else {
					sendNotPermisions(ctx, b, update.Message.Chat.ID)
				}
				return
			}
			if IsAdminByUserID(update.Message.From.ID) || idOwner == update.Message.From.ID {
				setOwnerMessageText(update.Message.Text)
				deleteUserMessageID(update.Message.Chat.ID)
				MessageUserID, err := extractUserIDFromMessage()
				if err != nil {
					sendError(ctx, b, update.Message.Chat.ID)
					fmt.Println("error", update.Message.Text)
				} else {
					sendUpdatePermisions(ctx, b, update.Message.From.ID, MessageUserID)
				}
				return
			}
		}
	}
	if update.CallbackQuery == nil {
		log.Println("CallbackQuery is nil")
		return
	}

	callbackQuery := update.CallbackQuery
	message := callbackQuery.Message.Message
	chatID := message.Chat.ID
	callbackData := callbackQuery.Data
	userID := callbackQuery.From.ID

	if isSpamming(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "💢 Не торопись молод(ой)ая..",
		})
		return
	}

	state := getUserState(chatID)
	if state == nil {
		userStates.Lock()
		userStates.data[chatID] = make(map[string]string)
		userStates.Unlock()
		state = getUserState(chatID)
	}

	if callbackData == "home" {
		sendStart(ctx, b, chatID)
		return
	} else if callbackData == "Сделать редактором" {
		if IsAdminByUserID(userID) || idOwner == userID {
			setredactorID, err := extractUserIDFromMessage()
			if setredactorID == userID {
				b.SendMessage(ctx, &bot.SendMessageParams{
					Text:   "Нихуя.. сам себе роли назначает)",
					ChatID: chatID,
				})
			}
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				addRedactor(ctx, setredactorID)
				sendNotifyRole(ctx, b, userID, setredactorID, UpArrow, "«Редактор»")
				sendAddRole(ctx, b, setredactorID, "Редактор")
			}
			sendUpdatePermisions(ctx, b, chatID, setredactorID)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Убрать редактора" {
		if IsAdminByUserID(userID) || idOwner == userID {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				deleteRedactor(ctx, setredactorID)
				sendNotifyRole(ctx, b, userID, setredactorID, DownArrow, "«Редактор»")
				sendDeleteRole(ctx, b, setredactorID, "Редактор")
			}
			sendUpdatePermisions(ctx, b, chatID, setredactorID)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Сделать админом" {
		if userID == idOwner {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				setRoleAdmin(ctx, setredactorID, true)
				sendNotifyRole(ctx, b, userID, setredactorID, UpArrow, "«Администратор»")
				sendAddRole(ctx, b, setredactorID, "Администратор")
			}
			sendUpdatePermisions(ctx, b, chatID, setredactorID)
		} else if IsAdminByUserID(userID) || idOwner == userID {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				sendRequestForSetRoleAdmin(ctx, b, chatID, userID, setredactorID)
			}
			sendUpdatePermisions(ctx, b, chatID, setredactorID)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Убрать админку" {
		if userID == idOwner {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				setRoleAdmin(ctx, setredactorID, false)
				sendNotifyRole(ctx, b, userID, setredactorID, DownArrow, "«Администратор»")
				sendDeleteRole(ctx, b, setredactorID, "Администратор")
			}
			sendUpdatePermisions(ctx, b, chatID, setredactorID)
		} else {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				sendNotifyRole(ctx, b, userID, setredactorID, DownArrow, "«Администратор»")
				sendNotPermisions(ctx, b, chatID)
			}
		}
		return
	} else if callbackData == "back" {
		if state["day"] != "" && state["group"] != "" && state["course"] != "" {
			deleteUserState(chatID, "day")
			sendDaySelection(ctx, b, chatID, state)
		} else if state["group"] != "" && state["course"] != "" {
			deleteUserState(chatID, "day")
			deleteUserState(chatID, "group")
			sendGroupSelection(ctx, b, chatID, state)
		} else if state["course"] != "" {
			sendCourseSelection(ctx, b, chatID, true)
		}
		return

	} else if callbackData == "Расписание" {
		sendCourseSelection(ctx, b, chatID, true)
		return
	} else if callbackData == "Моя группа" {
		course, group, err := GetUserCourseAndGroup(ctx, chatID)
		if course != "" && group != "" && err == nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      fmt.Sprintf("✅ Ваша группа установлена:\n<blockquote>Уровень обучения: %s\nГруппа: %s</blockquote>", course, group),
				ParseMode: models.ParseModeHTML,
			})
			setUserState(chatID, "course", course)
			setUserState(chatID, "group", group)
			state := getUserState(chatID)

			sendDaySelection(ctx, b, chatID, state)
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      "⭕️ За вами не закреплены курс и группа.. Закрепите группу в расписании",
				ParseMode: models.ParseModeHTML,
			})
		}
		return
	} else if callbackData == "Открепить группу" {
		_, err := PinGroup(ctx, userID, "", "")
		if err != nil {
			fmt.Println("error", callbackData)
			return
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      "🪪 Ваша данные очищены:\n<blockquote>Уровень обучения: \nГруппа: </blockquote>",
			ParseMode: models.ParseModeHTML,
		})

		return
	} else if callbackData == "Уполномоченным" {
		if IsRedactorsByUserID(userID) || userID == idOwner {
			sendRedactorPanel(ctx, b, chatID)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Отправить файл" {
		if userID == idOwner {
			nameFileSlice, _ := getExcelName()
			nameFile := nameFileSlice[0]
			fileLocate := "./uploaded_files"
			sendFile(b, chatID, fileLocate, nameFile)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Информация" {
		sendInfo(ctx, b, chatID)
		return
	} else if callbackData == "Информационный лист" {
		if IsAdminByUserID(userID) || idOwner == userID {
			sendAdminInfo(ctx, b, chatID)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Панель управления" {
		if IsAdminByUserID(userID) || idOwner == userID {
			sendControPanel(ctx, b, chatID)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return

	} else if callbackData == "Мои записи" {
		if IsRedactorsByUserID(userID) || idOwner == userID {
			sendMyNotices(ctx, b, chatID, userID, 0)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if strings.Contains(callbackData, "Мои записи:") {
		if IsAdminByUserID(userID) || userID == idOwner {
			ind := strings.Index(callbackData, ":")
			page, err := strconv.Atoi(callbackData[ind+1:])
			if err == nil {
				sendMyNotices(ctx, b, chatID, userID, page)
			}
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Просмотреть записи" {
		if IsAdminByUserID(userID) || idOwner == userID {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				sendNoticesByUserID(ctx, b, chatID, setredactorID, 0)
			}
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if strings.Contains(callbackData, "Просмотреть записи:") {
		if IsAdminByUserID(userID) || userID == idOwner {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			} else {
				ind := strings.Index(callbackData, ":")
				page, err := strconv.Atoi(callbackData[ind+1:])
				if err == nil {
					sendNoticesByUserID(ctx, b, chatID, setredactorID, page)
				}
			}
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Просмотр пользователя" {
		setredactorID, err := extractUserIDFromMessage()
		if err != nil {
			fmt.Println("error", callbackData)
		} else {
			if IsAdminByUserID(userID) || idOwner == userID {
				sendUpdatePermisions(ctx, b, chatID, setredactorID)
			} else {
				sendNotPermisions(ctx, b, chatID)
			}
		}
		return
	} else if callbackData == "Очистить записи пользователя" {
		if IsAdminByUserID(userID) || idOwner == userID {
			setredactorID, err := extractUserIDFromMessage()
			if err != nil {
				fmt.Println("error", callbackData)
			}
			err = deleteNoticessByUserID(setredactorID)
			if err == nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   fmt.Sprintf("✅ Все записи пользователя %d были удалены", setredactorID),
				})
				sendNoticesByUserID(ctx, b, chatID, setredactorID, 0)
			} else {
				fmt.Println("error", callbackData)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   fmt.Sprintf("❌ Ошибка: %v", err),
				})
			}
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Очистить записи" {
		err := deleteNoticessByUserID(userID)
		if err == nil {
			msg := "✅ Все записи были удалены"
			sendOnlyMessage(ctx, b, chatID, msg)
		} else {
			fmt.Println("error", callbackData)
			sendError(ctx, b, chatID)
		}
		sendMyNotices(ctx, b, chatID, userID, 0)
		return
	} else if callbackData == "Группа" {
		course, group := getPermCourseGroupByUserID(userID)

		if course != "" && group != "" {
			setUserState(chatID, "course", course)
			setUserState(chatID, "group", group)
			state := getUserState(chatID)

			sendDaySelection(ctx, b, chatID, state)
		} else {
			msg := "⭕️ За вами не закреплены курс и группа.. Обратитесь в поддержку"
			sendOnlyMessage(ctx, b, chatID, msg)
		}
		return

	} else if callbackData == "Редакторы" {
		if userID == idOwner {
			sendGetRedactors(ctx, b, chatID, 0)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if strings.Contains(callbackData, "Редакторы:") {
		if IsAdminByUserID(userID) || userID == idOwner {
			ind := strings.Index(callbackData, ":")
			page, err := strconv.Atoi(callbackData[ind+1:])
			if err == nil {
				sendGetRedactors(ctx, b, chatID, page)
			}
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Администраторы" {
		if IsAdminByUserID(userID) || userID == idOwner {
			sendGetAdmins(ctx, b, chatID, 0)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if strings.Contains(callbackData, "Администраторы:") {
		if IsAdminByUserID(userID) || userID == idOwner {
			ind := strings.Index(callbackData, ":")
			page, err := strconv.Atoi(callbackData[ind+1:])
			if err == nil {
				sendGetAdmins(ctx, b, chatID, page)
			}
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Пользователи" {
		sendUsers(ctx, b, chatID, 0)
		return
	} else if strings.Contains(callbackData, "Пользователи:") {
		ind := strings.Index(callbackData, ":")
		page, err := strconv.Atoi(callbackData[ind+1:])
		if err == nil {
			sendUsers(ctx, b, chatID, page)
		}
		return
	} else if callbackData == "Загрузить расписание" {
		if userID == idOwner {
			sendUploadFile(ctx, b, chatID)
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if callbackData == "Закрепить группу" {
		if state["course"] != "" && state["group"] != "" {
			msg, err := PinGroup(ctx, userID, state["course"], state["group"])
			if err == nil {
				sendOnlyMessage(ctx, b, chatID, msg)
			} else {
				msg = "💢 Группа не зафиксирована.."
				sendOnlyMessage(ctx, b, chatID, msg)
			}
		} else {
			msg := "💢 Пожалуйста, выберите курс и группу перед закреплением."
			sendOnlyMessage(ctx, b, chatID, msg)
			// Без редактирования
			sendCourseSelection(ctx, b, chatID, false)
		}
		return
	} else if callbackData == "Добавить" {
		if IsRedactorsByUserID(userID) || idOwner == userID {
			course, group := getPermCourseGroupByUserID(userID)
			if course == "" || group == "" {
				msg := "⭕️ За вами не закреплены курс и группа.. Обратитесь в поддержку"
				sendOnlyMessage(ctx, b, chatID, msg)
			} else if state["course"] != course || state["group"] != group {
				var keyboardRows [][]models.InlineKeyboardButton
				keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
					{Text: "Перейти к группе", CallbackData: "Группа"},
				})

				keyboard := &models.InlineKeyboardMarkup{
					InlineKeyboard: keyboardRows,
				}
				msg := fmt.Sprintf("<b>⛔️ У вас нет доступа добавлять записи в:</b><blockquote>%s / %s</blockquote>\n\n<i>У вас есть доступ только тут:</i>\n<blockquote>Уровень обучения: %s\nКурс: %s</blockquote>", state["course"], state["group"], course, group)

				sendEditMessage(ctx, b, chatID, msg, keyboard)
			} else if state["course"] != "" && state["group"] != "" && state["day"] != "" {
				username := callbackQuery.From.Username
				if username == "" {
					username = fmt.Sprintf("id: %d", userID)
				} else {
					username = fmt.Sprintf("@%s", username)
				}
				msg := fmt.Sprintf("<b>✏️ Добавления записи</b> by %s\n\nУровень обучения: <u>%s</u>\nГруппа: <u>%s</u>\nДень: <u>%s</u>\n\n<i>Для добавления записи необходимо ввести:\n<blockquote>1 Время\n2 Мероприятие\n3 Аудитория\n4 Преподаватель\n5 Недели</blockquote></i>\n\n<b>Пример:</b>\n<blockquote><i>Добавить\n1 15.11-15.11\n2 Сдача лабораторных работ по Физике\n3 ауд. 666\n</i></blockquote>\n<i>Можете заметить, что некоторые пункты (4, 5) пропущены - это значит они будут пустые\nВАЖНО: строка начинается с <u><b>«Добавить»</b></u> и каждый новый пункт должен начинаться с новой строки.</i>", username, state["course"], state["group"], state["day"])
				sendOnlyMessage(ctx, b, chatID, msg)
			} else {
				msg := "💢 Курс/группа/день не выбраны.."
				sendOnlyMessage(ctx, b, chatID, msg)
			}
		} else {
			sendNotPermisions(ctx, b, chatID)
		}
		return
	} else if state["course"] == "" && state["group"] == "" && state["day"] == "" {
		setUserState(chatID, "course", callbackData)
		sendGroupSelection(ctx, b, chatID, state)
		return
	} else if state["group"] == "" && state["day"] == "" {
		setUserState(chatID, "group", callbackData)
		sendDaySelection(ctx, b, chatID, state)
		return
	} else {
		setUserState(chatID, "day", callbackData)
		schedule := getSchedule(state)
		sendSchedule(ctx, b, chatID, schedule, state)
		return
	}
}

func databaseHandler() {
	// =============================================POSTGRESQL==============================================
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к базе данных: %s", err)
	}
	defer db.Close()
	createFilesTable(db)
	createTableUsers(db)
	createDataBasesExcel(db)
	// ==============================================FOLDER==============================================
	saveDir := "uploaded_files"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err := os.Mkdir(saveDir, 0755)
		if err != nil {
			log.Printf("Failed to create directory: %v", err)
		}
	}
}
