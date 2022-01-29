package configs

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Conf struct {
	Token string
	Chat  int64
	Redis RedisConf
}

type RedisConf struct {
	Addr     string
	Password string
	DB       int
}

func InitConf() (*Conf, error) {
	var local bool
	flag.BoolVar(&local, "local", false, "хост")
	flag.Parse()
	return envVar(local)
}

func envVar(local bool) (*Conf, error) {

	if local {
		err := godotenv.Load(".env")
		if err != nil {
			return &Conf{}, err
		}
	}

	chat := os.Getenv("CHAT_ID")
	chatInt, err := strconv.Atoi(chat)
	if err != nil {
		println(err.Error())
		return &Conf{}, err
	}

	redisDBstr := os.Getenv("REDIS_DB")
	redisDB, err := strconv.Atoi(redisDBstr)
	if err != nil {
		println(err.Error())
		return &Conf{}, err
	}

	return &Conf{
		os.Getenv("TOKEN"),
		int64(chatInt),
		RedisConf{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASS"),
			DB:       redisDB,
		},
	}, nil
}
