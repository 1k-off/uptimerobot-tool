package uptimerobot_tool

import uptimerobot "github.com/bitfield/uptimerobot/pkg"

type Sitelist []Website

type Website struct {
	WebSiteName string        `json:"web-site-name"`
	Config      WebsiteConfig `json:"config"`
}
type WebsiteConfig struct {
	Keyword     string   `json:"keyword"`
	KeywordType string   `json:"keyword_type"`
	Contact     []string `json:"contact"`
	Scheme      string   `json:"scheme"`
	Port        int      `json:"port"`
}

type Uptimerobot struct {
	Token  string `yaml:"token" json:"token" toml:"token"`
	Email  string `yaml:"email" json:"email" toml:"email"`
	Client uptimerobot.Client
}
