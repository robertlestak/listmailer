package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	serverPort   string
	campaignID   string
	fromEmail    string
	subject      string
	serverMode   bool
	pathToBody   string
	pathToEmails string
	outputFile   string
)

func init() {
	log.SetLevel(log.DebugLevel)
	flag.StringVar(&serverPort, "port", "8080", "Port to listen on")
	flag.StringVar(&fromEmail, "from", "", "Email address to send from")
	flag.StringVar(&subject, "subject", "", "Subject of email")
	flag.BoolVar(&serverMode, "server", false, "Run in server mode")
	flag.StringVar(&pathToBody, "body", "", "Path to body file")
	flag.StringVar(&pathToEmails, "emails", "", "Path to emails file")
	flag.StringVar(&outputFile, "output", "output.json", "Path to output file")
	flag.StringVar(&campaignID, "id", "", "ID of campaign")

	flag.Parse()
	_, err := CreateEmailClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
}

func HandleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"method": "HandleCreateCampaign",
	})
	l.Debug("Entering")
	defer l.Debug("Exiting")
	c := &Campaign{}
	defer r.Body.Close()
	bd, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.WithFields(log.Fields{
			"error": err,
		}).Error("Error reading body")
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bd, c)
	if err != nil {
		l.WithFields(log.Fields{
			"error": err,
		}).Error("Error unmarshaling body")
		http.Error(w, "Error unmarshaling body", http.StatusBadRequest)
		return
	}
	c.Create()
	j, err := json.Marshal(c)
	if err != nil {
		l.WithFields(log.Fields{
			"error": err,
		}).Error("Error marshaling body")
		http.Error(w, "Error marshaling body", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func HandleGetCampaign(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"method": "HandleGetCampaign",
	})
	l.Debug("Entering")
	defer l.Debug("Exiting")
	c := &Campaign{}
	id := mux.Vars(r)["id"]
	for _, cl := range campaigns {
		if cl.ID == id {
			c = cl
			break
		}
	}
	if c == nil {
		l.WithFields(log.Fields{
			"id": id,
		}).Error("Campaign not found")
		http.Error(w, "Campaign not found", http.StatusNotFound)
		return
	}
	j, err := json.Marshal(c)
	if err != nil {
		l.WithFields(log.Fields{
			"error": err,
		}).Error("Error marshaling body")
		http.Error(w, "Error marshaling body", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func HandleClearCampaigns(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"method": "HandleClearCampaigns",
	})
	l.Debug("Entering")
	defer l.Debug("Exiting")
	campaigns = []*Campaign{}
	j, err := json.Marshal(campaigns)
	if err != nil {
		l.WithFields(log.Fields{
			"error": err,
		}).Error("Error marshaling body")
		http.Error(w, "Error marshaling body", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func HandleListCampaigns(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"method": "HandleListCampaigns",
	})
	l.Debug("Entering")
	defer l.Debug("Exiting")
	j, err := json.Marshal(campaigns)
	if err != nil {
		l.WithFields(log.Fields{
			"error": err,
		}).Error("Error marshaling body")
		http.Error(w, "Error marshaling body", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func server() {
	l := log.WithFields(log.Fields{
		"method": "server",
	})
	l.Debug("Entering")
	defer l.Debug("Exiting")
	port := "8080"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	r := mux.NewRouter()
	r.HandleFunc("/campaigns", HandleCreateCampaign).Methods("POST")
	r.HandleFunc("/campaigns/{id}", HandleGetCampaign).Methods("GET")
	r.HandleFunc("/campaigns", HandleListCampaigns).Methods("GET")
	r.HandleFunc("/campaigns", HandleClearCampaigns).Methods("DELETE")
	log.Infof("Listening on port %s", port)
	http.ListenAndServe(":"+port, r)
}

func cliValidate() error {
	if len(pathToBody) == 0 {
		return fmt.Errorf("No path to body file specified")
	}
	if len(pathToEmails) == 0 {
		return fmt.Errorf("No path to emails file specified")
	}
	if len(outputFile) == 0 {
		return fmt.Errorf("No output file specified")
	}
	return nil
}

func cliCreateCampaign() error {
	if err := cliValidate(); err != nil {
		return err
	}
	c := &Campaign{}
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		o, err := ioutil.ReadFile(outputFile)
		if err != nil {
			return err
		}
		jerr := json.Unmarshal(o, &campaigns)
		if jerr != nil {
			return jerr
		}
		for _, cl := range campaigns {
			if cl.ID == campaignID {
				c = cl
				break
			}
		}
	}

	b, err := ioutil.ReadFile(pathToBody)
	if err != nil {
		return err
	}
	em, err := FileToAddresses(pathToEmails)
	if err != nil {
		return err
	}
	c.Body = string(b)
	c.To = em
	c.Subject = subject
	c.From = fromEmail
	if c.ID == "" {
		c.Create()
	} else {
		c.Resume()
	}
	return nil
}

func cliWatchCampaigns() []*Campaign {
	for {
		time.Sleep(time.Second * 1)
		var running bool
		for _, cl := range campaigns {
			if cl.Running {
				running = true
				break
			}
		}
		if !running {
			return campaigns
		}
	}
}

func cli() {
	l := log.WithFields(log.Fields{
		"method": "cli",
	})
	l.Debug("Entering")
	defer l.Debug("Exiting")
	e := cliValidate()
	if e != nil {
		flag.Usage()
		os.Exit(1)
	}
	if err := cliCreateCampaign(); err != nil {
		log.Fatal(err)
	}
	c := cliWatchCampaigns()
	if len(c) == 0 {
		log.Fatal("No campaigns created")
	}
	jd, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(outputFile, jd, 0644)
	if err != nil {
		log.Fatal(err)
	}
	l.Debug("Campaigns: %v", c)

}

func main() {
	if serverMode {
		server()
		return
	} else {
		cli()
	}
}
