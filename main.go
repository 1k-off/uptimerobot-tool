package main

import (
	ut "github.com/1k-off/uptimerobot-tool/pkg"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

// Environment is an application data structure
type Environment struct {
	Uptimerobot []ut.Uptimerobot `yaml:"uptimerobot"`
}

func main() {
	var e Environment
	yamlFile, err := ioutil.ReadFile("data/config.yml")
	if err != nil {
		log.Println("Error while reading yaml file: ", err)
	}
	err = yaml.Unmarshal(yamlFile, &e)
	if err != nil {
		log.Println("Error while decoding yaml file: ", err)
	}
	sitelist := ut.GetSitelistFromFile("data/sitelist.json")
	ut.ProcessMonitors(e.Uptimerobot, sitelist)
}
