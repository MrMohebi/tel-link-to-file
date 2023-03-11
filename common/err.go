package common

import "log"

func IsErr(err error, message ...string) {
	if err != nil {
		if len(message) != 0 {
			log.Fatal(message)
			return
		}
		log.Fatal(err)
	}
}
