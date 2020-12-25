package uptimerobot_tool

import (
	"encoding/json"
	"fmt"
	uptimerobot "github.com/bitfield/uptimerobot/pkg"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	schemeHttp      = "http"
	schemeHttpFull  = schemeHttp + "://"
	schemeHttps     = "https"
	schemeHttpsFull = schemeHttps + "://"
)

//GetSitelistFromFile retrieves list with sites to check from local file (sitelist.json) and returns all
//sites as array of strings
func GetSitelistFromFile(path string) Sitelist {
	var sitelist Sitelist
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Can't open sitelist file.", err)
	}
	_ = json.Unmarshal([]byte(file), &sitelist)
	return sitelist
}

// IsEmptyString - checks if provided string is empty
func IsEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// freeMonitors - checks if we still can create new monitors in provided account.
func (account Uptimerobot) freeMonitors() bool {
	return len(account.getAllMonitors()) < 50
}

func getWebsiteAlertContactsFromAccount(token, website string) (alertContacts []string) {
	type ReceivedMonitors struct {
		Monitors []struct {
			ID            int    `json:"id"`
			FriendlyName  string `json:"friendly_name"`
			AlertContacts []struct {
				ID    string `json:"id"`
				Value string `json:"value"`
				Type  int    `json:"type"`
			} `json:"alert_contacts"`
		} `json:"monitors"`
	}
	var monitors ReceivedMonitors

	url := "https://api.uptimerobot.com/v2/getMonitors"
	payload := strings.NewReader(fmt.Sprintf("api_key=%s&format=json&alert_contacts=1", token))
	req, _ := http.NewRequest(http.MethodPost, url, payload)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&monitors)
	for _, m := range monitors.Monitors {
		if m.FriendlyName == website {
			for _, ac := range m.AlertContacts {
				alertContacts = append(alertContacts, ac.ID)
			}
		}
	}
	return alertContacts
}

func findMonitorAccount(uptimerobotAccount []Uptimerobot, monitor uptimerobot.Monitor) (account Uptimerobot) {
	for _, a := range uptimerobotAccount {
		accountMonitors := a.getAllMonitors()
		for _, m := range accountMonitors {
			if m.FriendlyName == monitor.FriendlyName {
				return a
			}
		}
	}
	return
}

func findFreeAccount(uptimerobotAccount []Uptimerobot) (account Uptimerobot) {
	for _, a := range uptimerobotAccount {
		if a.freeMonitors() {
			account = a
			break
		}
	}
	return account
}