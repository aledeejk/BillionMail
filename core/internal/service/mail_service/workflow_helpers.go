package mail_service

import (
	"context"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

func GetDefaultSender(ctx context.Context) (string, error) {
	var row struct {
		Username string `json:"username"`
	}
	err := g.DB().Model("mailbox").Ctx(ctx).
		Fields("username").
		Where("active", 1).
		Order("id asc").
		Limit(1).
		Scan(&row)
	if err != nil {
		return "", err
	}
	return row.Username, nil
}

func NewWorkflowEmailSender(ctx context.Context, senderEmail string) (*EmailSender, error) {
	if senderEmail == "" {
		var err error
		senderEmail, err = GetDefaultSender(ctx)
		if err != nil || senderEmail == "" {
			senderEmail = g.Cfg().MustGet(ctx, "workflow.smtp.from").String()
		}
	}

	if senderEmail != "" {
		es, err := NewEmailSenderWithLocal(senderEmail)
		if err == nil {
			return es, nil
		}
		g.Log().Debugf(ctx, "workflow send-email: local mailbox %q failed (%v), trying config SMTP", senderEmail, err)
	}

	return NewConfigEmailSender(ctx)
}

func NewConfigEmailSender(ctx context.Context) (*EmailSender, error) {
	host := g.Cfg().MustGet(ctx, "workflow.smtp.host").String()
	port := g.Cfg().MustGet(ctx, "workflow.smtp.port", "587").String()
	username := g.Cfg().MustGet(ctx, "workflow.smtp.username").String()
	password := g.Cfg().MustGet(ctx, "workflow.smtp.password").String()
	from := g.Cfg().MustGet(ctx, "workflow.smtp.from").String()

	g.Log().Debugf(ctx, "workflow SMTP config: host=%q port=%q username=%q from=%q", host, port, username, from)

	if host == "" {
		return nil, errors.New("workflow SMTP not configured: set workflow.smtp.host in config.yaml")
	}
	if from == "" {
		from = username
	}
	if from == "" {
		return nil, errors.New("workflow SMTP from address not configured")
	}

	sni := host
	if parts := strings.SplitN(host, ":", 2); len(parts) == 2 {
		sni = parts[0]
	}

	es := &EmailSender{
		Email:    from,
		UserName: username,
		Password: password,
		Host:     host,
		Port:     port,
		SNI:      sni,
	}
	return es, nil
}
