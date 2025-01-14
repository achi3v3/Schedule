package funcExcel

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func FunctionScheduleBot() {
	fmt.Println("func: Schedule Bot")

	// createDataBases()

	LaunchScheduleBot()
}

func LaunchScheduleBot() {
	ownerId := 5266257091
	GlobalWeek := 17

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("Telegram bot token is required!")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	log.Printf("Authorized as %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Error getting updates: %v", err)
	}
	state := make(map[string]string)
	// ==============================================POSTGRESQL==============================================
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к базе данных: %s", err)
	}
	defer db.Close()
	// ==============================================DataBasesCreate==============================================
	createTableUsers(db)
	createDataBasesExcel(db)
	// ==============================================FOLDER==============================================
	saveDir := "uploaded_files"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err := os.Mkdir(saveDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}
	// ==============================================UPDATES==============================================

	var currentMessage *tgbotapi.Message
	for update := range updates {
		if update.Message != nil {
			if !isRequestAllowed(int64(update.Message.From.ID)) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚫 Куда торопишься молодой.")
				bot.Send(msg)
				continue
			}
			if update.Message != nil && update.Message.From != nil {
				userId := update.Message.From.ID
				username := update.Message.From.UserName
				firstname := update.Message.From.FirstName
				flag := addUser(db, userId, username, firstname)
				if flag {
					if update.Message.From.UserName == "" {
						username = "none"
					}
					if update.Message.From.FirstName == "" {
						firstname = "none"
					}
					msgText := fmt.Sprintf("<b>✅ Новый пользователь</b>\n\n<blockquote>id: %d</blockquote>\nusername: @%s\n<blockquote>firstname: %s</blockquote>\n", userId, username, firstname)
					msg := tgbotapi.NewMessage(int64(ownerId), msgText)
					msg.ParseMode = tgbotapi.ModeHTML

					bot.Send(msg)
				}
			}
			if update.Message.Document != nil && update.Message.From.ID == ownerId {
				document := update.Message.Document
				fileID := document.FileID

				file, err := bot.GetFile(tgbotapi.FileConfig{
					FileID: fileID,
				})
				if err != nil {
					log.Printf("Ошибка получения файла: %v", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🆘 Произошла ошибка при получении файла.")
					bot.Send(msg)
					continue
				}

				// Генерируем имя для сохранения файла
				fileName := fmt.Sprintf("%d_%s", update.Message.Chat.ID, document.FileName)
				savePath := filepath.Join(saveDir, fileName)

				// URL для файла
				fileURL := "https://api.telegram.org/file/bot" + bot.Token + "/" + file.FilePath

				filenm, err := downloadFile(fileURL, savePath)
				if err != nil {
					log.Printf("Ошибка загрузки файла: %v", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🆘 Не удалось загрузить файл.")
					bot.Send(msg)
					continue
				}
				ReloadFile(filenm)

				createTableUsers(db)
				createDataBasesExcel(db)

				createButtonActions(state)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Файл успешно сохранен как %s", fileName))
				bot.Send(msg)
			}
			if update.Message.Text == "/snupload_schedule" {
				if update.Message.From.ID == ownerId {
					fmt.Println("owner")
					msgText := "✅ Администратор\nОжидание Excel-файла:"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🆘 Недостаточно прав.")
					bot.Send(msg)
				}
			} else if update.Message.Text == "/start" {
				msgText := fmt.Sprintf("<a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Бета-версия)\n\n📆 Установленная неделя: %d\n🔎 Для поиска расписания:\n/snstart_schedule", GlobalWeek)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				msg.ParseMode = tgbotapi.ModeHTML

				sentMsg, _ := bot.Send(msg)
				currentMessage = &sentMsg
			} else if update.Message.Text == "/snstart_schedule" {
				state["course"], state["group"], state["day"] = "", "", ""
				msgText := fmt.Sprintf("🏛 Расписание by <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a>\n📆 Установленная неделя: %d\n\nВыберите уровень обучения:", GlobalWeek)

				buttonActions := createButtonActions(state)
				inlineKeyboard := dynamic_buttonsFromActions(buttonActions, state)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = tgbotapi.ModeHTML

				sentMsg, _ := bot.Send(msg)
				currentMessage = &sentMsg
			} else {
				text := "🆘 Извините, я вас не понимаю.\nДоступные команды:\n/start\n/snstart_schedule"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
				bot.Send(msg)
			}
		}

		allrangetime := []string{
			"08.00-09.20",
			"09.35-10.55",
			"11.35-12.55",
			"13.10-14.30",
			"15.10-16.30",
			"16.45-18.05",
			"18.20-19.40",
			"18.20-19.40",
			"19.55-21.15",
		}

		if update.CallbackQuery != nil {

			callbackData := update.CallbackQuery.Data
			handleButtonClick(update, bot, callbackData, state)
			var inlineKeyboard *tgbotapi.InlineKeyboardMarkup
			var msgText string

			if state["course"] == "" && state["group"] == "" {
				state["course"], state["group"], state["day"] = "", "", ""
				msgText := fmt.Sprintf("🏛 Расписание <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a>\n📆 Установленная неделя: %d\n\nВыберите уровень обучения:", GlobalWeek)

				buttonActions := createButtonActions(state)
				inlineKeyboard := dynamic_buttonsFromActions(buttonActions, state)

				if currentMessage != nil {
					editMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, currentMessage.MessageID, msgText)
					editMsg.ReplyMarkup = inlineKeyboard
					editMsg.ParseMode = tgbotapi.ModeHTML
					if _, err := bot.Send(editMsg); err != nil {
						log.Printf("Ошибка при редактировании сообщения: %v", err)
					}
				}

			} else if state["course"] != "" && state["group"] == "" {
				groups := get_groups(get_file_excel(), state["course"])
				inlineKeyboard = dynamic_buttons(groups, state)
				msgText = fmt.Sprintf("🏛 Расписание <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a>\n📆 Установленная неделя: %d\n\nУровень обучения: %s\nВыберите группу:", GlobalWeek, state["course"])
				backButton := tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_course")
				inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{backButton})

				if currentMessage != nil {
					editMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, currentMessage.MessageID, msgText)
					editMsg.ReplyMarkup = inlineKeyboard
					editMsg.ParseMode = tgbotapi.ModeHTML
					if _, err := bot.Send(editMsg); err != nil {
						log.Printf("Ошибка при редактировании сообщения: %v", err)
					}
				}

			} else if state["course"] != "" && state["group"] != "" && state["day"] == "" {
				days := get_days_for_couple()
				inlineKeyboard = dynamic_buttons(days, state)
				backButton := tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_group")
				inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{backButton})

				msgText = fmt.Sprintf("🏛 Расписание <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a>\n📆 Установленная неделя: %d\n\nУровень обучения: %s\nГруппа: %s\nВыберите день:", GlobalWeek, state["course"], state["group"])

				if currentMessage != nil {
					editMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, currentMessage.MessageID, msgText)
					editMsg.ReplyMarkup = inlineKeyboard
					editMsg.ParseMode = tgbotapi.ModeHTML
					if _, err := bot.Send(editMsg); err != nil {
						log.Printf("Ошибка при редактировании сообщения: %v", err)
					}
				}

			} else if state["course"] != "" && state["group"] != "" && state["day"] != "" {
				startcoupleString := fmt.Sprintf("🏛 Расписание <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a>\n📆 Установленная неделя: %d\n\nУровень обучения: %s\nГруппа: %s\n\n📅 %s\n\n", GlobalWeek, state["course"], state["group"], state["day"])
				coupleList := FunctionDataBaseTableData(state["course"], state["group"], state["day"])
				coupleString, flagConcatenateAuditory, flagConcatenateTeacher := "", "", ""
				for i := 0; i < len(coupleList); i++ {
					numberCoupleTime := 0
					CoupleTime := coupleList[i][1]
					if contains(allrangetime, CoupleTime) {
						numberCoupleTime = findIndex(allrangetime, CoupleTime) + 1
					}
					CoupleSubject := coupleList[i][2]
					CoupleAuditory := coupleList[i][3]
					CoupleTeacher := coupleList[i][4]
					CoupleWeeks := coupleList[i][5]
					if flagConcatenateAuditory != "" {
						CoupleAuditory = fmt.Sprintf("%s / %s", CoupleAuditory, flagConcatenateAuditory)
						flagConcatenateAuditory = ""
					}
					if flagConcatenateTeacher != "" {
						CoupleTeacher = fmt.Sprintf("%s / %s", CoupleTeacher, flagConcatenateTeacher)
						flagConcatenateTeacher = ""
					}
					if i+1 < len(coupleList) {
						if CoupleTime == coupleList[i+1][1] && CoupleSubject == coupleList[i+1][2] {
							if CoupleTeacher == coupleList[i+1][4] && CoupleAuditory != coupleList[i+1][3] {
								flagConcatenateAuditory = CoupleAuditory
								continue
							} else if CoupleTeacher != coupleList[i+1][4] && CoupleAuditory != coupleList[i+1][3] {
								flagConcatenateAuditory = CoupleAuditory
								flagConcatenateTeacher = CoupleTeacher
								continue
							} else if CoupleTeacher != coupleList[i+1][4] && CoupleAuditory == coupleList[i+1][3] {
								flagConcatenateTeacher = CoupleTeacher
								continue
							}
						}
					}
					if CoupleWeeks != "—" {
						coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(%d пара)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i>\n    🔍 <i>%s</i></blockquote>\n", CoupleTime, numberCoupleTime, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher, removeBrackets(CoupleWeeks))
					} else {
						coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(%d пара)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i></blockquote>\n", CoupleTime, numberCoupleTime, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher)

					}
				}

				coupleString = startcoupleString + coupleString

				prevDay, nextDay := getAdjacentDays(state["day"])
				prevButton := tgbotapi.NewInlineKeyboardButtonData(prevDay, prevDay)
				nextButton := tgbotapi.NewInlineKeyboardButtonData(nextDay, nextDay)

				backButton := tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_day")

				inlineKeyboard := dynamic_buttons([]string{}, state)
				inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{prevButton, backButton, nextButton})

				if currentMessage != nil {
					editMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, currentMessage.MessageID, coupleString)
					editMsg.ReplyMarkup = inlineKeyboard
					editMsg.ParseMode = tgbotapi.ModeHTML
					if _, err := bot.Send(editMsg); err != nil {
						log.Printf("Ошибка при редактировании сообщения: %v", err)
					}
				}
			}
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Я всё вижу")
			bot.AnswerCallbackQuery(callback)
		}
	}
}
