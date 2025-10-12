package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Message struct {
	From    string
	To      string
	Subject string
	Body    string
}

func SendEmail(config *SMTPConfig, msg *Message) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	emailMsg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s\r\n", msg.From, msg.To, msg.Subject, msg.Body)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("Failed to connect to SMTP server: %v", err)
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		log.Printf("Failed to create SMTP client: %v", err)
		return err
	}
	defer client.Close()

	if err = client.Hello("localhost"); err != nil {
		log.Printf("Failed to send EHLO: %v", err)
		return err
	}

	tlsConfig := &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: false,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("Failed to start TLS: %v", err)
		return err
	}

	if err = client.Auth(auth); err != nil {
		log.Printf("Failed to authenticate: %v", err)
		return err
	}

	if err = client.Mail(msg.From); err != nil {
		log.Printf("Failed to set sender: %v", err)
		return err
	}

	if err = client.Rcpt(msg.To); err != nil {
		log.Printf("Failed to set recipient: %v", err)
		return err
	}

	w, err := client.Data()
	if err != nil {
		log.Printf("Failed to create data writer: %v", err)
		return err
	}

	_, err = w.Write([]byte(emailMsg))
	if err != nil {
		log.Printf("Failed to write email body: %v", err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Printf("Failed to close data writer: %v", err)
		return err
	}

	log.Printf("Email sent successfully to %s", msg.To)
	return client.Quit()
}
