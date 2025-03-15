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

func HandleCheckStravaActivityUpdate(ctx context.Context, request ScheduleEventMsg) error {
	for _, inputNickname := range request.CheckedNameList {
		userNickname, userID, userRealName, err := GetStravaUserIDFromDDB(inputNickname)
		if err != nil {
			return fmt.Errorf("failed to get user %s's profile from DDB: %w", inputNickname, err)
		}
		if len(userID) == 0 {
			continue // user not exist
		}
		if len(userRealName) == 0 {
			userRealName = userNickname
		}

		activities, err := QueryAthlete(userID)
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
		for _, ac := range activities {
			acKey := getActivityKey(ac)
			if _, ok := dedupMap[acKey]; !ok { // new activity!
				newAc = &ac
				log.Printf("New activity found: %+v\n", ac)
				break
			}
		}

		if newAc == nil {
			continue // no update => don't send tg message for this user
		}

		newAcKey := getActivityKey(*newAc)
		if err = InsertActivityRecord(userID, newAcKey); err != nil {
			log.Println(err)
			return fmt.Errorf("failed to insert new activity item to DDB, err=%w", err)
		}

		log.Printf("Insert activity ok: %s, %s\n", userID, newAcKey)

		replyMsg := fmt.Sprintf("%s just finished a new %s! Duration: %s", userRealName, newAc.Type, newAc.Duration)
		if newAc.Type == "Run" {
			replyMsg += fmt.Sprintf(" | Distance: %s | Pace: %s", newAc.Distance, newAc.Pace)
		}
		replyMsg += " | " + STRAVA_ATHLETE_URL + userID
		botReplyURL := fmt.Sprintf("%s%s/sendMessage?chat_id=%s&text=%s", TG_OPENAPI_URL,
			TG_BOT_TOKEN, CHAT_ID, replyMsg)
		log.Printf("send message ok: %s\n", botReplyURL)

		_, err = http.Get(botReplyURL)
		if err != nil {
			log.Print(err)
			return fmt.Errorf("failed to call tg sendmessage API, err=%w", err)
		}
	}

	return nil
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

func HandleCheckStravaActivityUpdateV2(ctx context.Context, request ScheduleEventMsg) error {
	for _, inputNickname := range request.CheckedNameList {
		_, userID, _, err := GetStravaUserIDFromDDB(inputNickname)
		if err != nil {
			return fmt.Errorf("failed to get user %s's profile from DDB: %w", inputNickname, err)
		}
		if len(userID) == 0 {
			continue // user not exist
		}

		ath, err := QueryAthleteV2(userID)
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
