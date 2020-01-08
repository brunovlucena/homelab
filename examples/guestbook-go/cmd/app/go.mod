module github.com/brunovlucena/guestbook-go/app

go 1.13

replace github.com/brunovlucena/guestbook-go/cmd/utils => ../utils

require (
	github.com/brunovlucena/guestbook-go/cmd/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/viper v1.6.1 // indirect
)
