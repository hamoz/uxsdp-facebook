package common

import "net/http"

type ChannelHandler interface {
	HandleIncoming(w http.ResponseWriter, r *http.Request)
	HandleOutgoing(w http.ResponseWriter, r *http.Request)
}
