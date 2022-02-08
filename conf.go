package tigerbalm

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"gopkg.in/yaml.v2"
)

var (
	Conf      *Config
	RotateLog *rotatelogs.RotateLogs

	h    bool
	file string

	defaultFile string = "./tigerbalm.yaml"
)

type Config struct {
	Web struct {
		Addr string `yaml:"addr"`
	} `yaml:"web"`

	Kafka struct {
		Enable   bool     `yaml:"enable"`
		Brokers  []string `yaml:"brokers"`
		Consumer struct { //dumped from sarama
			Group struct {
				Session struct {
					Timeout time.Duration `yaml:"timeout"`
				} `yaml:"session`
				Heartbeat struct {
					Interval time.Duration `yaml:"interval"`
				} `yaml:"heartbeat"`
			} `yaml:"group"`
			Offsets struct {
				Initial int64 `yaml:"initial"`
			} `yaml:"offsets"`
		} `yaml:"consumer"`
	} `yaml:"kafka"`

	Plugin struct {
		Path      string `yaml:"path"`
		WatchPath bool   `yaml:"watch_path"`
		Log       struct {
			Enable   bool   `yaml:"enable"`
			Path     string `yaml:"path"`
			Level    string `yaml:"level"`
			MaxSize  int64  `yaml:"maxsize"`
			MaxRolls uint   `yaml:"maxrolls"`
		} `yaml:"log"`
	} `yaml:"plugin"`

	Log struct {
		Level    string `yaml:"level"`
		File     string `yaml:"file"`
		MaxSize  int64  `yaml:"maxsize"`
		MaxRolls uint   `yaml:"maxrolls"`
	} `yaml:"log"`

	Env []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"env"`
}

func Init() error {
	time.LoadLocation("Asia/Shanghai")

	err := initCmd()
	if err != nil {
		return err
	}

	err = initConf()
	if err != nil {
		return err
	}

	err = initLog()
	return err
}

func initCmd() error {
	flag.StringVar(&file, "f", defaultFile, "configuration file")
	flag.BoolVar(&h, "h", false, "help")
	flag.Parse()
	if h {
		flag.Usage()
		return fmt.Errorf("invalid usage for command line")
	}
	return nil
}

func initConf() error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	Conf = &Config{}
	err = yaml.Unmarshal([]byte(data), Conf)
	return err
}

func initLog() error {
	level, err := tblog.ParseLevel(Conf.Log.Level)
	if err != nil {
		return err
	}
	tblog.SetLevel(level)
	RotateLog, err = rotatelogs.New(Conf.Log.File,
		rotatelogs.WithRotationCount(Conf.Log.MaxRolls),
		rotatelogs.WithRotationSize(Conf.Log.MaxSize))
	if err != nil {
		return err
	}
	tblog.SetOutput(RotateLog)
	return nil
}
