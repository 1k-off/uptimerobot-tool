package main

import (
	"fmt"
	uptimerobot "github.com/bitfield/uptimerobot/pkg"
	"log"
)

// MyUptimerobotClient is a type that extends original uptimerobot.Client type.
type MyUptimerobotClient struct {
	uptimerobot.Client
}

// uptimerobotAccount is a structure for accounts data
type uptimerobotAccount struct {
	Token  string
	Email  string
	Client MyUptimerobotClient
}
type httpScheme struct {
	http  string
	https string
}

var scheme httpScheme = httpScheme{http: "http://", https: "https://"}

// UptimerobotWorkflow is a main function for work with uptimerobot.
// Receives environment and sitelist.
// This function checks if every site from sitelist exists in monitoring accounts and removes monitors from monitoring accounts, that are not exists in sitelist.
func UptimerobotWorkflow(env *Environment, sitelist Sitelist) {
	var (
		uptimeAccount []uptimerobotAccount
		clients []MyUptimerobotClient
	)

	for i, _ := range env.Uptimerobot.Token {
		uptimeAccount = append(uptimeAccount, uptimerobotAccount{Email: env.Uptimerobot.Email[i], Token: env.Uptimerobot.Token[i]})
	}
	for i, _ := range uptimeAccount {
		c := uptimerobot.New(uptimeAccount[i].Token)
		clients = append(clients, MyUptimerobotClient{c})
		uptimeAccount[i].Client = MyUptimerobotClient{c}
	}

	enabledMonitors := getAllMonitors(clients)

	for _, m := range enabledMonitors {
		if checkMonitorShouldBeDeleted(sitelist, m) {
			deleteMonitor(uptimeAccount, m)
			LogInfo.Println("Deleted monitor", m.FriendlyName)
		}
	}
	for _, s := range sitelist {
		if isExists, monitor := isMonitorExists(enabledMonitors, s.WebSiteName); !(isExists) {
			id, email := createNewMonitor(clients, s.WebSiteName, s.Config.Keyword)
			LogInfo.Println("Created monitor", s.WebSiteName, "in account", email, "with ID", id)
		} else {
			if s.Config.Keyword != "" && monitor.Type == 1 { // monitor.Type == 1 is basic http status code monitor. https://github.com/bitfield/uptimerobot/blob/master/pkg/monitor.go
				// We could delete monitor with old type and create new one with new type
				// because we can't update monitor type
				deleteMonitor(uptimeAccount, monitor)
				_, _ = createNewMonitor(clients, s.WebSiteName, s.Config.Keyword)
				LogInfo.Println("Changed monitor", s.WebSiteName, "type from HTTP to keyword.")
			} else if s.Config.Keyword != "" && s.Config.Keyword != monitor.KeywordValue {
				editMonitorKeyword(uptimeAccount, monitor, s.Config.Keyword)
				LogInfo.Println("Changed monitor", monitor.URL, "keyword from", monitor.KeywordValue, "to", s.Config)
			} else if s.Config.Keyword == "" && monitor.Type == 2 {
				deleteMonitor(uptimeAccount, monitor)
				_, _ = createNewMonitor(clients, s.WebSiteName, s.Config.Keyword)
				LogInfo.Println("Changed monitor", s.WebSiteName, "type from keyword to HTTP.")
			}
		}
	}

	//Delete all monitors in all accounts. Need for development cleanup.
	//for _, m := range enabledMonitors {
	//	deleteMonitor(uptimeAccount, m)
	//}
}

// getAllMonitors - method returns all monitors from current client.
func (c MyUptimerobotClient) getAllMonitors() []uptimerobot.Monitor {
	monitors, err := c.AllMonitors()
	if err != nil {
		LogError.Fatalln("Can't get monitors for provided client.", err)
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

// getAlertContact - returns alert contact id by client
func (c MyUptimerobotClient) getAlertContact() string {
	var contacts []uptimerobot.AlertContact
	var id string
	contacts, _ = c.AllAlertContacts()
	for _, contact := range contacts {
		if contact.Type == 11 {
			id = contact.ID
		}
	}
	return id
}

// createNewMonitor is a function to create new monitor for provided url in one of available accounts.
// Receives array of clients and url for monitor, returns created monitor id and email of account in which monitor created.
func createNewMonitor(clients []MyUptimerobotClient, url, keyword string) (int64, string) {
	var (
		id           int64
		email        string
		targetClient MyUptimerobotClient
		m            uptimerobot.Monitor
	)
	if isExists, _ := isMonitorExists(getAllMonitors(clients), url); isExists == false {
		for _, c := range clients {
			if c.freeMonitors() {
				targetClient = c
				break
			}
		}
		if keyword != "" {
			m = uptimerobot.Monitor{
				FriendlyName:  url,
				URL:           scheme.https + url,
				Type:          2,
				KeywordType:   2,
				KeywordValue:  keyword,
				Port:          443,
				AlertContacts: []string{targetClient.getAlertContact()},
			}
		} else {
			m = uptimerobot.Monitor{
				FriendlyName:  url,
				URL:           scheme.https + url,
				Type:          1,
				Port:          443,
				AlertContacts: []string{targetClient.getAlertContact()},
			}
		}
		id, _ = targetClient.CreateMonitor(m)
		account, _ := targetClient.GetAccountDetails()
		email = account.Email
	}
	return id, email
}

// deleteMonitor deletes provided monitor from all accounts.
func deleteMonitor(UptimeAccount []uptimerobotAccount, m uptimerobot.Monitor) {
	for _, a := range UptimeAccount {
		monitors := a.Client.getAllMonitors()
		for _, mon := range monitors {
			if m.ID == mon.ID {
				_ = a.Client.DeleteMonitor(m.ID)
			}
		}
	}
}

// editMonitorKeyword is a function to change monitor keyword.
func (c MyUptimerobotClient) editMonitorKeyword(m uptimerobot.Monitor, newKeywordValue string) {
	r := uptimerobot.Response{}
	data := []byte(fmt.Sprintf("{\"id\": \"%d\",\"keyword_value\": \"%s\"}", m.ID, newKeywordValue))
	if err := c.MakeAPICall("editMonitor", &r, data); err != nil {
		log.Fatal(err)
	}
}

// editMonitorKeyword updates keyword on existing monitor.
func editMonitorKeyword(UptimeAccount []uptimerobotAccount, m uptimerobot.Monitor, newKeywordValue string) {
	for _, a := range UptimeAccount {
		monitors := a.Client.getAllMonitors()
		for _, mon := range monitors {
			if m.ID == mon.ID {
				a.Client.editMonitorKeyword(mon, newKeywordValue)
			}
		}
	}
}
