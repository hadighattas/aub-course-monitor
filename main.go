package main

import (
	"fmt"
	"io"
	"time"

	"log"

	"os"

	"regexp"

	"strconv"

	"github.com/gocolly/colly"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	Trace *log.Logger

	Info *log.Logger

	Warning *log.Logger

	Error *log.Logger
)

const ERROR int = 999

func Init(

	traceHandle io.Writer,

	infoHandle io.Writer,

	warningHandle io.Writer,

	errorHandle io.Writer) {

	Trace = log.New(traceHandle,

		"TRACE: ",

		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,

		"INFO: ",

		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,

		"WARNING: ",

		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,

		"ERROR: ",

		log.Ldate|log.Ltime|log.Lshortfile)

}

func main() {

	Init(os.Stdout, os.Stdout, os.Stdout,
		os.Stderr)

	var termIn = "201910"

	monitorChannel := make(chan MonitorMessage)

	addChannel := make(chan AddMessage)

	exit := make(chan string)

	// courseArray = append(courseArray, course{"11453", []string{"fkjeio"}})

	bot, err := tgbotapi.NewBotAPI(telegramBotKey)

	if err != nil {

		log.Panic(err)

	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	go telegramMessageHandler(*bot, termIn, addChannel)

	go courseNotifier(*bot, termIn, monitorChannel, addChannel)

	// fmt.Print(<-addChannel)

	// fmt.Print(<-monitorChannel)

	// fmt.Print(<-monitorChannel)

	// fmt.Print(<-monitorChannel)

	// fmt.Print(<-monitorChannel)

	fmt.Print(<-exit)

}

func telegramMessageHandler(bot tgbotapi.BotAPI, termIn string, addChannel chan AddMessage) {

	authenticatingChatIds := []int64{}

	addingChatIds := []int64{}

	authenticatedChatIds := []int64{}

	startRegex, _ := regexp.Compile("/start")

	addRegex, _ := regexp.Compile("/add")

	authenticateRegex, _ := regexp.Compile("/authenticate")

	crnRegex, _ := regexp.Compile("\\b\\d{5}\\b")

	keyRegex, _ := regexp.Compile("\\b\\d{6}\\b")

	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {

			continue

		}

		chatID := update.Message.Chat.ID

		messageText := update.Message.Text

		if startRegex.MatchString(update.Message.Text) {

			msg := tgbotapi.NewMessage(chatID, "Welcome")

			bot.Send(msg)

		} else if authenticateRegex.MatchString(update.Message.Text) {

			if indexOfInt64(authenticatingChatIds, chatID) == -1 && indexOfInt64(authenticatingChatIds, chatID) == -1 {

				authenticatingChatIds = append(authenticatingChatIds, chatID)

			}

			msg := tgbotapi.NewMessage(chatID, "Please enter the key provided by the admin")

			bot.Send(msg)

		} else if addRegex.MatchString(update.Message.Text) {

			if indexOfInt64(authenticatedChatIds, chatID) != -1 {

				if indexOfInt64(addingChatIds, chatID) == -1 {

					addingChatIds = append(addingChatIds, chatID)

				}

				msg := tgbotapi.NewMessage(chatID, "Enter the crn of the course you wish to monitor")

				bot.Send(msg)

			} else {

				msg := tgbotapi.NewMessage(chatID, "Please authenticate first.\nContact the admin for a key.")

				bot.Send(msg)

			}

		} else if keyRegex.MatchString(update.Message.Text) {

			chatIDIndex := indexOfInt64(authenticatingChatIds, chatID)

			if chatIDIndex != -1 {

				keyIndex := indexOfString(keys, messageText)

				if keyIndex != -1 {

					keys = removeElementString(keys, keyIndex)

					authenticatingChatIds = removeElementInt64(authenticatingChatIds, chatIDIndex)

					authenticatedChatIds = append(authenticatedChatIds, chatID)

					msg := tgbotapi.NewMessage(chatID, "You have successfully authenticated. Use /add to add a course monitor.")

					bot.Send(msg)

				} else {

					msg := tgbotapi.NewMessage(chatID, "AUTH FAIL\nPlease contact admin.")

					bot.Send(msg)

				}

			}

		} else if crnRegex.MatchString(update.Message.Text) {

			chatIDIndex := indexOfInt64(addingChatIds, chatID)

			if chatIDIndex != -1 {

				getCourseCapacity(termIn, messageText,

					func(callbackValue int) {

						if callbackValue != ERROR {

							addingChatIds = removeElementInt64(addingChatIds, chatIDIndex)

							addChannel <- AddMessage{messageText, chatID}

							msg := tgbotapi.NewMessage(chatID, "You are now monitoring "+messageText+".")

							bot.Send(msg)

						} else {

							msg := tgbotapi.NewMessage(chatID, "Invalid CRN: "+messageText+".")

							bot.Send(msg)

						}

					})

			} else {

				msg := tgbotapi.NewMessage(chatID, "Use /add to add a course to monitor.")

				bot.Send(msg)

			}

		} else {

			msg := tgbotapi.NewMessage(chatID, "Invalid message.")

			bot.Send(msg)

		}

	}

}

