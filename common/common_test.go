package common

import (
	"testing"
)

func TestInsertActivity(t *testing.T) {
	err := InsertActivityRecord("12345", "2025-01-01#12.2km#60:00")
	if err != nil {
		t.Errorf("failed, %s", err)
	}
}

func TestGetUserActivityHistoryList(t *testing.T) {
	_, err := GetUserActivityHistoryList("12345")
	if err != nil {
		t.Errorf("failed, %s", err)
	}
}

func TestQueryUserActivityList(t *testing.T) {
	ac, err := QueryUserActivityList("96951505")
	if err != nil {
		t.Errorf("failed, %s", err)
	}
	t.Log(ac)
}
