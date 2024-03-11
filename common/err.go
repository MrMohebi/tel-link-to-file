package common

import "log"

func IsErr(err error, fatal bool, message ...string) {

	if err != nil {

		if len(message) != 0 {
			if fatal {
				log.Fatal(message)
			} else {
				log.Println(message)
			}
			return
		}
		if fatal {
			log.Fatal(err)
		} else {
			log.Println(err)
		}
	}

}
