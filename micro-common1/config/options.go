package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/suiyunonghen/dxsvalue"
	"strings"
	"time"
)

type CustomOptions struct {
	Custom *dxsvalue.DxValue
}

func (opt *CustomOptions) setCustom(key string, value *dxsvalue.DxValue) {
	if opt.Custom == nil {
		opt.Custom = dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	}
	cache := opt.Custom.ValueCache()
	opt.Custom.SetKeyCached(key, value.DataType, cache).CopyFrom(value, cache)
}

func (opt *CustomOptions) clear() {
	if opt.Custom != nil {
		opt.Custom.Clear()
	}
}

type Options struct {
	Project       *Project
	Log           *LogOptions
	Jwt           *JwtOptions
	DFS           *DFSOptions
	Redis         *RedisOptions
	DB            *DBOptions
	Mqtt          *MqttOptions
	RabbitMq      *RabbitMqOptions
	Login         *LoginOptions
	Pprof         *PprofOptions
	Mail          *MailOptions
	File          *FileOptions
	Wechat        *WechatOptions
	Voip          *VoipOptions
	Miniapp       *MiniappOptions
	Elastic       *ElasticOptions
	Admin         *AdminOptions
	RelatedServer *RelatedServer
	Community     *CommunityOptions
	CustomOptions // 自定义的配置参数
}

func (opt *Options) LoadFromValue(rootValue *dxsvalue.DxValue) {
	opt.clear()
	rootValue.Visit(func(Key string, value *dxsvalue.DxValue) bool {
		switch strings.ToLower(Key) {
		case "project":
			if Data.Project == nil {
				Data.Project = &Project{}
			}
			Data.Project.LoadFromValue(value)
		case "dfs":
			if Data.DFS == nil {
				Data.DFS = &DFSOptions{}
			}
			Data.DFS.LoadFromValue(value)
		case "log":
			if Data.Log == nil {
				Data.Log = DefaultLogOptions()
			}
			Data.Log.LoadFromValue(value)
		case "jwt":
			if Data.Jwt == nil {
				Data.Jwt = &JwtOptions{}
			}
			Data.Jwt.LoadFromValue(value)
		case "redis":
			if Data.Redis == nil {
				Data.Redis = &RedisOptions{}
			}
			Data.Redis.LoadFromValue(value)
		case "db":
			if Data.DB == nil {
				Data.DB = &DBOptions{}
			}
			Data.DB.LoadFromValue(value)
		case "mqtt":
			if Data.Mqtt == nil {
				Data.Mqtt = &MqttOptions{}
			}
			value.ToStdValue(Data.Mqtt, true)
		case "rabbitmq":
			if Data.RabbitMq == nil {
				Data.RabbitMq = &RabbitMqOptions{}
			}
			value.ToStdValue(Data.RabbitMq, true)
		case "login":
			if Data.Login == nil {
				Data.Login = &LoginOptions{}
			}
			value.ToStdValue(Data.Login, true)
		case "pprof":
			if Data.Pprof == nil {
				Data.Pprof = &PprofOptions{}
			}
			value.ToStdValue(Data.Pprof, true)
		case "mail":
			if Data.Mail == nil {
				Data.Mail = &MailOptions{}
			}
			value.ToStdValue(Data.Mail, true)
		case "file":
			if Data.File == nil {
				Data.File = &FileOptions{}
			}
			value.ToStdValue(Data.File, true)
		case "wechat":
			if Data.Wechat == nil {
				Data.Wechat = &WechatOptions{}
			}
			value.ToStdValue(Data.Wechat, true)
		case "voip":
			if Data.Voip == nil {
				Data.Voip = &VoipOptions{}
			}
			value.ToStdValue(Data.Voip, true)
		case "miniapp":
			if Data.Miniapp == nil {
				Data.Miniapp = &MiniappOptions{}
			}
			value.ToStdValue(Data.Miniapp, true)
		case "elastic":
			if Data.Elastic == nil {
				Data.Elastic = &ElasticOptions{}
			}
			value.ToStdValue(Data.Elastic, true)
		case "admin":
			if Data.Admin == nil {
				Data.Admin = &AdminOptions{}
			}
			value.ToStdValue(Data.Admin, true)
		case "relatedserver":
			if Data.RelatedServer == nil {
				Data.RelatedServer = &RelatedServer{}
			}
			value.ToStdValue(Data.RelatedServer, true)
		case "community":
			if Data.Community == nil {
				Data.Community = &CommunityOptions{}
			}
			value.ToStdValue(Data.Community, true)
		default:
			opt.setCustom(Key, value)
		}
		return true
	})
}

