package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type Campaign struct {
	ID      string
	Subject string
	Body    string
	From    string
	To      []string
	Running bool
	Success []string
	Fail    []string
}

type Email struct {
	From    string
	To      string
	Subject string
	Body    string
	Error   error
}

type EmailClient struct {
	Host string
	Port int
	User string
	Pass string
}

var (
	emailClient EmailClient
	campaigns   []*Campaign
)

func (e *Email) Validate() error {
	l := log.WithFields(log.Fields{
		"action":  "Email Validate",
		"to":      e.To,
		"from":    e.From,
		"subject": e.Subject,
	})
	l.Debug("Validating email")
	if e.From == "" {
		l.Printf("Email validation error=%v", "From is empty")
		e.From = emailClient.User
	}
	if e.From == "" {
		l.Printf("Email validation error=%v", "From is empty")
		return fmt.Errorf("No From address")
	}
	if e.Subject == "" {
		l.Printf("Email validation error=%v", "Subject is empty")
		return fmt.Errorf("No Subject")
	}
	if e.Body == "" {
		l.Printf("Email validation error=%v", "Body is empty")
		return fmt.Errorf("No Body")
	}
	if e.To == "" {
		l.Printf("Email validation error=%v", "To is empty")
		return fmt.Errorf("No To address")
	}
	return nil
}

func (c *Campaign) IsEmailSuccess(to string) bool {
	for _, t := range c.Success {
		if t == to {
			return true
		}
	}
	return false
}

func (e *Email) Send() error {
	l := log.WithFields(log.Fields{
		"action":  "SendEmail",
		"to":      e.To,
		"from":    e.From,
		"subject": e.Subject,
	})
	l.Debug("Sending email")
	verr := e.Validate()
	if verr != nil {
		l.Errorf("Email validation error=%v", verr)
		return verr
	}
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", e.To)
	m.SetHeader("Subject", e.Subject)
	m.SetBody("text/html", e.Body)
	if emailClient.Pass == "" {
		emailClient.User = ""
	}
	d := gomail.NewDialer(
		emailClient.Host,
		emailClient.Port,
		emailClient.User,
		emailClient.Pass,
	)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		l.Printf("d.DialAndSend error=%v", err)
		return err
	}
	l.Debug("Email sent")
	return nil
}

func CreateEmailClientFromEnv() (EmailClient, error) {
	l := log.WithFields(log.Fields{
		"action": "CreateEmailClientFromEnv",
	})

	sp := os.Getenv("EMAIL_PORT")
	port, err := strconv.Atoi(sp)
	if err != nil {
		return EmailClient{}, err
	}
	emailClient = EmailClient{
		Host: os.Getenv("EMAIL_HOST"),
		Port: port,
		User: os.Getenv("EMAIL_USER"),
		Pass: os.Getenv("EMAIL_PASS"),
	}
	l.Debugf("Email client created v=%v", emailClient)
	return emailClient, nil
}

func sendWorker(emails <-chan *Email, resp chan<- *Email) {
	for e := range emails {
		err := e.Send()
		if err != nil {
			e.Error = err
		}
		resp <- e
	}
}

func (c *Campaign) Send() error {
	l := log.WithFields(log.Fields{
		"action":  "Campaign Send",
		"to":      len(c.To),
		"from":    c.From,
		"subject": c.Subject,
	})
	l.Debug("Sending campaign")

	var to []string
	for _, t := range c.To {
		if c.IsEmailSuccess(t) {
			continue
		}
		to = append(to, t)
	}

	emails := make(chan *Email, len(to))
	resp := make(chan *Email, len(to))

	for w := 1; w <= 20; w++ {
		go sendWorker(emails, resp)
	}
	for _, to := range c.To {
		emails <- &Email{
			From:    c.From,
			To:      to,
			Subject: c.Subject,
			Body:    c.Body,
		}
	}
	close(emails)
	for i := 0; i < len(to); i++ {
		r := <-resp
		if r.Error != nil {
			c.Fail = append(c.Fail, r.To)
		} else {
			c.Success = append(c.Success, r.To)
		}
	}
	c.Running = false
	l.Debug("Campaign sent")
	if len(c.Fail) > 0 {
		l.Errorf("Campaign send errors=%v", c.Fail)
		return fmt.Errorf("Failed to send campaign: %v", c.Fail)
	}
	return nil
}

func (c *Campaign) Create() {
	c.ID = uuid.New().String()
	l := log.WithFields(log.Fields{
		"action":  "Campaign Create",
		"to":      len(c.To),
		"from":    c.From,
		"subject": c.Subject,
		"id":      c.ID,
	})
	l.Debug("Creating campaign")
	campaigns = append(campaigns, c)
	c.Running = true
	go c.Send()
	l.Debug("Campaign created")
}

func (c *Campaign) Resume() {
	l := log.WithFields(log.Fields{
		"action":  "Campaign Resume",
		"to":      len(c.To),
		"from":    c.From,
		"subject": c.Subject,
		"id":      c.ID,
	})
	l.Debug("Resuming campaign")
	defer l.Debug("Campaign resumed")
	c.Running = true
	go c.Send()
}
