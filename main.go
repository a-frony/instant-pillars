package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	sqlx "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	pinapi "github.com/a-frony/go-pinterest" //pinterest bot api
	pincontrollers "github.com/a-frony/go-pinterest/controllers"
	pinmodels "github.com/a-frony/go-pinterest/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api" //telegram bot api
)

//ChatPhotoRow struct is for photos rows from sql query
type ChatPhotoRow struct {
	ID  int
	URL string
}

//NextPageRow struct is for saved next page row from sql query
type NextPageRow struct {
	ID       int
	PinsLink string `db:"pins_link"`
	Cursor   string
}

var cfg map[string]string
var lang map[string]string
var db *sqlx.DB

//CheckErr is the error check
func CheckErr(err error, msg string) {
	if err != nil {
		log.Printf(msg)
		panic(err)
	}
}

//LoadJSON read json from file and put it to map[string]string
func LoadJSON(filename string) (map[string]string, error) {
	output := make(map[string]string)
	file, _ := os.Open(filename)
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&output)
	return output, err
}

//GetPins is a recursive function. Returns array of all pins from all pages
func GetPins(pinterest *pinapi.Client, pinsLink string, optionals *pincontrollers.BoardsPinsFetchOptionals) (*[]pinmodels.Pin, error) {

	// Fetch the Pins on a Board:
	pins, page, err := pinterest.Boards.Pins.Fetch(pinsLink, optionals)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Printf("%s", page)

	//add next page to optionals
	optionals.Cursor = page.Cursor

	//call itself if there are more pages
	if page != nil {
		OutputTemp, err := GetPins(pinterest, pinsLink, optionals)
		if err == nil {
			*(pins) = append(*(pins), *(OutputTemp)...)
		}

		//if rate limit exceeded set up timer for 1 hour and 1 min and countinue
		if err, ok := err.(*pinmodels.PinterestError); ok && err.StatusCode == 429 {
			timer := time.NewTimer(1*time.Hour + 1*time.Minute)
			go func() {
				<-timer.C
				GetPins(pinterest, pinsLink, optionals)
			}()
		}
	}
	return pins, err
}

//LoadPins is the function to proceed loading images from pinterest to DB
func LoadPins(pinsLink string, optionals *pincontrollers.BoardsPinsFetchOptionals) string {
	var output string

	//connect to pinterest api
	pinterest := pinapi.NewClient().RegisterAccessToken(cfg["PinterestToken"])

	// Fetch the Pins on a Board:
	pins, err := GetPins(pinterest, pinsLink, optionals)
	if err != nil {
		output = fmt.Sprintf("%s @%s", lang["UsrErr"], cfg["BotAdmin"])
	} else {
		//load images from pins to database
		cadd := 0
		cdbl := 0
		var exists bool
		query := "INSERT INTO ic_photos (url) VALUES"
		for _, pin := range *pins {
			//before loading we have to prevent duplicates
			err := db.QueryRow("SELECT exists (SELECT id FROM ic_photos WHERE url=$1)", pin.Image.Original.Url).Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				output = fmt.Sprintf("%s @%s", lang["UsrErr"], cfg["BotAdmin"])
				continue
			}
			if exists == true {
				cdbl++
			} else {
				//add new photo to DB
				cadd++
				query = fmt.Sprintf("%s ('%s'),", query, pin.Image.Original.Url)
			}
		}
		//if there are new images
		if cadd > 0 {
			query = strings.TrimSuffix(query, ",")

			//execute the query
			_, err := db.Exec(query)
			if err != nil {
				log.Println(err)
				output = fmt.Sprintf("%s @%s", lang["UsrErr"], cfg["BotAdmin"])
			} else {
				output = fmt.Sprintf("%s: %d", lang["AdmAdded"], cadd)
			}
		}

		//if there are duplicate images
		if cdbl > 0 {
			output = fmt.Sprintf("%s\n%s: %d", output, lang["AdmDuplicates"], cdbl)
		}

		//if there is next page for the future
		if optionals.Cursor != "" {
			//add next page to DB
			_, err := db.Exec("DELETE FROM ic_page")
			if err != nil {
				log.Printf("Cannot wipe page table: %s", err)
			}
			_, err = db.Exec("INSERT INTO ic_page (pins_link, cursor) VALUES ($1, $2)", pinsLink, optionals.Cursor)
			if err != nil {
				log.Printf("Cannot add row to page table: %s", err)
			}
			//pinterest application exceeds limits per hour
			output = fmt.Sprintf("%s\n%s", output, lang["AdmLimitWait"])
		}
	}
	return output
}

