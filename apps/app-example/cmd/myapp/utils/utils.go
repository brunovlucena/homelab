package utils

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

func LogPrint(msg string) {
	if msg != "" {
		logrus.Info(msg)
	}
}

func LogErr(err error) {
	if err != nil {
		logrus.Error(err)
	}
}

func LoadJson(filePath string, configs *[]map[string]interface{}) {
	// Open our jsonFile
	jsonArrayFile, err := os.Open(filePath)
	// if we os.Open returns an error then handle it
	LogErr(err)
	LogPrint("Successfully Opened " + filePath)
	// defer the closing of our jsonFile
	defer jsonArrayFile.Close()
	// read our opened json
	byteValue, _ := ioutil.ReadAll(jsonArrayFile)
	// we unmarshal our byteArray
	raw := json.RawMessage(string(byteValue))
	bytes, err := raw.MarshalJSON()
	json.Unmarshal(bytes, configs)
}

// this helper function returns the ipv4 address from the server
func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		LogErr(err)
		return "error"
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	panic("Unable to determine local IP address (non loopback). Exiting.")
}
