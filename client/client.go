package client

import (
	"github.com/fedstackjs/azukiiro/common"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var http = resty.New()

func GetDefaultHTTPClient() *resty.Client {
	return http
}

func InitFromConfig() {
	serverAddr := viper.GetString("serverAddr")
	if serverAddr == "" {
		logrus.Fatalln("Server address not set")
	}
	http.SetBaseURL(serverAddr)
	runnerId := viper.GetString("runnerId")
	if runnerId == "" {
		logrus.Fatalln("Runner ID not set")
	}
	runnerKey := viper.GetString("runnerKey")
	if runnerKey == "" {
		logrus.Fatalln("Runner key not set")
	}
	http.SetHeaders(map[string]string{
		"X-AOI-Runner-Id":  runnerId,
		"X-AOI-Runner-Key": runnerKey,
		"User-Agent":       common.GetVersion(),
	})
}
