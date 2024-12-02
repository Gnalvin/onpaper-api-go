package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify" // viper 库监听文件变化时用到

	"github.com/spf13/viper" // 可以加载配置文件
)

// Conf 全局变量，用来保存程序的所有配置信息
var Conf = new(Config) // 创建指针类型结构体

// Config 对应主目录下 config.yaml
type Config struct {
	Name                string
	Mode                string
	Host                string
	Port                int
	Version             string
	TokenPublicKeyPath  string `mapstructure:"PUBLICKEY_PATH"`
	TokenPrivateKeyPath string `mapstructure:"PRIVATEKEY_PATH"`

	*MySQLConfig    `mapstructure:"Mysql"`
	*RedisConfig    `mapstructure:"Redis"`
	*MongodbConfig  `mapstructure:"MongoDB"`
	*LogConfig      `mapstructure:"Log"`
	*OssConfig      `mapstructure:"Oss"`
	*SMS            `mapstructure:"SMS"`
	*SnowFlake      `mapstructure:"SnowFlake"`
	*InvitationCode `mapstructure:"InvitationCode"`
	*MiniProgram    `mapstructure:"MiniProgram"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"HOST"`
	Port         int    `mapstructure:"PORT"`
	DbName       string `mapstructure:"DATABASE"`
	User         string `mapstructure:"USER"`
	Password     string `mapstructure:"PASSWORD"`
	MaxOpenConns int    `mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleConns int    `mapstructure:"MAX_IDLE_CONNS"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"ADDR"`
	Password string `mapstructure:"PASSWORD"`
	UserName string `mapstructure:"USERNAME"`
	PoolSize int    `mapstructure:"POOL_SIZE"`
}

type MongodbConfig struct {
	Host     string `mapstructure:"HOST"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
	DbName   string `mapstructure:"DATABASE"`
	PoolSize uint64 `mapstructure:"POOL_SIZE"`
}

type LogConfig struct {
	Level      string `mapstructure:"LEVEL"`
	Filename   string `mapstructure:"FILENAME"`
	MaxSize    int    `mapstructure:"MAX_SIZE"`
	MaxAge     int    `mapstructure:"MAX_AGE"`
	MaxBackups int    `mapstructure:"MAX_BACKUPS"`
}

type OssConfig struct {
	OssMaxSecretId  string `mapstructure:"OSS_MAX_SecretId"`
	OssMaxSecretKey string `mapstructure:"OSS_MAX_SecretKey"`
	OssStsSecretId  string `mapstructure:"OSS_STS_SecretId"`
	OssStsSecretKey string `mapstructure:"OSS_STS_SecretKey"`
	StsRoleArn      string `mapstructure:"STS_RoleArn"`
	Endpoint        string `mapstructure:"Endpoint"`
	PreviewBucket   string `mapstructure:"PreviewBucket"`
	OriginalBucket  string `mapstructure:"OriginalBucket"`
	TempBucket      string `mapstructure:"TempBucket"`
	AppId           string `mapstructure:"AppId"`
}

type SMS struct {
	SMSSecretId     string `mapstructure:"SecretId"`
	SMSSecretKey    string `mapstructure:"SecretKey"`
	SMSTemplateCode string `mapstructure:"TemplateCode"`
}

type SnowFlake struct {
	SnowStartTime string `mapstructure:"Start_Time"`
	MachineId     int64  `mapstructure:"Machine_Id"`
}

type InvitationCode struct {
	MagicCode string `mapstructure:"MagicCode"`
}

type MiniProgram struct {
	MiniAppID     string `mapstructure:"AppID"`
	MiniAppSecret string `mapstructure:"AppSecret"`
}

func ConfigInit() (err error) {
	//viper.SetConfigName("config") // 指定配置文件名称（不需要带后缀）
	//viper.AddConfigPath(".")   // 指定查找配置文件的路径（这里使用相对可执行文件.exe路径）

	// 从哪个文件中读取 配置信息
	viper.SetConfigFile("./config.yaml")

	err = viper.ReadInConfig() //读取配置信息
	if err != nil {
		// 读取配置信息失败
		fmt.Printf(" viper.ReadInConfig() 读取配置信息失败， err:%v\n", err)
		return
	}
	//把读取到的配置反序列化到 结构体中
	if err = viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed 反序列失败, err:%v\n", err)
	}
	viper.WatchConfig() // 监听配置文件修改
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了...")
		// 重新反序列化
		if err = viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed 反序列失败, err:%v\n", err)
		}
	})
	fmt.Printf("settings.Init() 配置初始化成功\n")
	return
}
