package request

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"

	"github.com/onfirebyte/chatt/common"
)

func CreateUser(name string) error {
	reqUrl, err := url.Parse(common.URL)
	if err != nil {
		log.Println(err)
	}

	reqUrl.Path = "/users"

	q := reqUrl.Query()
	q.Set("user", name)
	reqUrl.RawQuery = q.Encode()

	resp, err := http.Post(reqUrl.String(), "application/json", nil)
	if err != nil {
		log.Println(err)
		return err
	}

	defer resp.Body.Close()
	var body []byte

	resp.Body.Read(body)

	// check if status code is 2xx
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Error: %s", string(body))
	}

	return nil
}

func GetAllUsers() ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/users", common.URL))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var res []string
	err = json.NewDecoder(resp.Body).Decode(&res)

	sort.Strings(res)

	return res, err
}

func GetAllRooms() ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/rooms", common.URL))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var res []string
	err = json.NewDecoder(resp.Body).Decode(&res)

	sort.Strings(res)

	return res, err
}