type DFSOptions struct {
	Addr     string
	Username string
	Password string
	Bucket   string
	Domain   string
	SSL      bool
	CustomOptions
}

func (m *DFSOptions) LoadFromValue(rootValue *dxsvalue.DxValue) {
	m.clear()
	rootValue.Visit(func(Key string, value *dxsvalue.DxValue) bool {
		switch strings.ToLower(Key) {
		case "addr":
			m.Addr = value.String()
		case "username":
			m.Username = value.String()
		case "password":
			m.Password = value.String()
		case "bucket":
			m.Bucket = value.String()
		case "domain":
			m.Domain = value.String()
		case "ssl":
			m.SSL = value.Bool()
		default:
			m.setCustom(Key, value)
		}
		return true
	})
}

type LogOptions struct {
	ShowCaller  bool   // 是否展示调用者（打印位置）
	ShowConsole bool   // 是否输出到控制台
	ColorLog    bool   // 是否彩色输出
	JsonFormat  bool   // 是否输出json格式日志
	LazyWrite   bool   // 是否启用异步输出
	SplitTime   uint8  // 日志的分割时长，小时
	SplitSize   int    // 日志分割大小 LogSplit为0时启用
	Level       string // 日志级别
	File        string // 日志保存路径
	Project     string // 工程名
	Author      string // 作者
	Machine     string // 日志机器信息
	CustomOptions
}

func (logoptions *LogOptions) LoadFromValue(value *dxsvalue.DxValue) {
	logoptions.clear()
	value.Visit(func(Key string, value *dxsvalue.DxValue) bool {
		switch strings.ToLower(Key) {
		case "project":
			logoptions.Project = value.String()
		case "showcaller":
			logoptions.ShowCaller = value.Bool()
		case "showconsole":
			logoptions.ShowConsole = value.Bool()
		case "colorlog":
			logoptions.ColorLog = value.Bool()
		case "jsonformat":
			logoptions.JsonFormat = value.Bool()
		case "lazywrite":
			logoptions.LazyWrite = value.Bool()
		case "splittime":
			logoptions.SplitTime = uint8(value.Int())
		case "splitsize":
			logoptions.SplitSize = int(value.Int())
		case "level":
			logoptions.Level = value.String()
		case "file":
			logoptions.File = value.String()
		case "author":
			logoptions.Author = value.String()
		case "machine":
			if value.DataType == dxsvalue.VT_Object {
				logoptions.Machine = ""
			} else {
				logoptions.Machine = value.String()
			}
		default:
			logoptions.setCustom(Key, value)
		}
		return true
	})
}

func DefaultLogOptions() *LogOptions {
	return &LogOptions{
		ShowCaller:  true,
		ShowConsole: true,
		ColorLog:    true,
		JsonFormat:  false,
		LazyWrite:   false,
		SplitTime:   24,
		SplitSize:   0,
		Level:       "info",
		File:        "",
		Project:     "",
		Author:      "",
		Machine:     "",
	}
}

type Project struct {
	Name        string
	Port        uint16 // http端口
	TCP         uint16 // tcp端口
	WorkerId    uint16 // 机器的工作ID，主要用来生成雪花ID的，分布式部署，不要重复
	LogResponse bool   // 是否打印http的response
	Swagger     bool   // 是否开启swagger
	CustomOptions
}

func (proj *Project) LoadFromValue(rootValue *dxsvalue.DxValue) {
	proj.clear()
	rootValue.Visit(func(Key string, value *dxsvalue.DxValue) bool {
		switch strings.ToLower(Key) {
		case "name":
			proj.Name = value.String()
		case "port":
			proj.Port = uint16(value.Int())
		case "tcp":
			proj.TCP = uint16(value.Int())
		case "workerid":
			proj.WorkerId = uint16(value.Int())
		case "logresponse":
			proj.LogResponse = value.Bool()
		case "swagger":
			proj.Swagger = value.Bool()
		default:
			proj.setCustom(Key, value)
		}
		return true
	})
}

type JwtOptions struct {
	Issuer         string
	Secret         string
	Expires        time.Duration
	RefreshExpires time.Duration
	CustomOptions
}

