package common

import (
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

const (
	TG_OPENAPI_URL           = "https://api.telegram.org/bot"
	STRAVA_ATHLETE_URL       = "https://www.strava.com/athletes/"
	STRAVA_ACTIVITY_URL      = "https://www.strava.com/activities/"
	MAX_NICKNAME_LEN         = 12
	DDB_STRAVA_ID_TABLE_NAME = "strava_id"
)

var (
	OkResp            = events.APIGatewayProxyResponse{StatusCode: http.StatusOK}
	InternalErrorResp = events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}
	CHAT_ID           = os.Getenv("CHAT_ID")
	TG_BOT_TOKEN      = os.Getenv("TG_BOT_TOKEN")
)
