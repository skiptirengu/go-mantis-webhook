package mantis

import "fmt"

func Rest(method string) (string) {
	return fmt.Sprintf("%s/%s", RestEndpoint(), method)
}

func RestEndpoint() (string) {
	return fmt.Sprintf("%s/api/rest", getHost())
}
