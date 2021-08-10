package email

import (
	"crypto/tls"
	"net/smtp"
	"testing"
)

func TestEmail(t *testing.T) {
	e := NewEmail()
	e.From = "it_system@example.cn"
	e.To = []string{"test@example.com"}
	e.Bcc = []string{"skirrund@example.com"}
	e.Cc = []string{"skirrund@example.com"}
	e.Subject = "Test test"
	//e.Text = []byte("Text Body is, of course, supported!\n")
	e.HTML = []byte("<h1>Fancy Html is supported, too!</h1>\n")
	err := e.SendWithStartTLS("smtp.exmail.qq.com:465", smtp.PlainAuth("", e.From, "", "smtp.exmail.qq.com"), &tls.Config{InsecureSkipVerify: true, ServerName: "smtp.exmail.qq.com"})
	if err != nil {
		t.Error(err)
	}
}
