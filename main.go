package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/Syfaro/telegram-bot-api"
	"github.com/apsdehal/go-logger"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type cellsOrder struct {
	courseNum   int
	groupName   int
	date        int
	time        int
	subjectName int
	subjectType int
	teacher     int
	format      int
	classroom   int
}

func getPageTitleByMonth(month string) string {
	month_int_num, _ := strconv.Atoi(month)
	switch time.Month(month_int_num) {
	case time.February:
		return "фев21"
	case time.March:
		return "мар21"
	case time.April:
		return "апр21"
	case time.May:
		return "май21"
	case time.June:
		return "июн21"
	default:
		return "error"
	}
}

func getMonthCorrectName(month string) string {
	correct_name := getPageTitleByMonth(month)
	return correct_name
}

func performDate(_date string) string {
	return strings.Replace(_date, ".", "\".\"", 2)
}

func parseExcelFile(group_and_date string) string {
	curOrder := cellsOrder{
		courseNum:   1,
		groupName:   2,
		date:        3,
		time:        4,
		subjectName: 5,
		subjectType: 6,
		teacher:     7,
		format:      8,
		classroom:   9,
	}

	group := strings.Split(group_and_date, " ")
	group_name_by_user := group[0]
	_date := group[1]
	_date_splited := strings.Split(_date, ".")

	page_name := getMonthCorrectName(_date_splited[1])

	f, err := excelize.OpenFile("rhga.xlsx")
	if err != nil {
		return "can't open xlsx"
	}

	rows, err := f.GetRows(page_name)
	if err != nil {
		return "can't get neccessary row"
	}

	performed_date := performDate(_date)
	res := ""
	for _, row := range rows {
		group_name := row[curOrder.groupName]
		subject_name := row[curOrder.subjectName]
		_time := row[curOrder.time]
		_date := row[curOrder.date]
		subject_type := row[curOrder.subjectType]
		classroom := row[curOrder.classroom]
		teacher_name := row[curOrder.teacher]

		if strings.Contains(group_name, group_name_by_user) && strings.Contains(_date, performed_date) {

			if strings.TrimSpace(subject_name) == "" {
				subject_name = "-"
			}

			if strings.TrimSpace(subject_type) == "" {
				subject_type = "-"
			}

			if strings.TrimSpace(classroom) == "" {
				classroom = "-"
			} else if strings.Contains(classroom, "http") {
				classroom = "дистанционно"
			}

			if strings.TrimSpace(teacher_name) == "" {
				teacher_name = "-"
			}

			res = res + fmt.Sprintf("%v | %v | %v | %v | %v\n", _time, subject_name, subject_type, classroom, teacher_name)
		}
	}
	if len(res) == 0 {
		res = "Извините, но по совокупности названия группы и даты расписание не найдено"
	}
	return res
}

func sendMessageToUser(message string, update tgbotapi.Update, bot tgbotapi.BotAPI) {
	log, err := logger.New("main_logger", 1, os.Stdout)
	if err != nil {
		panic(err)
	}

	log.NoticeF("Answer '%s' was sent to user '%s'", message, update.Message.From.FirstName)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	bot.Send(msg)
}

func validateReq(req string) (bool, int) {
	first_check, _ := regexp.MatchString(`^[А-Я]{3}\.\d{3}\.[А-Я]$`, req)
	second_check, _ := regexp.MatchString(`^[А-Я]{3}\.\d{3}\.[А-Я] \d{2}\.\d{2}\.\d{2}`, req)
	if first_check {
		return true, 1
	} else if second_check {
		return true, 2
	} else {
		return false, 0
	}
}

func startTelegramBotCycle() {
	log, err := logger.New("main_logger", 1, os.Stdout)
	if err != nil {
		panic(err)
	}
	bot, err := tgbotapi.NewBotAPI("1880254654:AAE4ahj7KwFh0nct-k7z2Tsqgz1SFexs4FE")
	if err != nil {
		panic(err)
	}
	log.Info("Bot started successfully")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u) // get updates

	for update := range updates { // perform updates
		if update.Message == nil {
			continue
		}

		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			log.InfoF("Received command '%s' from '%s'", update.Message.Text, update.Message.From.FirstName)

			switch update.Message.Text {
			case "/start":
				msg := "Привет, " + update.Message.From.FirstName
				sendMessageToUser(msg, update, *bot)
			case "/about":
				msg := "Cделано на голанге.\nСвязаться с разработчиком:\nhttps://github.com/sputnik5459"
				sendMessageToUser(msg, update, *bot)
			default:
				income_msg := update.Message.Text
				is_valid, mode := validateReq(income_msg)
				if !is_valid {
					sendMessageToUser("Извините, но текст не соответствует шаблону. Пример: ФИС.203.Б", update, *bot)
				} else {
					if mode == 1 {
						income_msg = income_msg + " " + time.Now().Format("02.01.06")
					}
					msg := parseExcelFile(income_msg)
					sendMessageToUser(msg, update, *bot)
				}
			}
		} else {
			sticker_id := "CAACAgIAAxkBAAEBeaNg0lpXpk45fzjRcz3iWVQGnxm3DwACKAADBc7CLQO8N3eoe09PHwQ"
			log.InfoF("Answer 'sticker: %s' was sent to '%s'", sticker_id, update.Message.From.FirstName)
			msg := tgbotapi.NewStickerShare(update.Message.Chat.ID, sticker_id)
			bot.Send(msg)
		}
	}
}

func main() {
	startTelegramBotCycle()
}
