package common

import (
	"net/http"
)

/* interface for platforms message handler
this handler handles
1.incomming message from the mesaging platform (e.g :FB, Whatsapp, Telegram)
and submit it to rapidpro
2.Outgoing message from rapidpro and submit them to the platform

*/
type PlatformHandler interface {
	HandleIncoming(w http.ResponseWriter, r *http.Request)
	HandleOutgoing(w http.ResponseWriter, r *http.Request)
}
