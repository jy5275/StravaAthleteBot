
# run `aws configure` first 

chk_strava_pack = check_strava_update.zip

all: build push

build: cmd/*.go common/*
	echo "building golang..."
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap cmd/check_strava_update.go
	zip $(chk_strava_pack) bootstrap
	rm bootstrap

push:
	echo "push to lambda..."
	aws lambda update-function-code --function-name check-strava-update --zip-file fileb://$(chk_strava_pack)

test:
	cd common && go test ./... -v

clear:
	rm -f $(chk_strava_pack) bootstrap