package telebot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

func api_GET(method string, token string, params url.Values) ([]byte, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s?%s",
		token, method, params.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return json, nil
}

func api_POST(method, token, name, path string, params url.Values) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(name, filepath.Base(path))
	if err != nil {
		return []byte{}, err
	}

	if _, err = io.Copy(part, file); err != nil {
		return []byte{}, err
	}

	for field, values := range params {
		if len(values) > 0 {
			writer.WriteField(field, values[0])
		}
	}

	if err = writer.Close(); err != nil {
		return []byte{}, err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return json, nil
}

func api_getMe(token string) (User, error) {
	me_json, err := api_GET("getMe", token, url.Values{})
	if err != nil {
		return User{}, err
	}

	var bot_info struct {
		Ok          bool
		Result      User
		Description string
	}

	err = json.Unmarshal(me_json, &bot_info)
	if err != nil {
		return User{}, err
	}

	if bot_info.Ok {
		return bot_info.Result, nil
	} else {
		return User{}, AuthError{bot_info.Description}
	}
}

func api_getUpdates(token string, offset int, updates chan<- Update) error {
	params := url.Values{}
	params.Set("offset", strconv.Itoa(offset))
	updates_json, err := api_GET("getUpdates", token, params)
	if err != nil {
		return err
	}

	var updates_recieved struct {
		Ok          bool
		Result      []Update
		Description string
	}

	err = json.Unmarshal(updates_json, &updates_recieved)
	if err != nil {
		return err
	}

	if !updates_recieved.Ok {
		return FetchError{updates_recieved.Description}
	}

	for _, update := range updates_recieved.Result {
		updates <- update
	}

	return nil
}

func api_sendMessage(token string, recipient User, text string) error {
	params := url.Values{}
	params.Set("chat_id", strconv.Itoa(recipient.Id))
	params.Set("text", text)
	response_json, err := api_GET("sendMessage", token, params)
	if err != nil {
		return err
	}

	var response_recieved struct {
		Ok          bool
		Description string
	}

	err = json.Unmarshal(response_json, &response_recieved)
	if err != nil {
		return err
	}

	if !response_recieved.Ok {
		return SendError{response_recieved.Description}
	}

	return nil
}

func api_forwardMessage(token string, recipient User, message Message) error {
	params := url.Values{}
	params.Set("chat_id", strconv.Itoa(recipient.Id))
	params.Set("from_chat_id", strconv.Itoa(message.Origin().Id))
	params.Set("message_id", strconv.Itoa(message.Id))

	response_json, err := api_GET("forwardMessage", token, params)
	if err != nil {
		return err
	}

	var response_recieved struct {
		Ok          bool
		Description string
	}

	err = json.Unmarshal(response_json, &response_recieved)
	if err != nil {
		return err
	}

	if !response_recieved.Ok {
		return SendError{response_recieved.Description}
	}

	return nil
}

func api_sendPhoto(token string, recipient User, photo *Photo) error {
	params := url.Values{}
	params.Set("chat_id", strconv.Itoa(recipient.Id))
	params.Set("caption", photo.Caption)

	var response_json []byte
	var err error

	if photo.Exists() {
		params.Set("photo", photo.FileId)
		response_json, err = api_GET("sendPhoto", token, params)
	} else {
		response_json, err = api_POST("sendPhoto", token, "photo",
			photo.filename, params)
	}

	if err != nil {
		return err
	}

	var response_recieved struct {
		Ok          bool
		Result      Message
		Description string
	}

	err = json.Unmarshal(response_json, &response_recieved)
	if err != nil {
		return err
	}

	if !response_recieved.Ok {
		return SendError{response_recieved.Description}
	}

	thumbnails := &response_recieved.Result.Photo
	photo.File = (*thumbnails)[len(*thumbnails)-1].File

	return nil
}
