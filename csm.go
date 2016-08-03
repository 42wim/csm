package main

import (
	"flag"
	"fmt"
	"github.com/42wim/go-syslog"
	"github.com/42wim/matterbridge/matterhook"
	"net"
	"strings"
	"sync"
	"time"
)

type config struct {
	Username, Channel, Listen, OurUser string
	MatterURL                          string `yaml:"matter_url"`
	Buffer                             int
	Debug, DebugAll                    bool
}

type cache struct {
	sync.RWMutex
	msgs map[string][]*message
	last map[string]time.Time
}

type message struct {
	text string
	ts   time.Time
}

var cfg = config{
	Channel:   "town-square",
	MatterURL: "",
	Username:  "bigbrother",
	Listen:    ":514",
	OurUser:   "root",
	Buffer:    30,
	Debug:     false,
	DebugAll:  false,
}

var msgCache = cache{msgs: make(map[string][]*message), last: make(map[string]time.Time)}

func init() {
	flag.StringVar(&cfg.Channel, "c", cfg.Channel, "Post input values to specified channel or user.")
	flag.StringVar(&cfg.MatterURL, "m", cfg.MatterURL, "Mattermost incoming webhooks URL.")
	flag.StringVar(&cfg.Username, "u", cfg.Username, "This username is used for posting.")
	flag.StringVar(&cfg.Listen, "l", cfg.Listen, "ip:port to listen on")
	flag.StringVar(&cfg.OurUser, "o", cfg.OurUser, "our user that we trust")
	flag.IntVar(&cfg.Buffer, "b", cfg.Buffer, "seconds to buffer messages per switch for")
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
					msgCache.Lock()
					_, ok := msgCache.last[hostname]
					if !ok {
						msg.Text = first + "*" + strings.TrimSpace(user) + "*@**" + hostname + "**" + " (" + ip + ") - _" + now + "_"
						msg.Text += "\n" + "**" + strings.TrimSpace(content) + "** "
					} else {
						msg.Text = "**" + strings.TrimSpace(content) + "** "
					}
					msgCache.msgs[hostname] = append(msgCache.msgs[hostname], &message{msg.Text, time.Now()})
					msgCache.last[hostname] = time.Now()
					msgCache.Unlock()
				}
			}
		}
	}(channel)

	go func() {
		for {
			omsg := matterhook.OMessage{UserName: cfg.Username, Channel: cfg.Channel}
			msgCache.Lock()
			for hostname, v := range msgCache.last {
				if time.Since(v) > time.Second*time.Duration(cfg.Buffer) {
					msgs := msgCache.msgs[hostname]
					for _, mymsg := range msgs {
						omsg.Text = omsg.Text + "\n" + mymsg.text
					}
					delete(msgCache.msgs, hostname)
					delete(msgCache.last, hostname)
				}
				m.Send(omsg)
			}
			msgCache.Unlock()
			time.Sleep(time.Second)
		}
	}()

	server.Wait()
}
