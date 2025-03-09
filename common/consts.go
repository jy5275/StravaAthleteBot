package common

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

const (
	TG_OPENAPI_URL     = "https://api.telegram.org/bot"
	STRAVA_ATHLETE_URL = "https://www.strava.com/athletes/"
	MAX_NICKNAME_LEN   = 12

	TG_BOT_TOKEN_ENV_NAME = "TG_BOT_TOKEN"
	CHAT_ID_ENV_NAME      = "CHAT_ID"
)

var (
	OkResp            = events.APIGatewayProxyResponse{StatusCode: http.StatusOK}
	InternalErrorResp = events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}
)
