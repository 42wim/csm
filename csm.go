package main

import (
	"flag"
	"fmt"
	"github.com/42wim/go-syslog"
	"github.com/42wim/matterbridge/matterhook"
	"net"
	"strings"
	"time"
)

type config struct {
	Username, Channel, Listen, OurUser string
	MatterURL                          string `yaml:"matter_url"`
	Debug, DebugAll                    bool
}

var cfg = config{
	Channel:   "town-square",
	MatterURL: "",
	Username:  "bigbrother",
	Listen:    ":514",
	OurUser:   "root",
	Debug:     false,
	DebugAll:  false,
}

func init() {
	flag.StringVar(&cfg.Channel, "c", cfg.Channel, "Post input values to specified channel or user.")
	flag.StringVar(&cfg.MatterURL, "m", cfg.MatterURL, "Mattermost incoming webhooks URL.")
	flag.StringVar(&cfg.Username, "u", cfg.Username, "This username is used for posting.")
	flag.StringVar(&cfg.Listen, "l", cfg.Listen, "ip:port to listen on")
	flag.StringVar(&cfg.OurUser, "o", cfg.OurUser, "our user that we trust")
	flag.BoolVar(&cfg.Debug, "d", cfg.Debug, "debug messages send to mattermost")
	flag.BoolVar(&cfg.DebugAll, "dd", cfg.Debug, "more debug, print all received syslog messages")
	flag.Parse()
}

func main() {
	if cfg.MatterURL == "" {
		fmt.Println("Please specify your mattermost webhook url (-m option)")
		return
	}
	fmt.Println("listening on", cfg.Listen, "using", cfg.Username, "and sending to channel", cfg.Channel, "on", cfg.MatterURL)
	m := matterhook.New(cfg.MatterURL, matterhook.Config{DisableServer: true})
	msg := matterhook.OMessage{UserName: cfg.Username, Channel: cfg.Channel}

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)
	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164cisco)
	server.SetHandler(handler)
	server.ListenUDP(cfg.Listen)
	server.Boot()

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			if cfg.DebugAll {
				fmt.Printf("%#v\n", logParts)
			}
			if strings.Contains(logParts["tag"].(string), "PARSER") {
				var hostname string
				ip, _, _ := net.SplitHostPort(logParts["client"].(string))
				names, err := net.LookupAddr(ip)
				if err != nil {
					hostname = ip
				} else {
					hostname = names[0]
					hostnames := strings.Split(hostname, ".")
					hostname = hostnames[0]
				}

				res := strings.Split(logParts["content"].(string), "logged command:")

				var user, content string
				if len(res) == 2 {
					user = strings.Replace(res[0], "User:", "", -1)
					content = res[1]
				}
				first := ""
				if !strings.Contains(content, "exec enable") && !strings.Contains(content, "!exec") {
					if !strings.Contains(user, cfg.OurUser) {
						first = ":boom: "
					}
					now := time.Now().Format(time.Kitchen)
					msg.Text = first + "**" + strings.TrimSpace(content) + "** " + "*" + strings.TrimSpace(user) + "*@**" + hostname + "**" + " (" + ip + ") - _" + now + "_"
					if cfg.Debug || cfg.DebugAll {
						fmt.Println(msg.Text)
					}
					m.Send(msg)
				}
			}
		}
	}(channel)

	server.Wait()
}
