// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/smtp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type conf struct {
	SMTPHost  string   `yaml:"smtp_host"`
	SMTPPort  string   `yaml:"smtp_port"`
	Avatar    string   `yaml:"avatar"`
	EmailAddr string   `yaml:"email_addr"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Receivers []string `yaml:"receivers"`
	SendTime  string   `yaml:"send_time"`
	Title     string   `yaml:"title"`
	Content   string   `yaml:"content"`
}

func (c *conf) sendInbox(title, body string) error {
	// Text in an encoded-word in a display-name must not contain certain
	// characters like quotes or parentheses (see RFC 2047 section 5.3).
	// When this is the case encode the title using base64 encoding.
	if strings.ContainsAny(title, "\"#$%&'(),.:;<>@[]^`{|}~") {
		title = mime.BEncoding.Encode("utf-8", title)
	} else {
		title = mime.QEncoding.Encode("utf-8", title)
	}

	err := smtp.SendMail(
		c.SMTPHost+":"+c.SMTPPort,
		smtp.PlainAuth("", c.Username, c.Password, c.SMTPHost),
		c.EmailAddr, c.Receivers,
		[]byte(fmt.Sprintf("Subject: %s\r\nFrom: %s <%s>\r\nTo: %s\r\n%s",
			// Content-Type: text/plain; charset=utf-8; format=flowed
			// Content-Transfer-Encoding: 7bit
			// Content-Language: en-US
			title,
			c.Avatar, c.EmailAddr, strings.Join(c.Receivers, ","),
			body,
		)))
	if err != nil {
		return err
	}
	return nil
}

var c conf

const domain = "@ifi.lmu.de"

func init() {
	d, err := ioutil.ReadFile("conf.yml")
	if err != nil {
		panic("cannot find the file")
	}

	err = yaml.Unmarshal(d, &c)
	if err != nil {
		panic(fmt.Errorf("cannot parse yaml file: %v", err))
	}

	c.EmailAddr += domain
	for i := range c.Receivers {
		c.Receivers[i] += domain
	}
}

func main() {
	t, err := time.Parse(time.RFC3339, c.SendTime)
	if err != nil {
		panic(err)
	}
	now := time.Now().In(t.Location())
	sendTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		t.Hour(),
		// now.Hour(),
		t.Minute(),
		0, 0, t.Location())
	if now.Sub(sendTime) > 0 {
		sendTime = sendTime.Add(time.Hour * 24)
		// sendTime = sendTime.Add(time.Hour)
	}

	for {
		fmt.Printf("next email will be send at: %v\n", sendTime)
		time.Sleep(sendTime.Sub(time.Now().In(t.Location())))
		err := c.sendInbox(
			fmt.Sprintf(
				c.Title,
				time.Now().Day(),
				time.Now().Month(),
				time.Now().Year()),
			fmt.Sprintf(
				c.Content,
				time.Now().Day(),
				time.Now().Month(),
				time.Now().Year()),
		)
		if err != nil {
			fmt.Printf("cannot send email: %v\n", err)
		}
		fmt.Printf("email was sent at: %v\n", sendTime)
		sendTime = sendTime.Add(time.Hour)
	}
}
