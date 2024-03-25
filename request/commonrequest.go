package request

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"

	"github.com/onfirebyte/chatt/common"
	"github.com/onfirebyte/chatt/dto"
)

func CreateUser(name string, password string) (string, error) {
	reqUrl, err := url.Parse(common.URL)
	if err != nil {
		log.Println(err)
	}

	reqUrl.Path = "/users"

	q := reqUrl.Query()
	q.Set("user", name)
	q.Set("password", password)
	reqUrl.RawQuery = q.Encode()

	resp, err := http.Post(reqUrl.String(), "application/json", nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Println("BODY", string(body))

	// check if status code is 2xx
	if resp.StatusCode/100 != 2 {
		if resp.StatusCode == http.StatusUnauthorized {
			return "", fmt.Errorf("Please provide a valid password")
		}
		return "", fmt.Errorf("Error: %s", string(body))
	}

	log.Println("Token acquired", string(body))
	return string(body), nil
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

func GetAllRooms() ([]dto.Room, error) {
	resp, err := http.Get(fmt.Sprintf("%s/rooms", common.URL))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var res []dto.Room
	err = json.NewDecoder(resp.Body).Decode(&res)

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res, err
}
