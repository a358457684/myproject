package mail

import (
	"common/config"
	"errors"
)

var (
	username = ""
	password = ""
	host     = ""
)

func initmail(mail *config.MailOptions) error {
	if mail == nil {
		return errors.New("邮箱配置为空")
	}

	username = config.Data.Mail.Username
	password = config.Data.Mail.Password
	host = config.Data.Mail.Host
	return nil
}