func (jwt *JwtOptions) LoadFromValue(rootvalue *dxsvalue.DxValue) {
	jwt.clear()
	rootvalue.Visit(func(Key string, value *dxsvalue.DxValue) bool {
		switch strings.ToLower(Key) {
		case "issuer":
			jwt.Issuer = value.String()
		case "secret":
			jwt.Secret = value.String()
		case "expires":
			duration, _ := time.ParseDuration(value.String())
			jwt.Expires = duration
		case "refreshexpires":
			duration, _ := time.ParseDuration(value.String())
			jwt.RefreshExpires = duration
		default:
			jwt.setCustom(Key, value)
		}
		return true
	})
}

type RedisOptions struct {
	Single   *redis.Options
	Sentinel *redis.FailoverOptions
	Cluster  *redis.ClusterOptions
	CustomOptions
}

func (p *RedisOptions) LoadFromValue(root *dxsvalue.DxValue) {
	p.clear()
	root.Visit(func(Key string, value *dxsvalue.DxValue) bool {
		switch strings.ToLower(Key) {
		case "single":
			if p.Single == nil {
				p.Single = &redis.Options{}
			}
			value.ToStdValue(p.Single, true)
		case "sentinel":
			if p.Sentinel == nil {
				p.Sentinel = &redis.FailoverOptions{}
			}
			value.ToStdValue(p.Sentinel, true)
		case "cluster":
			if p.Cluster == nil {
				p.Cluster = &redis.ClusterOptions{}
			}
			value.ToStdValue(p.Cluster, true)
		default:
			p.setCustom(Key, value)
		}
		return true
	})
}

type CacheStyle uint8

const (
	Cache_None CacheStyle = iota
	Cache_Mem
	Cache_Redis
)

type DBOptions struct {
	CacheStyle CacheStyle `yaml:"CacheStyle"`
	UnUseOrm   bool       // 是否启用orm
	Master     *DBOptionsEntry
	Slave      *DBOptionsEntry
	CustomOptions
}

func (dbopt *DBOptions) LoadFromValue(root *dxsvalue.DxValue) {
	dbopt.clear()
	root.Visit(func(Key string, value *dxsvalue.DxValue) bool {
		switch strings.ToLower(Key) {
		case "cachestyle":
			dbopt.CacheStyle = CacheStyle(value.Int())
		case "unuseorm":
			dbopt.UnUseOrm = value.Bool()
		case "master":
			if dbopt.Master == nil {
				dbopt.Master = &DBOptionsEntry{}
			}
			value.ToStdValue(dbopt.Master, true)
		case "slave":
			if dbopt.Slave == nil {
				dbopt.Slave = &DBOptionsEntry{}
			}
			value.ToStdValue(dbopt.Slave, true)
		default:
			dbopt.setCustom(Key, value)
		}
		return true
	})
}

type RelatedServer struct {
	BizCore  string
	Dispatch string
	Notice   string
}

type DBOptionsEntry struct {
	Dialector string
	Username  string
	Password  string
	Host      string
	Path      string
	RawQuery  string
	MaxOpen   int
	MaxIdle   int
}

type MqttOptions struct {
	Services []string
	Username string
	Password string
	ClientId string
}

type RabbitMqOptions struct {
	Service  string
	Host     string
	Username string
	Password string
}

type LoginOptions struct {
	User     string
	Password string
}

type PprofOptions struct {
	Ip   string
	Port int
}

type MailOptions struct {
	Username  string
	Password  string
	Host      string
	From      string
	Defaultto string
}

type FileOptions struct {
	MapDirectory       string
	RobotIconPath      string
	VoicePath          string
	RepeatMsgCacheTime int64
	ExpireCacheTime    int
}

type WechatOptions struct {
	Appid             string
	Secret            string
	Token             string
	AesKey            string
	ArriveTemplateId  string
	RecieveTemplateId string
	MiniappId         string
}

type VoipOptions struct {
	AccessKeyId     string
	AccessKeySecret string
	RegionId        string
	WarnTtsCode     string
	ArriveTtsCode   string
}

type MiniappOptions struct {
	Appid  string
	Secret string
}

type ElasticOptions struct {
	Hosts    []string
	Username string
	Password string
}

type AdminOptions struct {
	Frontwebbaseurl string
	TokenExpire     int64
}

type CommunityOptions struct {
	ImgPath string
	FileUrl string
}
