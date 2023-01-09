package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	log.SetLevel(log.DebugLevel)
	config := MustLoadYamlConfig("config.yml", "config.yaml", "~/.config/static-short-link/config.yml", "/usr/local/etc/static-short-link/config.yml", "/etc/static-short-link/config.yml")
	CreateCfPagesConfig(config, "site/_redirects")
	CreateVercelConfig(config, "site/vercel.json")
	t := template.New("ssl")
	t.Funcs(template.FuncMap{
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},
	})
	template.Must(t.ParseGlob("views/*.html"))

	pages := []string{
		"404",
		"index",
		"lists",
	}
	var output *os.File
	var err error
	for _, page := range pages {
		output, err = os.Create(fmt.Sprintf("site/%s.html", page))
		if err != nil {
			log.Errorf("fail to create page %q for %q", page, err)
			continue
		}
		err = t.ExecuteTemplate(output, fmt.Sprintf("%s.html", page), config)
		if err != nil {
			log.Errorf("fail to render page %q for %q", page, err)
			continue
		}
		log.Debugf("%q rendered ", page)
	}
}
func CreateVercelConfig(cfg *SiteConfig, filename string) {
	m := map[string]interface{}{"redirects": cfg.Redirects}
	b, _ := json.MarshalIndent(m, "", " ")
	ioutil.WriteFile(filename, b, 0x666)
}
func CreateCfPagesConfig(cfg *SiteConfig, filename string) {
	lines := []string{}
	for _, redirect := range cfg.Redirects {
		lines = append(lines, redirect.String())
	}
	ioutil.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0x666)
}

type Redirection struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Permanent   bool   `json:"permanent"`
	Name        string `json:"-"`
}

func (r Redirection) String() string {
	code := 301
	if r.Permanent {
		code = 302
	}
	return fmt.Sprintf("%s %s %d", r.Source, r.Destination, code)
}

type SiteConfig struct {
	Redirects []*Redirection `yaml:"redirects"`
	Site      struct {
		Name        string `yaml:"name"`
		PoweredBy   string `yaml:"powered_by"`
		Description string `yaml:"description"`
	} `yaml:"site"`
}

func MustLoadYamlConfig(filenames ...string) *SiteConfig {
	var config SiteConfig
	for _, filename := range filenames {
		var e error
		if fi, e := os.Stat(filename); os.IsNotExist(e) {
			continue
		} else {
			log.Infof("use config %s %s", filename, fi.ModTime())
		}
		body, e := ioutil.ReadFile(filename)
		if e != nil {
			panic(e)
		}
		e = yaml.Unmarshal(body, &config)
		if e != nil {
			panic(e)
		} else {
			break
		}
	}
	return &config
}
