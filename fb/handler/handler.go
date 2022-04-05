package fb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
	"github.com/hamoz/uxsdp-facebook/common"
	api "github.com/hamoz/uxsdp-facebook/fb/api"
	model "github.com/hamoz/uxsdp-facebook/fb/model"
)

var _ common.PlatformHandler = (*facebookHandler)(nil)

var (
	errUnknownWebHookObject = errors.New("unknown web hook object")
	errNoMessageEntry       = errors.New("there is no message entry")
)

type facebookHandler struct {
	verifyToken string
	appSecret   string
	//accessToken string
	facebookApi *api.FacebookApi
	rapidproApi *common.RapidProApi
}

func NewHandler(rapidproApi *common.RapidProApi, verifyToken, appSecret, accessToken string) common.PlatformHandler {
	return &facebookHandler{
		rapidproApi: rapidproApi,
		facebookApi: api.New(accessToken),
		verifyToken: verifyToken,
		appSecret:   appSecret,
		//accessToken: accessToken,
	}
}

// HandleMessenger handles all incoming webhooks from Facebook Messenger.
func (fb *facebookHandler) HandleIncoming(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("incoming message")
	if r.Method == http.MethodGet {
		fb.handleVerification(w, r)
		return
	}
	fb.handleWebHook(w, r)
}

// HandleVerification handles the verification request from Facebook.
func (fb *facebookHandler) handleVerification(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("v code = " + fb.verifyToken)
	if fb.verifyToken != r.URL.Query().Get("hub.verify_token") {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(r.URL.Query().Get("hub.challenge")))
}

// HandleWebHook handles a webhook incoming from Facebook.
func (fb *facebookHandler) handleWebHook(w http.ResponseWriter, r *http.Request) {
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
		log.Err(err)
		return
	}
	log.Info().Msg("<<" + string(body))
	wr := model.WebHookRequest{}
	err = json.Unmarshal(body, &wr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
		log.Err(err)
		return
	}

	err = fb.handleWebHookRequest(wr, channelId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal"))
		log.Error().Msg("cannot handle webhook request," + err.Error())
		return
	}

	// Facebook waits for the constant message to get that everything is OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

func (fb *facebookHandler) handleWebHookRequest(r model.WebHookRequest, channelId string) error {
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

func (fb *facebookHandler) handleWebHookRequestEntry(we model.WebHookRequestEntry, channelId string) error {
	if len(we.Messaging) == 0 { // Facebook claims that the arr always contains a single item but we don't trust them :)
		return errNoMessageEntry
	}
	em := we.Messaging[0]
	msg := common.RapidMessage{ChannelId: channelId, ChannelType: "Facebook",
		Sender:    common.User{ID: em.Sender.ID},
		Recipient: common.User{ID: em.Recipient.ID},
	}
	// message action
	if em.Message != nil {
		msg.ID = em.Message.Mid
		if len(em.Message.Attachments) == 0 {
			msg.Text = em.Message.Text
		} else {
			attachment := em.Message.Attachments[0]
			switch attachment.Type {
			case "location":
				if coordinates, ok := attachment.Payload["coordinates"].(map[string]interface{}); ok {
					lat := coordinates["lat"].(float64)
					long := coordinates["long"].(float64)
					msg.Text = "#location=" + strconv.FormatFloat(lat, 'f', -1, 32) + "," + strconv.FormatFloat(long, 'f', -1, 32)
				}

			case "image", "video", "audio", "file":
				title, _ := attachment.Payload["title"].(string)
				url, _ := attachment.Payload["url"].(string)
				msg.Text = "#type=" + attachment.Type + "#title=" + title +
					"#url=" + url
			}
		}

	} else if em.Postback != nil {
		msg.Text = em.Postback.Payload
	}
	log.Info().Msg(fmt.Sprint("text, from : %s, to : %s\n", msg.Text, msg.Sender.ID, msg.Recipient.ID))
	err := fb.rapidproApi.CallApi(context.TODO(), msg)
	if err != nil {
		return fmt.Errorf("handle message: %w", err)
	}
	return nil
}

func (fb facebookHandler) HandleOutgoing(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	id := r.PostForm.Get("id")
	from := r.PostForm.Get("from")
	to := r.PostForm.Get("to")
	text := r.PostForm.Get("text")
	accessToken := r.PostForm.Get("access_token")
	log.Info().Msg(fmt.Sprint(">> id : %s, from : %s, to : %s, text : %s", id, from, to, text))

	var err error
	smr := model.SendMessageRequest{
		MessagingType: api.MessageTypeResponse,
		RecipientID:   model.MessageRecipient{ID: to},
	}
	var message model.Message
	if err = json.Unmarshal([]byte(text), &message); err != nil {
		//normal text message
		message = map[string]interface{}{"text": text}
	}
	smr.Message = message
	err = fb.facebookApi.CallAPI(context.TODO(), accessToken, smr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Err(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
