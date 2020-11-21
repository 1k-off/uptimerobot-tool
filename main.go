package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)


type Sitelist []struct {
	WebSiteName string `json:"web-site-name"`
	Config SitelistConfig `json:"config"`
}

type SitelistConfig      struct {
	Keyword string `json:"keyword"`
	Contact string `json:"contact"`
}

var (
	LogInfo, LogError            *log.Logger
	env                          = NewEnv()
	httpClient, httpClientConfig = NewHTTPClient()
)

var ApplicationEnvironment string

func init() {
	LogInfo = log.New(os.Stdout, "INFO: ", 0)
	LogError = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
}
func main() {
	var (
		sitelist Sitelist
	)

	sitelist = getSitelistFromFile("./sitelist.json")
	UptimerobotWorkflow(env, sitelist)
}

//getSitelistFromFile retrieves list with sites to check from local file
// returns all sites as array of strings
func getSitelistFromFile(path string) Sitelist {
	var sitelist Sitelist
	file, err := ioutil.ReadFile(path)
	if err != nil {
		LogError.Fatalln("Can't open sitelist file.", err)
	}
	_ = json.Unmarshal([]byte(file), &sitelist)
	return sitelist
}