package uptimerobot_tool

import (
	"encoding/json"
	"fmt"
	uptimerobot "github.com/bitfield/uptimerobot/pkg"
	"log"
	"io/ioutil"
)

type Sitelist []Website

type Website struct {
	WebSiteName string        `json:"web-site-name"`
	Config      WebsiteConfig `json:"config"`
}
type WebsiteConfig struct {
	Keyword     string   `json:"keyword"`
	Contact []string `json:"contact"`
}

type Uptimerobot struct {
	Token string `yaml:"token" json:"token" toml:"token"`
	Email string `yaml:"email" json:"email" toml:"email"`
	Client MyUptimerobotClient
}

// MyUptimerobotClient is a type that extends original uptimerobot.Client type.
type MyUptimerobotClient struct {
	uptimerobot.Client
}

type httpScheme struct {
	http  string
	https string
}

var scheme httpScheme = httpScheme{http: "http://", https: "https://"}

func init() {
	log.SetFlags(0)
}
//GetSitelistFromFile retrieves list with sites to check from local file
// returns all sites as array of strings
func GetSitelistFromFile(path string) Sitelist {
	var sitelist Sitelist
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Can't open sitelist file.", err)
	}
	_ = json.Unmarshal([]byte(file), &sitelist)
	return sitelist
}

// UptimerobotWorkflow is a main function for work with uptimerobot.
// Receives environment and sitelist.
// This function checks if every site from sitelist exists in monitoring accounts and removes monitors from monitoring accounts, that are not exists in sitelist.
func UptimerobotWorkflow(sitelist Sitelist, uptimerobotAccount []Uptimerobot) {
	var (
		clients []MyUptimerobotClient
	)

	for i, _ := range uptimerobotAccount {
		c := uptimerobot.New(uptimerobotAccount[i].Token)
		clients = append(clients, MyUptimerobotClient{c})
		uptimerobotAccount[i].Client = MyUptimerobotClient{c}
	}

	enabledMonitors := getAllMonitors(clients)

	for _, m := range enabledMonitors {
		if checkMonitorShouldBeDeleted(sitelist, m) {
			deleteMonitor(uptimerobotAccount, m)
			log.Println("Deleted monitor", m.FriendlyName)
		}
	}
	for _, s := range sitelist {
		if isExists, monitor := isMonitorExists(enabledMonitors, s.WebSiteName); !(isExists) {
			id, email := createNewMonitor(clients, s)
			log.Printf("Created monitor %s in account %s with: ID %v",s.WebSiteName, email, id)
		} else {
			if s.Config.Keyword != "" && monitor.Type == 1 { // monitor.Type == 1 is basic http status code monitor. https://github.com/bitfield/uptimerobot/blob/master/pkg/monitor.go
				// We could delete monitor with old type and create new one with new type
				// because we can't update monitor type
				deleteMonitor(uptimerobotAccount, monitor)
				_, _ = createNewMonitor(clients, s)
				log.Println("Changed monitor", s.WebSiteName, "type from HTTP to keyword.")
			} else if s.Config.Keyword != "" && s.Config.Keyword != monitor.KeywordValue {
				editMonitorKeyword(uptimerobotAccount, monitor, s.Config.Keyword)
				log.Println("Changed monitor", monitor.URL, "keyword from", monitor.KeywordValue, "to", s.Config)
			} else if s.Config.Keyword == "" && monitor.Type == 2 {
				deleteMonitor(uptimerobotAccount, monitor)
				_, _ = createNewMonitor(clients, s)
				log.Println("Changed monitor", s.WebSiteName, "type from keyword to HTTP.")
			}
		}
	}
}

// getAllMonitors - method returns all monitors from current client.
func (c MyUptimerobotClient) getAllMonitors() []uptimerobot.Monitor {
	monitors, err := c.AllMonitors()
	if err != nil {
		log.Println("Can't get monitors for provided client.", err)
	}
	return monitors
}

// getAllMonitors is a function that returns all monitors from all accounts.
func getAllMonitors(clients []MyUptimerobotClient) []uptimerobot.Monitor {
	var monitors []uptimerobot.Monitor
	for _, c := range clients {
		monitors = append(monitors, c.getAllMonitors()...)
	}
	return monitors
}

// isMonitorExists is a function that checks if website from sitelist exists in provided monitors array.
func isMonitorExists(monitors []uptimerobot.Monitor, url string) (bool, uptimerobot.Monitor) {
	var monitor uptimerobot.Monitor
	for _, m := range monitors {
		if m.URL == scheme.http+url || m.URL == scheme.https+url || m.URL == url {
			monitor = m
			return true, monitor
		}
	}
	return false, monitor
}

