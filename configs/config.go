package configs

import (
	"encoding/json"
	"flag"
	"github.com/joho/godotenv"
	"io/ioutil"
	"os"
	"strconv"
)

type Conf struct {
	Token string
	Chat  int64
}

func InitConf() *Conf {
	var local bool
	flag.BoolVar(&local, "local", false, "хост")
	flag.Parse()
	return envVar(local)
}

func envVar(local bool) *Conf {
	if local {
		err := godotenv.Load(".env")
		if err != nil {
			println(err.Error())
			return &Conf{}
		}
	}
	chat:=os.Getenv("CHAT_ID")
	chatInt,err:=strconv.Atoi(chat)
	if err != nil {
		println(err.Error())
		return &Conf{}
	}
	return &Conf{
		os.Getenv("TOKEN"),
		int64(chatInt),
	}
}

func InitUsers() (map[int64]bool, error) {
	var users map[int64]bool
	jsonFile, err := os.Open("users.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err != nil {
		return map[int64]bool{}, err
	}
	defer jsonFile.Close()
	err = json.Unmarshal(byteValue, &users)
	if err != nil {
		return map[int64]bool{}, err
	}
	return users, nil
}

func SaveUsers(users map[int64]bool) error {
	filePath := "users.json"
	jsonUsers, err := json.Marshal(users)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, jsonUsers, 0644)
	if err != nil {
		return err
	}
	return nil
}