//main function
func main() {

	var err error
	//load config from json file
	cfg, err = LoadJSON("config.json")
	CheckErr(err, "Config file doesn't load")

	//load localization from json file
	lang, err = LoadJSON(fmt.Sprintf("i18n/%s.json", cfg["Language"]))
	CheckErr(err, "Language file doesn't load")

	//connect to telegram
	bot, err := tgbotapi.NewBotAPI(cfg["TelegramBotToken"])
	CheckErr(err, "Can't connect to Telegram server")

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	//process any https requests to prevent unneccesary ones
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(fmt.Sprintf("<!DOCTYPE html><html><head><meta http-equiv='refresh' content='0; url=https://t.me/%[1]s'></head><body><p>To chat with the bot please open this link:<a href='https://t.me/%[1]s'>https://t.me/%[1]s</a></p></body></html>", bot.Self.UserName)))
	})
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	//connect to database
	//	db, err = sqlx.Connect("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", cfg["DBUser"], cfg["DBPass"], cfg["DBName"]))
	db, err = sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	CheckErr(err, "Can't connect postgres database")
	defer db.Close()

	//telegram bot check every 60 sec new messages (only for local machine)
	/*	u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates, err := bot.GetUpdatesChan(u)
		CheckErr(err, "Can't connect to Telegram server") */

	//listening for webhooks (for cloud service)
	updates := bot.ListenForWebhook("/" + bot.Token)

	//empty struct for sent message
	var prMsg tgbotapi.Message

	//struct for pinterest loading
	optionals := pincontrollers.BoardsPinsFetchOptionals{
		Cursor: "",
		Limit:  "100",
	}
	var page []NextPageRow
	//Get from DB next page for continue if there is one
	err = db.Select(&page, "SELECT * FROM ic_page")
	if err == nil && len(page) != 0 {
		optionals.Cursor = page[0].Cursor
	}
	loading := false

	//new messages proccessing
	for update := range updates {

		//what has been sent by bot previously
		log.Printf("Sent:[%d] %s", prMsg.MessageID, prMsg.Text)

		//any message proccessing ignoring non-message updates
		if update.Message != nil {

			//whats coming
			log.Printf("Recieve:[%s] %s", update.Message.From.UserName, update.Message.Text)

			//help message to any request
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			//everyones commands processing
			switch update.Message.Command() {
			case "start", "moar":
				//everyones caterpillar proccessing

				//Check database to prevent duplicated photos

				//get all photos except being sent to this chat
				var photos []ChatPhotoRow
				err = db.Select(&photos, "SELECT id, url FROM ic_photos WHERE id NOT IN (SELECT photo_id FROM ic_chats_photos WHERE chat_id=$1)", update.Message.Chat.ID)
				if err != nil {
					log.Println(err)
					msg.Text = fmt.Sprintf("%s @%s", lang["UsrErr"], cfg["BotAdmin"])
				} else {
					if len(photos) == 0 { //if there are no photos
						msg.Text = fmt.Sprintf("%s @%s", lang["UsrErrZero"], cfg["BotAdmin"])
					} else {

						//get random number for random photo
						rand.Seed(time.Now().UnixNano())
						r := rand.Intn(len(photos))

						//add this photo to db log. This photo will not be shown for this user next time
						_, err = db.Exec("INSERT INTO ic_chats_photos (chat_id, photo_id) VALUES ($1, $2)", update.Message.Chat.ID, photos[r].ID)
						if err != nil {
							log.Println(err)
						}

						//make new message with the photo
						photo := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, nil)
						photo.FileID = photos[r].URL
						photo.UseExisting = true
						prMsg, _ = bot.Send(photo)
						continue
					}
				}

			case "help":
				msg.Text = lang["UsrHelp"]
			case "about":
				msg.Text = lang["UsrAbout"]
			}

			//admin commands processing
			if update.Message.From.UserName == cfg["BotAdmin"] {

				if loading == true {
					loading = false

					//parse pinterest url
					url := update.Message.Text
					if !strings.Contains(url, "https://www.pinterest.ru/") {
						msg.Text = lang["AdmWrongURL"]
					} else {
						pinsLink := strings.TrimPrefix(url, "https://www.pinterest.ru/")
						pinsLink = strings.TrimSuffix(pinsLink, "/")

						//load photos
						msg.Text = LoadPins(pinsLink, &optionals)
					}
				}

				switch update.Message.Command() {
				case "load":
					loading = true
					msg.Text = lang["AdmLoad"]
					if optionals.Cursor != "" {
						msg.Text = lang["AdmContinue"]
					}
				case "continue":
					loading = true
					msg.Text = lang["AdmLoad"]
					if optionals.Cursor != "" {
						msg.Text = LoadPins(page[0].PinsLink, &optionals)
					}

				case "help":
					msg.Text = lang["AdmHelp"]
				}
			}

			//send text message
			if msg.Text != "" {
				prMsg, _ = bot.Send(msg)
			}
		}
	}
}
