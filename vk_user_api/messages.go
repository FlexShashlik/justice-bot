package vk

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Messages https://new.vk.com/dev/messages
type Messages struct {
	client *VK
}

// Send https://vk.com/dev/messages.send
func (messages *Messages) Send(params RequestParams) (int64, error) {
	resp, err := messages.client.CallMethod("messages.send", params)

	if err != nil {
		return 0, err
	}

	type JSONBody struct {
		MessageID int64 `json:"response"`
	}

	var body JSONBody

	if err := json.Unmarshal(resp, &body); err != nil {
		return 0, err
	}

	return body.MessageID, nil
}

// GetByID https://vk.com/dev/messages.getById
func (messages *Messages) GetByID(params RequestParams) (string, string, error) {
	resp, err := messages.client.CallMethod("messages.getById", params)

	if err != nil {
		return "", "", err
	}

	type CreateAttach struct {
		ID        int64  `json:"id"`
		OwnerID   int64  `json:"owner_id"`
		AccessKey string `json:"access_key"`
	}

	type CreateAttachments struct {
		Type         string       `json:"type"`
		Photo        CreateAttach `json:"photo"`
		Video        CreateAttach `json:"video"`
		Audio        CreateAttach `json:"audio"`
		Doc          CreateAttach `json:"doc"`
		Wall         CreateAttach `json:"wall"`
		Market       CreateAttach `json:"market"`
		Poll         CreateAttach `json:"poll"`
		Graffiti     CreateAttach `json:"graffiti"`
		AudioMessage CreateAttach `json:"audio_message"`
	}

	type CreateItems struct {
		Attachments []CreateAttachments `json:"attachments"`
		FromID      int64               `json:"from_id"`
	}

	type CreateResponse struct {
		Items []CreateItems `json:"items"`
	}

	type JSONBody struct {
		Response CreateResponse `json:"response"`
	}

	var body JSONBody

	if err := json.Unmarshal(resp, &body); err != nil {
		return "", "", err
	}

	var result string

	from := strconv.FormatInt(body.Response.Items[0].FromID, 10)

	for _, item := range body.Response.Items[0].Attachments {
		fmt.Println(item.Type)
		result += item.Type

		switch item.Type {
		case "photo":
			result += strconv.FormatInt(item.Photo.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Photo.ID, 10)
			if item.Photo.AccessKey != "" {
				result += "_" + item.Photo.AccessKey
			}
		case "video":
			result += strconv.FormatInt(item.Video.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Video.ID, 10)
			if item.Video.AccessKey != "" {
				result += "_" + item.Video.AccessKey
			}
		case "audio":
			result += strconv.FormatInt(item.Audio.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Audio.ID, 10)
			if item.Audio.AccessKey != "" {
				result += "_" + item.Audio.AccessKey
			}
		case "doc":
			result += strconv.FormatInt(item.Doc.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Doc.ID, 10)
			if item.Doc.AccessKey != "" {
				result += "_" + item.Doc.AccessKey
			}
		case "wall":
			result += strconv.FormatInt(item.Wall.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Wall.ID, 10)
			if item.Wall.AccessKey != "" {
				result += "_" + item.Wall.AccessKey
			}
		case "market":
			result += strconv.FormatInt(item.Market.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Market.ID, 10)
			if item.Market.AccessKey != "" {
				result += "_" + item.Market.AccessKey
			}
		case "poll":
			result += strconv.FormatInt(item.Poll.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Poll.ID, 10)
			if item.Poll.AccessKey != "" {
				result += "_" + item.Poll.AccessKey
			}
		case "graffiti":
			result += strconv.FormatInt(item.Graffiti.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.Graffiti.ID, 10)
			if item.Graffiti.AccessKey != "" {
				result += "_" + item.Graffiti.AccessKey
			}
		case "audio_message":
			result += strconv.FormatInt(item.AudioMessage.OwnerID, 10)
			result += "_" + strconv.FormatInt(item.AudioMessage.ID, 10)
			if item.AudioMessage.AccessKey != "" {
				result += "_" + item.AudioMessage.AccessKey
			}
		}

		result += ","
	}

	return result, from, nil
}
