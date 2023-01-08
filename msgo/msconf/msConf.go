package msconf

import (
	"errors"
	"flag"
	"github.com/BurntSushi/toml"
	"os"
)

type MsConf struct {
	Log      map[string]any
	Db       map[string]any
	Redis    map[string]any
	Template map[string]any
	Pool     map[string]any
}

var Conf = &MsConf{}

func init() {
	loadToml(Conf)
}

func loadToml(conf *MsConf) {

	confPara := flag.String("conf", "conf/conf.toml", "app config file")
	flag.Parse()
	if *confPara == "" {
		panic(errors.New("conf para is empty"))
	}

	_, err := os.Stat(*confPara)
	if err != nil {
		panic(errors.New("conf.toml file is not exists"))
	}

	_, err = toml.DecodeFile(*confPara, conf)
	if err != nil {
		panic(errors.New("conf.toml file parse error"))
	}

}
