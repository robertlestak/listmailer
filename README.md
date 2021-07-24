# listmailer

Send an email to a list of recipients.

## Build

```bash
# native
go build -o /usr/bin/listmailer *.go

# docker
docker build . -t listmailer
```

## Usage

```bash

# configure SMTP server
cp .env-sample .env

# export environment variables
export $(<.env)

# run in server mode
listmailer -server -port 8080

# or, run in cli mode
listmailer \
	-emails /path/to/emails.txt \
	-subject "Subject" \
	-from "hello@example.com" \
	-body /path/to/the/body.html \
	-output /path/to/output.json \
	-id id-of-campaign-if-resuming 		# this is optional if resuming a stopped campaign
```

## REST API

```golang
r.HandleFunc("/campaigns", HandleCreateCampaign).Methods("POST")
r.HandleFunc("/campaigns/{id}", HandleGetCampaign).Methods("GET")
r.HandleFunc("/campaigns", HandleListCampaigns).Methods("GET")
r.HandleFunc("/campaigns", HandleClearCampaigns).Methods("DELETE")
```

When creating a campaign, use the following JSON stuct:

```json
{
	"Subject": "Subject",
	"From": "From",
	"To": ["hello@example.com", "world@example.com"],
	"Body": "email html body"
}
```