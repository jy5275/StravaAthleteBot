package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

func ExtractAthleteDetailFromResp(body io.ReadCloser) (*Athlete, error) {
	rawJSON, err := extractNextDataJSONFromResp(body)
	if err != nil {
		return nil, err
	}
	return parseAthlete(rawJSON)
}

// Only need startlocal field from this
func ExtractActivityDateTimeFromResp(body io.ReadCloser) (string, error) {
	rawJSON, err := extractNextDataJSONFromResp(body)
	if err != nil {
		return "", err
	}
	return parseActivity(rawJSON)
}

func extractNextDataJSONFromResp(body io.ReadCloser) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := html.Parse(buf)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %v", err)
	}

	var extractJSON func(*html.Node) string
	extractJSON = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "__NEXT_DATA__" {
					return strings.TrimSpace(n.FirstChild.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if res := extractJSON(c); res != "" {
				return res
			}
		}
		return ""
	}

	rawJSON := extractJSON(doc)
	if rawJSON == "" {
		return "", fmt.Errorf("__NEXT_DATA__ not found")
	}

	return rawJSON, nil
}

func parseActivity(jsonStr string) (string, error) {
	var rawData struct {
		Props struct {
			PageProps struct {
				Activity struct {
					StartLocal string `json:"startLocal"`
				} `json:"activity"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return rawData.Props.PageProps.Activity.StartLocal, nil
}

func parseAthlete(jsonStr string) (*Athlete, error) {
	var rawData struct {
		Props struct {
			PageProps struct {
				AthleteID   int64 `json:"athleteId"`
				AthleteData struct {
					Athlete struct {
						ID   int64  `json:"id"`
						Name string `json:"name"`
					} `json:"athlete"`
					RecentActivities []*Activity `json:"recentActivities"`
					Stats            struct {
						MonthlyDistance string `json:"monthlyDistance"`
						MonthlyTime     string `json:"monthlyTime"`
					} `json:"stats"`
				} `json:"athleteData"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	athlete := &Athlete{
		ID:               rawData.Props.PageProps.AthleteData.Athlete.ID,
		Name:             rawData.Props.PageProps.AthleteData.Athlete.Name,
		RecentActivities: rawData.Props.PageProps.AthleteData.RecentActivities,
		MonthlyDistance:  rawData.Props.PageProps.AthleteData.Stats.MonthlyDistance,
		MonthlyTime:      rawData.Props.PageProps.AthleteData.Stats.MonthlyTime,
	}

	for _, ac := range athlete.RecentActivities {
		if ac.DetailedType == "Run" {
			ac.Pace = calculatePace(ac.MovingTime, ac.Distance)
		}
	}

	return athlete, nil
}
