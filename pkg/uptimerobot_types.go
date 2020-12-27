package uptimerobot_tool

import uptimerobot "github.com/bitfield/uptimerobot/pkg"

// Sitelist type is a representation of a sitelist.json file (array of websites)
type Sitelist []Website

// Website represents a website from sitelist and its config
type Website struct {
	WebSiteName string        `json:"web-site-name"`
	Config      WebsiteConfig `json:"config"`
}

// WebsiteConfig is a configuration data for website. It need to create monitor with proper configuration.
type WebsiteConfig struct {
	Keyword     string   `json:"keyword"`
	KeywordType string   `json:"keyword_type"`
	Contact     []string `json:"contact"`
	Scheme      string   `json:"scheme"`
	Port        int      `json:"port"`
}

// Uptimerobot represents main account.
type Uptimerobot struct {
	Token  string `yaml:"token" json:"token" toml:"token"`
	Email  string `yaml:"email" json:"email" toml:"email"`
	Client uptimerobot.Client
}
