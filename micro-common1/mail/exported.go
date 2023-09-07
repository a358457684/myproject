package mail

import (
	"common/config"
	"common/log"
	"errors"
	"gopkg.in/gomail.v2"
	"strconv"
)

func Init() {
	var err error
	defer func() {
		if err != nil {
			log.WithError(err).Error("邮件通知初始化失败")
			panic(err)
		}
		log.Info("邮件通知初始化成功")
	}()
	if config.Data.Mail == nil {
		err = errors.New("读取邮件通知配置失败")
		return
	}
	err = initmail(config.Data.Mail)
}

func Sendmail(mailTo []string, subject string, body string, annexPath, annexName string) error {

	mailConn := map[string]string{
		"user": username,
		"pass": password,
		"host": host,
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(mailConn["user"], "notice")) //这种方式可以添加别名，即“XX官方”
	//m.SetHeader("From", mailConn["user"])
	m.SetHeader("To", mailTo...)    //发送给多个用户
	m.SetHeader("Subject", subject) //设置邮件主题
	m.SetBody("text/plain", body)   //设置邮件正文

	if annexPath != "" {
		if annexName != "" {
			m.Attach(annexPath, gomail.Rename(annexName)) //添加附件
		} else {
			m.Attach(annexPath) //添加附件
		}
	}

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	log.LogWithErrorf(err, "Sendmail fail. %v", m)
	return err

}
