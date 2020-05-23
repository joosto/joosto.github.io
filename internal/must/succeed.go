package must

import "log"

func Succeed(err error) {
	if err != nil {
		log.Fatalf("error: %+v", err)
	}
}
