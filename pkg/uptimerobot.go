package uptimerobot_tool

import (
	uptimerobot "github.com/bitfield/uptimerobot/pkg"
	"log"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ProcessMonitors is an entrypoint for uptimerobot tool. It receives an array of uptimerobot accounts and sitelist for processing.
func ProcessMonitors(uptimerobotAccount []Uptimerobot, sitelist Sitelist) {
	var (
		enabledMonitors []uptimerobot.Monitor
	)
	getUptimerobotAccountsInfo(uptimerobotAccount)

	for _, a := range uptimerobotAccount {
		enabledMonitors = append(enabledMonitors, a.getAllMonitors()...)
	}

	for _, m := range enabledMonitors {
		if monitorShouldBeDeleted(sitelist, m) {
			account := findMonitorAccount(uptimerobotAccount, m)
			account.deleteMonitor(m)
		}
	}

	for _, w := range sitelist {
		if isExist, monitor := w.isMonitorExists(enabledMonitors); !(isExist) {
			account := findFreeAccount(uptimerobotAccount)
			account.createNewMonitor(w)
		} else {
			account := findMonitorAccount(uptimerobotAccount, monitor)
			if !(w.isMonitorEqualToWebsite(monitor, account)) {
				account.deleteMonitor(monitor)
				account.createNewMonitor(w)
			}
		}
	}
}
// getUptimerobotAccountsInfo receives an array of uptimerobot accounts and creates new client for each account.
func getUptimerobotAccountsInfo(account []Uptimerobot) {
	for i := range account {
		account[i].Client = uptimerobot.New(account[i].Token)
	}
}

// getAllMonitors - returns all monitors from provided account.
func (account Uptimerobot) getAllMonitors() []uptimerobot.Monitor {
	monitors, err := account.Client.AllMonitors()
	if err != nil {
		log.Printf("Can't get monitors for account %s. %e", account.Email, err)
	}
	return monitors
}

// monitorShouldBeDeleted is a function for detecting monitors that is not exists in provided sitelist,
// but still present in uptimerobot.
func monitorShouldBeDeleted(sitelist Sitelist, m uptimerobot.Monitor) bool {
	for _, s := range sitelist {
		if m.URL == schemeHttpsFull+s.WebSiteName || m.URL == schemeHttpFull+s.WebSiteName || m.URL == s.WebSiteName {
			return false
		}
	}
	return true
}

// deleteMonitor is a method for deleting provided monitor from uptimerobot account.
func (account Uptimerobot) deleteMonitor(m uptimerobot.Monitor) {
	err := account.Client.DeleteMonitor(m.ID)
	if err != nil {
		log.Printf("Error while deleting monitor %s from account %s. %e", m.FriendlyName, account.Email, err)
	} else {
		log.Printf("Deleted monitor %v from account %v.", m.FriendlyName, account.Email)
	}
}

// isMonitorExists receives all existing monitors for all accounts and checks if monitor for provided website exists.
// Returns true and monitor (if exists) or false and empty monitor.
func (website Website) isMonitorExists(monitors []uptimerobot.Monitor) (bool, uptimerobot.Monitor) {
	for _, m := range monitors {
		if m.URL == schemeHttpsFull+website.WebSiteName || m.URL == schemeHttpFull+website.WebSiteName || m.URL == website.WebSiteName {
			return true, m
		}
	}
	return false, uptimerobot.Monitor{}
}

// createNewMonitor is a method to create new monitor for provided website in the provided account.
// Receives website object, returns created monitor id and email of account in which monitor created.
func (account Uptimerobot) createNewMonitor(website Website) (id int64, email string) {
	var (
		monitorType        = uptimerobot.TypeHTTP
		monitorKeywordType int
		m                  uptimerobot.Monitor
	)
	if website.Config.Scheme != schemeHttp {
		website.Config.Scheme = schemeHttps
	}
	if website.Config.Port == 0 && website.Config.Scheme == schemeHttps {
		website.Config.Port = 443
	} else {
		website.Config.Port = 80
	}

	if !IsEmptyString(website.Config.Keyword) {
		monitorType = uptimerobot.TypeKeyword

		if !(IsEmptyString(website.Config.KeywordType)) && !(strings.Contains(website.Config.KeywordType, "not")) {
			monitorKeywordType = uptimerobot.KeywordExists
		} else {
			monitorKeywordType = uptimerobot.KeywordNotExists
		}
		m = uptimerobot.Monitor{
			FriendlyName:  website.WebSiteName,
			URL:           website.Config.Scheme + "://" + website.WebSiteName,
			Port:          website.Config.Port,
			Type:          monitorType,
			KeywordType:   monitorKeywordType,
			KeywordValue:  website.Config.Keyword,
			AlertContacts: website.getAlertContactsFromSitelist(account),
		}
	} else {
		m = uptimerobot.Monitor{
			FriendlyName:  website.WebSiteName,
			URL:           website.Config.Scheme + "://" + website.WebSiteName,
			Port:          website.Config.Port,
			Type:          monitorType,
			AlertContacts: website.getAlertContactsFromSitelist(account),
		}
	}

	id, err := account.Client.CreateMonitor(m)
	if err != nil {
		log.Printf("Error while creating monitor for website %s in account %s. %e", website.WebSiteName, account.Email, err)
	} else {
		log.Printf("Monitor for %s created in account %s", website.WebSiteName, account.Email)
	}
	return id, account.Email
}

// getAlertContacts is an account method that returns all existing alert contacts in this account.
func (account Uptimerobot) getAlertContacts() (contacts []uptimerobot.AlertContact) {
	contacts, err := account.Client.AllAlertContacts()
	if err != nil {
		log.Printf("Failed to get alert contacts for account %s. %e", account.Email, err)
	}
	return contacts
}

// getAlertContactsFromSitelist is a website method that receives account in which monitor for this website will be created
// and returns an array with alert contact IDs. This function returns only alert contacts that exists in provided account.
// Alert contacts that not exists in provided account will not be returned.
func (website Website) getAlertContactsFromSitelist(account Uptimerobot) (contact []string) {
	allContacts := account.getAlertContacts()
	for _, wc := range website.Config.Contact {
		contactFound := false
		for _, c := range allContacts {
			if wc == c.FriendlyName {
				contact = append(contact, c.ID)
				contactFound = true
			}
		}
		if !contactFound {
			log.Printf("Failed to find alert contact %s for website %s in account %s.", wc, website.WebSiteName, account.Email)
		}
	}
	return contact
}

// isMonitorEqualToWebsite checks if existing monitor data in uptimerobot account equals to data in provided website.
// Method receives monitor and account and returns bool value. We need this check, because we can't change monitor
// in some cases and we need to delete it and create new one. For example, when monitor type changes from http to keyword.
func (website Website) isMonitorEqualToWebsite(m uptimerobot.Monitor, account Uptimerobot) bool {
	var (
		monitorAlertContactsInt, websiteAlertContactsInt []int
		monitorKeywordType int
	)
	monitorAlertContacts := getWebsiteAlertContactsFromAccount(account.Token, m.FriendlyName)
	websiteAlertContacts := website.getAlertContactsFromSitelist(account)

	for _, mac := range monitorAlertContacts {
		c, _ := strconv.Atoi(mac)
		monitorAlertContactsInt = append(monitorAlertContactsInt, c)
	}
	for _, wac := range websiteAlertContacts {
		c, _ := strconv.Atoi(wac)
		websiteAlertContactsInt = append(websiteAlertContactsInt, c)
	}
	sort.Ints(monitorAlertContactsInt)
	sort.Ints(websiteAlertContactsInt)

	if !(IsEmptyString(website.Config.Keyword)) {
		if !(IsEmptyString(website.Config.KeywordType)) && !(strings.Contains(website.Config.KeywordType, "not")) {
			monitorKeywordType = uptimerobot.KeywordExists
		} else {
			monitorKeywordType = uptimerobot.KeywordNotExists
		}
	}

	if website.Config.Scheme != schemeHttp {
		website.Config.Scheme = schemeHttps
	}

	u, _ := url.Parse(m.URL)
	if !(reflect.DeepEqual(monitorAlertContactsInt, websiteAlertContactsInt)) || (monitorKeywordType != m.KeywordType) || (u.Scheme != website.Config.Scheme) || (website.Config.Keyword != m.KeywordValue) {
		return false
	} else {
		return true
	}
}
