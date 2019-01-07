package vk

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-resty/resty"
)

const (
	longPollInstallFlags = 2
	longPollNewMessage   = 4
)

// Message's flags
const (
	FlagMessageUnread = 1 << iota
	FlagMessageOutBox
	FlagMessageReplied
	FlagMessageImportant
	FlagMessageChat
	FlagMessageFriends
	FlagMessageSpam
	FlagMessageDeleted
	FlagMessageFixed
	FlagMessageMedia
)

//FlagsInstaller struct
type FlagsInstaller struct {
	MessageID, Flags, PeerID int64
}

//EventInstallFlags delegate
type EventInstallFlags func(*FlagsInstaller)

// LPMessage struct
type LPMessage struct {
	ID, Flags, PeerID, Timestamp int64
	Subject, Text                string
	Attachments                  string
	FromID                       string
}

// EventNewMessage delegate
type EventNewMessage func(*LPMessage)

type longPoll struct {
	client *VK

	chanInstallFlags  chan *FlagsInstaller
	eventInstallFlags EventInstallFlags

	chanNewMessage  chan *LPMessage
	eventNewMessage EventNewMessage

	data struct {
		server string
		key    string
		ts     int64
	}
}

func (lp *longPoll) update() error {
	fmt.Println("First attempt..")
	resp, err := lp.client.CallMethod("messages.getLongPollServer", RequestParams{
		"use_ssl":  "0",
		"need_pts": "0",
	})

	for fail := err != nil; fail; fail = (err != nil) {
		fmt.Println("Try one more time...")
		resp, err = lp.client.CallMethod("messages.getLongPollServer", RequestParams{
			"use_ssl":  "0",
			"need_pts": "0",
		})
	}

	fmt.Println("Success!")

	type JSONBody struct {
		Response struct {
			Server string `json:"server"`
			Key    string `json:"key"`
			Ts     int64  `json:"ts"`
		} `json:"response"`
	}

	var body JSONBody

	if err := json.Unmarshal(resp, &body); err != nil {
		return err
	}

	lp.data.server = body.Response.Server
	lp.data.key = body.Response.Key
	lp.data.ts = body.Response.Ts

	fmt.Println("Server " + lp.data.server + "\n" +
		"Key " + lp.data.key + "\n" +
		"TS " + strconv.FormatInt(lp.data.ts, 10) + "\n")

	return nil
}

func (lp *longPoll) process() {
	resp, err := resty.R().
		SetQueryParams(RequestParams{
			"act":  "a_check",
			"key":  lp.data.key,
			"ts":   strconv.FormatInt(lp.data.ts, 10),
			"wait": "25",
			"mode": "2",
		}).
		Get("https://" + lp.data.server)

	if err != nil {
		lp.client.Log("[Error] longPoll::process:", err.Error(), "WebResponse:", string(resp.Body()))
		return
	}

	type jsonBody struct {
		Failed  int64           `json:"failed"`
		Ts      int64           `json:"ts"`
		Updates [][]interface{} `json:"updates"`
	}

	var body jsonBody

	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		lp.client.Log("[Error] longPoll::process:", err.Error(), "WebResponse:", string(resp.Body()))
		return
	}

	switch body.Failed {
	case 0:
		for _, update := range body.Updates {
			updateID := update[0].(float64)

			switch updateID {
			case longPollInstallFlags:
				flags := new(FlagsInstaller)
				flags.MessageID = int64(update[1].(float64))
				flags.Flags = int64(update[2].(float64))
				flags.PeerID = int64(update[3].(float64))

				fmt.Println("Message Id:", flags.MessageID, "Flags:", flags.Flags, "Peer ID:", flags.PeerID)

				lp.chanInstallFlags <- flags
			case longPollNewMessage:
				message := new(LPMessage)

				message.ID = int64(update[1].(float64))
				message.Flags = int64(update[2].(float64))
				message.PeerID = int64(update[3].(float64))
				message.Timestamp = int64(update[4].(float64))
				message.Subject = update[5].(string)
				message.Text = update[6].(string)

				attachments := make(map[string]string)

				for key, value := range update[7].(map[string]interface{}) {
					attachments[key] = value.(string)
				}

				if len(attachments) >= 1 {
					message.FromID = attachments["from"]
				}

				fmt.Println("Id", strconv.FormatInt(message.ID, 10),
					"Flags", strconv.FormatInt(message.Flags, 10),
					"PeerID", strconv.FormatInt(message.PeerID, 10),
					"Text", message.Text, "FromID", message.FromID)

				fmt.Println(attachments, "Lenght:", len(attachments))

				for i := 0; i < len(attachments)/2; i++ {
					message.Attachments += attachments["attach"+strconv.FormatInt(int64(i+1), 10)+"_type"]

					message.Attachments += attachments["attach"+strconv.FormatInt(int64(i+1), 10)]

					if i != len(attachments)/2-1 {
						message.Attachments += ","
					}
				}

				fmt.Println(message.Attachments)

				lp.chanNewMessage <- message
			}
		}

		lp.data.ts = body.Ts
	case 1:
		lp.data.ts = body.Ts
		lp.client.Log("ts updated")
	case 2, 3:
		if err := lp.update(); err != nil {
			lp.client.Log("Longpoll update error:", err.Error())
			return
		}
		lp.client.Log("Longpoll data updated")
	}

	lp.process()
}

func (lp *longPoll) processEvents() {
	for {
		select {
		case flags := <-lp.chanInstallFlags:
			if lp.eventInstallFlags != nil {
				lp.eventInstallFlags(flags)
			}
		case message := <-lp.chanNewMessage:
			if lp.eventNewMessage != nil {
				lp.eventNewMessage(message)
			}
		}
	}
}
