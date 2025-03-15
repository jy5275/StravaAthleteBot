package common

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var (
	svc *dynamodb.DynamoDB
)

type ActivityHistoryItem struct {
	User_id     string `dynamodbav:"user_id"`
	Activity_id string `dynamodbav:"activity_id"`
}

func initDDBSvc() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc = dynamodb.New(sess)
}

func GetStravaUserIDFromDDB(cmd string) (string, string, string, error) {
	if len(cmd) > MAX_NICKNAME_LEN {
		return "", "", "", nil
	}

	nickname, found := strings.CutPrefix(cmd, "/")
	if !found {
		return "", "", "", nil
	}

	if svc == nil {
		initDDBSvc()
	}
	tableName := "red_book_profile"
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"nickname": {
				S: aws.String(strings.ToLower(nickname)),
			},
		},
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to call DDB GetItem: %w", err)
	}
	if result.Item == nil {
		return "", "", "", nil
	}

	type Item struct {
		Nickname   string
		Profile_id string
		Real_name  string
		Strava_id  string
	}
	var res Item
	err = dynamodbattribute.UnmarshalMap(result.Item, &res)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to unmarshal DDB item: %w", err)
	}

	return nickname, res.Strava_id, res.Real_name, nil
}

func InsertActivityRecord(userID string, activityID string) error {
	if svc == nil {
		initDDBSvc()
	}

	item := ActivityHistoryItem{
		User_id:     userID,
		Activity_id: activityID,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal activity %v, err=%w", activityID, err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("strava_activity_history"),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to insert activity %s to DDB, err=%w", activityID, err)
	}

	return nil
}

func GetAthleteActivityHistoryList(userID string) (map[string]bool, error) {
	if svc == nil {
		initDDBSvc()
	}

	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String("strava_activity_history"),
		KeyConditions: map[string]*dynamodb.Condition{
			"user_id": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(userID),
					},
				},
			},
		},
	}

	resp, err := svc.Query(queryInput)
	if err != nil {
		return nil, err
	}
	var res []ActivityHistoryItem
	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal activity history: %s", err)
	}

	dedupMap := map[string]bool{}
	for _, item := range res {
		dedupMap[item.Activity_id] = true
	}

	return dedupMap, nil
}
