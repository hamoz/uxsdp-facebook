package fb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Facebook credentials. It's better to store it in your secret storage.
const (
	verifyToken = ""
	appSecret   = ""
	accessToken = ""
)

// errors
var (
	errUnknownWebHookObject = errors.New("unknown web hook object")
	errNoMessageEntry       = errors.New("there is no message entry")
)

// HandleMessenger handles all incoming webhooks from Facebook Messenger.
func HandleMessenger(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		HandleVerification(w, r)
		return
	}

	HandleWebHook(w, r)
}

// HandleVerification handles the verification request from Facebook.
func HandleVerification(w http.ResponseWriter, r *http.Request) {
	if verifyToken != r.URL.Query().Get("hub.verify_token") {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(r.URL.Query().Get("hub.challenge")))
}

// HandleWebHook handles a webhook incoming from Facebook.
func HandleWebHook(w http.ResponseWriter, r *http.Request) {
	err := Authorize(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		log.Println("authorize", err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		log.Println("read webhook body", err)
		return
	}

	wr := WebHookRequest{}
	err = json.Unmarshal(body, &wr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		log.Println("unmarshal request", err)
		return
	}

	err = handleWebHookRequest(wr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal"))
		log.Println("handle webhook request", err)
		return
	}

	// Facebook waits for the constant message to get that everything is OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

func handleWebHookRequest(r WebHookRequest) error {
	if r.Object != "page" {
		return errUnknownWebHookObject
	}

	for _, we := range r.Entry {
		err := handleWebHookRequestEntry(we)
		if err != nil {
			return fmt.Errorf("handle webhook request entry: %w", err)
		}
	}

	return nil
}

func handleWebHookRequestEntry(we WebHookRequestEntry) error {
	if len(we.Messaging) == 0 { // Facebook claims that the arr always contains a single item but we don't trust them :)
		return errNoMessageEntry
	}

	em := we.Messaging[0]

	// message action
	if em.Message != nil {
		err := handleMessage(em.Sender.ID, em.Message.Text)
		if err != nil {
			return fmt.Errorf("handle message: %w", err)
		}
	}

	return nil
}

func handleMessage(recipientID, msgText string) error {
	msgText = strings.TrimSpace(msgText)

	var responseText string
	switch msgText {
	case "hello":
		responseText = "world"
	// @TODO your custom cases
	default:
		responseText = "What can I do for you?"
	}

	return Respond(context.TODO(), recipientID, responseText)
}
