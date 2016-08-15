# Cisco Syslog to Mattermost (csm)

Send in real-time your cisco config changes to mattermost
Requires mattermost 1.2.0+

## binaries
You can find the binaries [here](https://github.com/42wim/csm/releases/)

## building
Go 1.6+ is required. Make sure you have [Go](https://golang.org/doc/install) properly installed, including setting up your [GOPATH] (https://golang.org/doc/code.html#GOPATH)

```
cd $GOPATH
go get github.com/42wim/csm
```

You should now have csm binary in the bin directory:

```
$ ls bin/
csm
```

## running
```
Usage of ./csm:
  -b int
        seconds to buffer messages per switch for (default 30)
  -c string
        Post input values to specified channel or user. (default "town-square")
  -d    debug messages send to mattermost
  -dd
        more debug, print all received syslog messages
  -l string
        ip:port to listen on (default ":514")
  -m string
        Mattermost incoming webhooks URL.
  -o string
        our user that we trust (default "root")
  -u string
        This username is used for posting. (default "bigbrother")
```

* -m is your incoming webhook url (account settings - integrations - incoming webhooks)  e.g -m "http://mattermost.yourdomain.com/hooks/incomingwebhookkey"  
* -l to change your syslog listen address, by default on port 514.

### cisco
You'll have to configure your IOS to 
* send syslog messages to the ip:port you're running csm on
* send the commands to syslog

```
conf t
logging host 1.2.3.4 transport udp port 5555
archive
 log config
  logging enable
  notify syslog contenttype plaintext
  hidekeys
```

Now you can run csm:
```
csm -l "1.2.3.4:5555" -m http://mattermost.yourdomain.com/hooks/incomingwebhookkey
```

### mattermost
You'll have to configure the incoming and outgoing webhooks. 

* incoming webhooks
Go to "account settings" - integrations - "incoming webhooks".  
Choose a channel at "Add a new incoming webhook", this will create a webhook URL right below.  
This URL should be set in the matterbridge.conf in the [mattermost] section (see above)  

### screenshot
![screen](https://i.snag.gy/e86Vhb.jpg)
