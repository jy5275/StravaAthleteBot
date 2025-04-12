package common

import (
	"fmt"
	"log"
	"net/http"
)

func QueryActivityDateTime(activityID int64) (string, error) {
	url := fmt.Sprintf("%s%v", STRAVA_ACTIVITY_URL, activityID)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch page, status code: %d", resp.StatusCode)
	}

	return ExtractActivityDateTimeFromResp(resp.Body)
}

func QueryAthlete(userID string) (*Athlete, error) {
	url := STRAVA_ATHLETE_URL + userID
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch page, status code: %d", resp.StatusCode)
	}

	a, err := ExtractAthleteDetailFromResp(resp.Body)
	if err != nil {
		return nil, err
	}

	// for i, ac := range a.RecentActivities {
	// 	datetime, err := QueryActivityDateTime(ac.ID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	a.RecentActivities[i].DateTime = datetime
	// }

	return a, nil
}
