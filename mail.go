package main

import (
	"net/smtp"
	"text/template"
	"bytes"
	"time"
	"fmt"
)

func SendMail(to string, subject string, content string) (err error) {
	tmpl, _ := template.New("Mail").Parse(`To: {{.To}}
Date: {{.Date}}
Subject: {{.Subject}}

{{.Message}}
	`)
	buffer := new(bytes.Buffer)

	tmpl.Execute(buffer, map[string]interface{}{
		"To": to,
		"Date": time.Now().String(),
		"Subject": subject,
		"Message": content,
	})

	auth := smtp.PlainAuth("", Config.Mail.Username, Config.Mail.Password, Config.Mail.Server)
	if err := smtp.SendMail(
		fmt.Sprintf("%s:%d", Config.Mail.Server, Config.Mail.Port),
		auth,
		Config.Mail.Username,
		[]string{to},
		buffer.Bytes(),
	); err != nil {
		return err
	}

	return nil
}
