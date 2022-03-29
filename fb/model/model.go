package fb

import "fmt"

type (
	// WebHookRequest received from Facebook server on webhook, contains messages, delivery reports and/or postbacks.
	WebHookRequest struct {
		Object string                `json:"object"`
		Entry  []WebHookRequestEntry `json:"entry"`
	}

	// WebHookRequestEntry is an entry in the Facebook web hooks.
	WebHookRequestEntry struct {
		Time      int                          `json:"time"`
		ID        string                       `json:"id"`
		Messaging []WebHookRequestEntryMessage `json:"messaging"`
	}

	// WebHookRequestEntryMessage is a message from user in the Facebook web hook request.
	WebHookRequestEntryMessage struct {
		Timestamp int              `json:"timestamp"`
		Message   *Message         `json:"message"`
		Delivery  *Delivery        `json:"delivery"`
		Postback  *Postback        `json:"postback"`
		Recipient MessageRecipient `json:"recipient"`
		Sender    MessageSender    `json:"sender"`
	}

	// MessageRecipient is the recipient data in the Facebook web hook request.
	MessageRecipient struct {
		ID string `json:"id"`
	}

	// MessageSender is the sender data in the Facebook web hook request.
	MessageSender struct {
		ID string `json:"id"`
	}

	// Message struct for text messaged received from facebook server as part of WebHookRequest struct.
	Message struct {
		Mid        string      `json:"mid,omitempty"`
		Seq        int         `json:"seq,omitempty"`
		Text       string      `json:"text"`
		Attachment *Attachment `json:"attachment,omitempty"`
	}

	// Attachment is the Facebook messenger message attachment. E.g. buttons.
	Attachment struct {
		Type    string            `json:"type"`
		Payload AttachmentPayload `json:"payload"`
	}

	// AttachmentPayload is the Facebook messenger message attachment payload.
	AttachmentPayload struct {
		TemplateType string            `json:"template_type"`
		Text         string            `json:"text"`
		Buttons      AttachmentButtons `json:"buttons"`
	}

	// AttachmentButtons is the Facebook messenger attachment buttons.
	AttachmentButtons []AttachmentButton

	// AttachmentButton is the Facebook messenger attachment button.
	AttachmentButton struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		Payload string `json:"payload"`
	}

	// AttachmentButtonPostbackPayload is the postback payload from the button attachment.
	// The version in object MUST BE ALWAYS PRESENTED. That's allows us to handle the postback correctly in future.
	AttachmentButtonPostbackPayload struct {
		Version int `json:"version"`
		AttachmentButtonPostbackPayloadV1
	}

	// AttachmentButtonPostbackPayloadV1 is the postback payload version 1 from the button attachment.
	AttachmentButtonPostbackPayloadV1 struct {
		Command string `json:"command"`
	}

	// Delivery struct for delivery reports received from Facebook server as part of WebHookRequest struct.
	Delivery struct {
		Mids      []string `json:"mids"`
		Seq       int      `json:"seq"`
		Watermark int      `json:"watermark"`
	}

	// Postback struct for postbacks received from Facebook server  as part of WebHookRequest struct.
	Postback struct {
		Payload string `json:"payload"`
	}

	// SendMessageRequest is a request to send message through FB Messenger
	SendMessageRequest struct {
		MessagingType string           `json:"messaging_type"`
		Tag           string           `json:"tag,omitempty"`
		RecipientID   MessageRecipient `json:"recipient"`
		Message       Message          `json:"message"`
	}

	// APIResponse received from Facebook server after sending the message.
	APIResponse struct {
		MessageID   string    `json:"message_id"`
		RecipientID string    `json:"recipient_id"`
		Error       *APIError `json:"error,omitempty"`
	}

	// APIError received from Facebook server if sending messages failed.
	APIError struct {
		Code      int    `json:"code"`
		FbtraceID string `json:"fbtrace_id"`
		Message   string `json:"message"`
		Type      string `json:"type"`
	}
)

// Error returns Go error object constructed from APIError data.
func (err *APIError) Error() error {
	return fmt.Errorf("FB Error %d: Type %s: %s; FB trace ID: %s", err.Code, err.Type, err.Message, err.FbtraceID)
}
