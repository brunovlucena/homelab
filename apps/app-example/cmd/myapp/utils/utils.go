package utils

import "github.com/sirupsen/logrus"

func LogPrint(msg string) {
	if msg != "" {
		logrus.Println(msg)
	}
}

func LogErr(err error) {
	if err != nil {
		logrus.Println(err)
	}
}
