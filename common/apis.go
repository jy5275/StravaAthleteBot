package common

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Activity struct {
	Name      string
	Date      string
	Type      string
	Duration  string
	Distance  string
	Elevation string
	Pace      string
}

var (
	tz, _ = time.LoadLocation("Asia/Singapore")
)

func extract(activitiesResp io.Reader) ([]Activity, error) {
	doc, err := goquery.NewDocumentFromReader(activitiesResp)
	if err != nil {
		return nil, err
	}

	var activities []Activity
	doc.Find("ol.RecentActivities_recentActivitiesList__HN_hR > li").Each(func(i int, s *goquery.Selection) {
		var activity Activity
		rawDateStr := s.Find("time.RecentActivities_timestamp__pB9a8").Text()
		activity.Date, err = parseDate(rawDateStr)
		if err != nil {
			log.Fatalf("failed to parse activity date field from Strava API: %s", err)
		}
		activity.Name = s.Find("button[data-cy='recent-activity-name']").Text()
		activity.Type = inferActivityType(s)

		s.Find("ul[class^='Stats_listStats__'] li").Each(func(i int, stat *goquery.Selection) {
			label := stat.Find("span[class^='Stat_statLabel__']").Text()
			value := stat.Find("div[class^='Stat_statValue__']").Text()
			switch label {
			case "Time":
				activity.Duration = value
			case "Distance":
				activity.Distance = value
			case "Elevation":
				activity.Elevation = value
			}
		})

		if activity.Type == "Run" {
			activity.Pace = calculatePace(activity.Duration, activity.Distance)
		}

		activities = append(activities, activity)
	})

	return activities, err
}

func QueryUserActivityList(userID string) ([]Activity, error) {
	url := STRAVA_ATHLETE_URL + userID
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch page, status code: %d", resp.StatusCode)
	}

	return extract(resp.Body)
}

func inferActivityType(s *goquery.Selection) string {
	title := s.Find("svg[data-testid='activity-icon'] title").Text()
	if title != "" {
		return title
	}
	return "Unknown"
}

func calculatePace(duration string, distance string) string {
	distance = strings.TrimSpace(strings.Replace(distance, "km", "", -1))
	durationParts := strings.Split(duration, ":")

	if len(durationParts) != 2 && len(durationParts) != 3 {
		return "N/A"
	}

	var totalSeconds int
	if len(durationParts) == 3 {
		hours, _ := strconv.Atoi(durationParts[0])
		minutes, _ := strconv.Atoi(durationParts[1])
		seconds, _ := strconv.Atoi(durationParts[2])
		totalSeconds = hours*3600 + minutes*60 + seconds
	} else {
		minutes, _ := strconv.Atoi(durationParts[0])
		seconds, _ := strconv.Atoi(durationParts[1])
		totalSeconds = minutes*60 + seconds
	}

	dist, err := strconv.ParseFloat(distance, 64)
	if err != nil || dist == 0 {
		return "N/A"
	}

	paceSeconds := int(float64(totalSeconds) / dist)
	minutes := paceSeconds / 60
	seconds := paceSeconds % 60

	return fmt.Sprintf("%d:%02d/km", minutes, seconds)
}

func parseDate(input string) (string, error) {
	today := time.Now().In(tz)
	var parsedTime time.Time
	var err error

	switch strings.ToLower(input) {
	case "today":
		parsedTime = today
	case "yesterday":
		parsedTime = today.AddDate(0, 0, -1)
	default:
		parsedTime, err = time.Parse("January 2, 2006", input)
		if err != nil {
			return "", fmt.Errorf("invalid date format: %s", input)
		}
	}

	return parsedTime.Format("2006-01-02"), nil
}
