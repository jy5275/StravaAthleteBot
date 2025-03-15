package common

import (
	"fmt"
	"os"
	"testing"
)

func TestInsertActivity(t *testing.T) {
	err := InsertActivityRecord("12345", "2025-01-01#12.2km#60:00")
	if err != nil {
		t.Errorf("failed, %s", err)
	}
}

func TestGetAthleteActivityHistoryList(t *testing.T) {
	_, err := GetAthleteActivityHistoryList("12345")
	if err != nil {
		t.Errorf("failed, %s", err)
	}
}

func TestQueryAthlete(t *testing.T) {
	ac, err := QueryAthlete("96951505")
	if err != nil {
		t.Errorf("failed, %s", err)
	}
	t.Log(ac)
}

func TestQueryActivity(t *testing.T) {
	datetime, err := QueryActivityDateTime(13884445800)
	if err != nil {
		t.Errorf("failed, %s", err)
	}
	t.Log(datetime)
}

func TestQueryAthleteV2(t *testing.T) {
	ath, err := QueryAthleteV2("96951505")
	if err != nil {
		t.Errorf("failed, %s", err)
	}
	t.Log(ath)
	for _, ac := range ath.RecentActivities {
		t.Log(*ac)
	}
}

func TestExtractAthleteDetailFromResp(t *testing.T) {
	file, err := os.Open("../samples/zack0315-raw.html")
	if err != nil {
		t.Fatalf("failed to open test HTML file: %v", err)
	}
	defer file.Close()

	athlete, err := ExtractAthleteDetailFromResp(file)
	if err != nil {
		t.Fatalf("ExtractAthleteDetailFromResp failed: %v", err)
	}

	if athlete.ID != 96951505 || athlete.Name != "Zack Wu" {
		t.Errorf("unexpected athlete data: %+v", athlete)
	}

	if len(athlete.RecentActivities) == 0 {
		t.Errorf("expected recent activities, got none")
	}

	fmt.Printf("+%v", athlete)
}