func courseNotifier(bot tgbotapi.BotAPI, termIn string, monitorChannel chan MonitorMessage, addChannel chan AddMessage) {

	var courseArray []course

	for {
		select {

		case addMessage, ok := <-addChannel:

			if ok {
				Trace.Println(addMessage)

				done := false

				for _, v := range courseArray {

					if v.crn == addMessage.crn {

						v.telegramIds = append(v.telegramIds, addMessage.telegramID)

						done = true
					}
				}

				if !done {

					courseArray = append(courseArray, course{addMessage.crn, []int64{addMessage.telegramID}})

					go trackCourseCapacity(termIn, addMessage.crn, monitorChannel)

					done = true
				}

			} else {

			}

		case monitorMessage, ok := <-monitorChannel:

			if ok {
				for _, course := range courseArray {

					if course.crn == monitorMessage.crn {
						for _, telegramID := range course.telegramIds {
							msg := tgbotapi.NewMessage(telegramID, "There is a new spot in "+monitorMessage.crn+".\nHurry there are "+strconv.Itoa(monitorMessage.spotsLeft)+" spots left.")
							bot.Send(msg)
						}
					}
				}
			} else {

			}

		}
	}

}

func trackCourseCapacity(termIn string, crn string, monitorChannel chan MonitorMessage) {
	capacity := 0

	for {

		getCourseCapacity(termIn, crn, func(callbackValue int) {

			if callbackValue != ERROR {

				if callbackValue > capacity {
					monitorChannel <- MonitorMessage{crn, callbackValue}
				}

				capacity = callbackValue
			} else {
				Error.Printf("CRN: %s TERM_IN: %s", crn, termIn)
			}

		})

		time.Sleep(1000 * time.Millisecond)
	}
}

func getCourseCapacity(termIn string, crn string, callback Callback) {

	c := colly.NewCollector(

		colly.AllowedDomains("www-banner.aub.edu.lb"),
	)

	index := 0

	c.OnHTML(".datadisplaytable>tbody>tr>td:nth-child(4)",
		func(e *colly.HTMLElement) {

			if index == 0 {

				capacity, err := strconv.Atoi(e.Text)

				if err != nil {

					callback(ERROR)

				} else {

					callback(capacity)

				}

				index++

			}

		})

	c.OnHTML(".errortext", func(e *colly.HTMLElement) {

		Error.Println("Capacity parsing failed")

		callback(ERROR)

	})

	c.OnRequest(func(r *colly.Request) {

		// Trace.Println("Visiting", r.URL)

	})

	c.Visit("https://www-banner.aub.edu.lb/pls/weba/bwckschd.p_disp_detail_sched?term_in=" + termIn + "&crn_in=" + crn)

}
