package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hamoz/uxsdp-facebook/common"
	"github.com/hamoz/uxsdp-facebook/fb/api"
	"github.com/hamoz/uxsdp-facebook/fb/model"
	"github.com/hamoz/uxsdp-facebook/rapidpro"
)

var (
	errUnknownWebHookObject = errors.New("unknown web hook object")
	errNoMessageEntry       = errors.New("there is no message entry")
)

type FacebookHandler struct {
	verifyToken string
	appSecret   string
	accessToken string
	rapidPro    *rapidpro.RapidExtChannel
	facebookApi *api.FacebookApi
}

func New(rapidPro *rapidpro.RapidExtChannel, verifyToken, appSecret, accessToken string) *FacebookHandler {
	return &FacebookHandler{facebookApi: api.New(accessToken), rapidPro: rapidPro, verifyToken: verifyToken, appSecret: appSecret, accessToken: accessToken}
}

// HandleMessenger handles all incoming webhooks from Facebook Messenger.
func (fb *FacebookHandler) HandleIncoming(w http.ResponseWriter, r *http.Request) {
	log.Println("new request")
	if r.Method == http.MethodGet {
		fb.HandleVerification(w, r)
		return
	}

	fb.HandleWebHook(w, r)
}

// HandleVerification handles the verification request from Facebook.
func (fb *FacebookHandler) HandleVerification(w http.ResponseWriter, r *http.Request) {
	if fb.verifyToken != r.URL.Query().Get("hub.verify_token") {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(r.URL.Query().Get("hub.challenge")))
}

// HandleWebHook handles a webhook incoming from Facebook.
func (fb *FacebookHandler) HandleWebHook(w http.ResponseWriter, r *http.Request) {
	//err := utils.Authorize(r, fb.accessToken)
	/*if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
		log.Println("authorize", err)
		return
	}*/
	vars := mux.Vars(r)
	channelId := vars["ChannelId"]
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		log.Println("read webhook body", err)
		return
	}

	wr := model.WebHookRequest{}
	err = json.Unmarshal(body, &wr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		log.Println("unmarshal request", err)
		return
	}

	err = fb.handleWebHookRequest(wr, channelId)
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

func (fb *FacebookHandler) handleWebHookRequest(r model.WebHookRequest, channelId string) error {
	if r.Object != "page" {
		return errUnknownWebHookObject
	}

	for _, we := range r.Entry {
		err := fb.handleWebHookRequestEntry(we, channelId)
		if err != nil {
			return fmt.Errorf("handle webhook request entry: %w", err)
		}
	}

	return nil
}

func (fb *FacebookHandler) handleWebHookRequestEntry(we model.WebHookRequestEntry, channelId string) error {
	if len(we.Messaging) == 0 { // Facebook claims that the arr always contains a single item but we don't trust them :)
		return errNoMessageEntry
	}

	em := we.Messaging[0]

	// message action
	if em.Message != nil {
		log.Printf("text : %s, from : %s, to : %s\n", em.Message.Text, em.Sender.ID, em.Recipient.ID)
		msg := common.RapidMessage{ID: em.Message.Mid, ChannelId: channelId, ChannelType: "Facebook", Text: em.Message.Text,
			Sender:    common.User{ID: em.Sender.ID},
			Recipient: common.User{ID: em.Recipient.ID},
		}
		err := fb.rapidPro.CallApi(context.TODO(), msg)
		//err := ra//fb.handleMessage(em.Sender.ID, em.Message.Text)
		if err != nil {
			return fmt.Errorf("handle message: %w", err)
		}
	}

	return nil
}

func (fb FacebookHandler) HandleOutgoing(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	id := r.PostForm.Get("id")
	from := r.PostForm.Get("from")
	to := r.PostForm.Get("to")
	text := r.PostForm.Get("text")
	log.Printf("id : %s, from : %s, to : %s, text : %s", id, from, to, text)
	fb.facebookApi.Respond(context.TODO(), to, text)
}

/*func (fb *Facebook) handleMessage(recipientID, msgText string) error {
	msgText = strings.TrimSpace(msgText)

	var responseText string
	switch msgText {
	case "hello":
		responseText = "world"
	// @TODO your custom cases
	default:
		responseText = "What can I do for you?"
	}

	return fb.Respond(context.TODO(), recipientID, responseText)
}*/
