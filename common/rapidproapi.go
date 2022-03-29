package common

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	defaultRequestTimeout = 10 * time.Second
)

var (
	client = fasthttp.Client{}
)

type RapidProApi struct {
	Url  string
	Auth string
}

func (rapid *RapidProApi) CallApi(ctx context.Context, msg RapidMessage) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	channelUrl := strings.Replace(rapid.Url, "{ChannelId}", msg.ChannelId, 1)
	log.Printf("url = %s", channelUrl)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(channelUrl)
	req.PostArgs().Add("from", msg.Sender.ID)
	req.PostArgs().Add("text", msg.Text)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	dl, ok := ctx.Deadline()
	if !ok {
		dl = time.Now().Add(defaultRequestTimeout)
	}

	err := client.DoDeadline(req, res, dl)
	if err != nil {
		return fmt.Errorf("do deadline: %w", err)
	}

	//resp := mAPIResponse{}
	//err = json.Unmarshal(res.Body(), &resp)
	if res.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("unexpected response status %d, error :%s", res.StatusCode(), string(res.Body()))
	}

	return nil
}
