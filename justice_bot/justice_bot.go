package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
	"vk_user_api"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

//Message struct
type Message struct {
	ID                bson.ObjectId `json:"id" bson:"_id,omitempty"`
	MessageID, PeerID int64
	Text              string
	Attachments       string
	FromID            string
}

//Query struct
type Query struct {
	MessageID int64
}

func main() {
	const MaxUint = ^uint(0)
	const MaxInt = int64(MaxUint >> 1)

	api := vk.New("ru")
	//Get token - https://oauth.vk.com/authorize?client_id={YOUR-APP-ID}&display=page&redirect_uri=https://oauth.vk.com/blank.html&scope=offline+messages&response_type=token&v=5.92&state=123456
	err := api.Init("{TOKEN}")

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("First attempt to connect db..")
	session, err := mgo.Dial("127.0.0.1:27017")

	for fail := err != nil; fail; fail = (err != nil) {
		fmt.Println("Try one more time...")
		session, err = mgo.Dial("127.0.0.1:27017")
	}

	fmt.Println("Success!")

	c := session.DB("bot").C("messages")

	/*fmt.Println(api.Messages.GetByID(vk.RequestParams{
		"message_ids": "859960",
	}))*/

	api.OnNewMessage(func(msg *vk.LPMessage) {
		//Add message to collection
		message := new(Message)

		message.MessageID = msg.ID
		message.PeerID = msg.PeerID
		message.Text = msg.Text
		message.Attachments, message.FromID, err = api.Messages.GetByID(vk.RequestParams{
			"message_ids": strconv.FormatInt(msg.ID, 10),
		})

		c.Insert(message)

		//Check if count of docs in collection >10k then delete 5k first
		count, _ := c.Count()
		messages := []Message{}

		if count > 10000 {
			c.Find(nil).Limit(5000).All(&messages)
			for i := range messages {
				c.RemoveId(messages[i].ID)
			}
		}
	})

	api.OnInstallFlags(func(flags *vk.FlagsInstaller) {
		if flags.Flags&vk.FlagMessageDeleted == 128 {
			fmt.Println("Message deleted!", "Id:", flags.MessageID)

			message := new(Message)
			query := new(Query)
			query.MessageID = flags.MessageID

			err := c.Find(query).One(&message)

			if err == nil {
				if len(message.Text) < 1800 {
					fmt.Println("Result:", message)

					if message.FromID != "{YOUR-ID}" {
						api.Messages.Send(vk.RequestParams{
							"peer_id":    strconv.FormatInt(message.PeerID, 10),
							"message":    "Удалено сообщение от *id" + message.FromID + ".\n" + "Текст сообщения:\n\"" + message.Text + "\"",
							"attachment": message.Attachments,
							"random_id":  strconv.FormatInt(rand.Int63n(MaxInt), 10),
						})
					}
				} else {
					if message.FromID != "{YOUR-ID}" {
						var j int64
						j = 1

						offset := 1000

						for i := 0; i < len(message.Text)-offset; i += offset {
							api.Messages.Send(vk.RequestParams{
								"peer_id":    strconv.FormatInt(message.PeerID, 10),
								"message":    "Удалено сообщение от *id" + message.FromID + ".\n" + "Текст сообщения(" + strconv.FormatInt(j, 10) + " часть):\n\"" + message.Text[i:i+offset] + "\"",
								"attachment": message.Attachments,
								"random_id":  strconv.FormatInt(rand.Int63n(MaxInt), 10),
							})

							time.Sleep(3 * time.Second)
							j++
						}
					}
				}

				time.Sleep(1 * time.Second)
			}
		}
	})

	api.RunLongPoll()
}

