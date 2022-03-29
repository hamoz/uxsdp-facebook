package fb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	fb "github.com/hamoz/uxsdp-facebook/fb/model"
	"github.com/valyala/fasthttp"
)

const (
	uriSendMessage = "https://graph.facebook.com/v12.0/me/messages"

	defaultRequestTimeout = 10 * time.Second
)

// https://developers.facebook.com/docs/messenger-platform/send-messages/#messaging_types
const (
	messageTypeResponse = "RESPONSE"
)

var (
	client = fasthttp.Client{}
)

type FacebookApi struct {
	accessToken string
}

func New(acesssToken string) *FacebookApi {
	return &FacebookApi{accessToken: acesssToken}
}

// Respond responds to a user in FB messenger. This includes promotional and non-promotional messages sent inside the 24-hour standard messaging window.
// For example, use this tag to respond if a person asks for a reservation confirmation or an status update.
func (api *FacebookApi) Respond(ctx context.Context, recipientID, msgText string) error {
	return api.CallAPI(ctx, fb.SendMessageRequest{
		MessagingType: messageTypeResponse,
		RecipientID: fb.MessageRecipient{
			ID: recipientID,
		},
		Message: fb.Message{
			Text: msgText,
		},
	})
}

func (api *FacebookApi) CallAPI(ctx context.Context, smr fb.SendMessageRequest) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fmt.Sprintf("%s?access_token=%s", uriSendMessage, api.accessToken))
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Add("Content-Type", "application/json")
	body, err := json.Marshal(&smr)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	req.SetBody(body)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	dl, ok := ctx.Deadline()
	if !ok {
		dl = time.Now().Add(defaultRequestTimeout)
	}

	err = client.DoDeadline(req, res, dl)
	if err != nil {
		return fmt.Errorf("do deadline: %w", err)
	}

	resp := fb.APIResponse{}
	err = json.Unmarshal(res.Body(), &resp)
	if err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("response error: %s", resp.Error.Error())
	}
	if res.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("unexpected rsponse status %d", res.StatusCode())
	}

	return nil
}