// checkMonitorShouldBeDeleted is a function for detecting monitors that is not exists in provided sitelist, but still present in uptimerobot.
func checkMonitorShouldBeDeleted(sitelist Sitelist, m uptimerobot.Monitor) bool {
	for _, s := range sitelist {
		if m.URL == scheme.http+s.WebSiteName || m.URL == scheme.https+s.WebSiteName || m.URL == s.WebSiteName {
			return false
		}
	}
	return true
}

// freeMonitors - checks if we still can create new monitors in provided client (account).
func (c MyUptimerobotClient) freeMonitors() bool {
	return len(c.getAllMonitors()) < 50
}

// getWebsiteAlertContact - returns alert contact id by client
func (w Website) getWebsiteAlertContact(c MyUptimerobotClient) (alertContacts []string) {
	contacts, err := c.AllAlertContacts()
	if err != nil {
		log.Println(err)
	}
	for _, neededContact := range w.Config.Contact {
		for _, contact := range contacts {
			if neededContact == contact.FriendlyName {
				alertContacts = append(alertContacts, contact.ID)
			}
		}
	}
	return alertContacts
}

// createNewMonitor is a function to create new monitor for provided url in one of available accounts.
// Receives array of clients and url for monitor, returns created monitor id and email of account in which monitor created.
func createNewMonitor(clients []MyUptimerobotClient, w Website) (id int64, email string) {
	var (
		targetClient MyUptimerobotClient
		m            uptimerobot.Monitor
	)
	for _, c := range clients {
		if c.freeMonitors() {
			targetClient = c
			break
		}
	}
	if w.Config.Keyword != "" {
		m = uptimerobot.Monitor{
			FriendlyName:  w.WebSiteName,
			URL:           scheme.https + w.WebSiteName,
			Type:          2,
			KeywordType:   2,
			KeywordValue:  w.Config.Keyword,
			Port:          443,
			AlertContacts: w.getWebsiteAlertContact(targetClient),
		}
	} else {
		m = uptimerobot.Monitor{
			FriendlyName:  w.WebSiteName,
			URL:           scheme.https + w.WebSiteName,
			Type:          1,
			Port:          443,
			AlertContacts: w.getWebsiteAlertContact(targetClient),
		}
	}
	id, _ = targetClient.CreateMonitor(m)
	account, _ := targetClient.GetAccountDetails()
	email = account.Email
	return id, email
}

// deleteMonitor deletes provided monitor from all accounts.
func deleteMonitor(UptimeAccount []Uptimerobot, m uptimerobot.Monitor) {
	for _, a := range UptimeAccount {
		monitors := a.Client.getAllMonitors()
		for _, mon := range monitors {
			if m.ID == mon.ID {
				err := a.Client.DeleteMonitor(m.ID)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

// editMonitorKeyword is a function to change monitor keyword.
func (c MyUptimerobotClient) editMonitorKeyword(m uptimerobot.Monitor, newKeywordValue string) {
	r := uptimerobot.Response{}
	data := []byte(fmt.Sprintf("{\"id\": \"%d\",\"keyword_value\": \"%s\"}", m.ID, newKeywordValue))
	if err := c.MakeAPICall("editMonitor", &r, data); err != nil {
		log.Println(err)
	}
}

// editMonitorKeyword updates keyword on existing monitor.
func editMonitorKeyword(UptimeAccount []Uptimerobot, m uptimerobot.Monitor, newKeywordValue string) {
	for _, a := range UptimeAccount {
		monitors := a.Client.getAllMonitors()
		for _, mon := range monitors {
			if m.ID == mon.ID {
				a.Client.editMonitorKeyword(mon, newKeywordValue)
			}
		}
	}
}

// DeleteAllMonitors is a function to delete all monitors from all accounts
func DeleteAllMonitors(uptimerobotAccount []Uptimerobot) {
	var (
		clients []MyUptimerobotClient
	)

	for i, _ := range uptimerobotAccount {
		c := uptimerobot.New(uptimerobotAccount[i].Token)
		clients = append(clients, MyUptimerobotClient{c})
		uptimerobotAccount[i].Client = MyUptimerobotClient{c}
	}

	enabledMonitors := getAllMonitors(clients)

	for _, m := range enabledMonitors {
		deleteMonitor(uptimerobotAccount, m)
		log.Println("Deleted monitor", m.FriendlyName)
	}
}

func GetAlertContacts (UptimeAccount []Uptimerobot, website Website) []string {
	var (
		clients []MyUptimerobotClient
		contacts []string
	)
	for _, a := range UptimeAccount {
		clients = append(clients, a.Client)
	}
	monitors := getAllMonitors(clients)
	for _, m := range monitors {
		if m.FriendlyName == website.WebSiteName {
			contacts = m.AlertContacts
		}
	}
	return contacts
}