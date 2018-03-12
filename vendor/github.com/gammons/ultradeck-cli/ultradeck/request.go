package ultradeck

const StartAuthRequest = "start_auth"
const StartAuthResponse = "auth_response"

const OpenAuthorizedConnectionRequest = "open_authorized_connection_request"
const OkResponse = "ok"

type Request struct {
	Request string                 `json:"request"`
	Data    map[string]interface{} `json:"data"`
}
