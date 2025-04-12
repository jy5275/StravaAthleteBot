package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ScheduleEventMsg struct {
	CheckedNameList []string `json:"checked_name_list"`
}

func getActivityKey(ac Activity) string {
	acKeyWithSpace := fmt.Sprintf("%s#%s#%s#%s", ac.Date, ac.Type, ac.Duration, ac.Distance)
	acKey := strings.Replace(acKeyWithSpace, " ", "", -1)
	return acKey

}

func sendTGMsg(ath *Athlete, newAc *Activity) error {
	replyMsgMultiline := ath.Name + " just finished a new " + newAc.DetailedType
	replyMsgMultiline = replyMsgMultiline + `!
Duration: ` + newAc.MovingTime
	if newAc.DetailedType == "Run" {
		replyMsgMultiline += " | Distance: " + newAc.Distance + " | Pace: " + newAc.Pace
	}
	replyMsgMultiline = replyMsgMultiline + `.
` + ath.Name + "'s workout stat this month: " + ath.MonthlyTime + ", " + ath.MonthlyDistance
	replyMsgMultiline = replyMsgMultiline + `.
` + STRAVA_ACTIVITY_URL + fmt.Sprint(newAc.ID)

	botReplyURL := TG_OPENAPI_URL + TG_BOT_TOKEN + "/sendmessage"
	data := url.Values{}
	data.Set("chat_id", CHAT_ID)
	data.Set("text", replyMsgMultiline)

	resp, err := http.Post(botReplyURL, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Printf("Response Status: %s, err: %v", resp.Status, err)
		return fmt.Errorf("failed to call tg sendmessage API, err=%w", err)
	}
	defer resp.Body.Close()
	log.Printf("send message ok: %s\n", replyMsgMultiline)

	return nil
}

func HandleCheckStravaActivityUpdate(ctx context.Context, request ScheduleEventMsg) error {
	for _, inputNickname := range request.CheckedNameList {
		_, userID, _, err := GetStravaUserIDFromDDB(inputNickname)
		if err != nil {
			return fmt.Errorf("failed to get user %s's profile from DDB: %w", inputNickname, err)
		}
		if len(userID) == 0 {
			continue // user not exist
		}

		ath, err := QueryAthlete(userID)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("failed to query activities of user %s, err=%w", userID, err)
		}

		dedupMap, err := GetAthleteActivityHistoryList(userID)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("failed to get user %s's activity history from DDB, err=%w", userID, err)
		}

		var newAc *Activity
		for _, ac := range ath.RecentActivities {
			if _, ok := dedupMap[fmt.Sprint(ac.ID)]; !ok { // new activity!
				newAc = ac
				log.Printf("New activity found: %+v\n", *ac)
				break
			}
		}

		if newAc == nil {
			continue // no update => don't send tg message for this user
		}

		if err = InsertActivityRecord(userID, fmt.Sprint(newAc.ID)); err != nil {
			log.Println(err)
			return fmt.Errorf("failed to insert new activity item to DDB, err=%w", err)
		}

		log.Printf("Insert activity ok: %s, %+v\n", userID, *newAc)

		err = sendTGMsg(ath, newAc)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	return nil
}
