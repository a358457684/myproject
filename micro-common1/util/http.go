package util

import (
	"common/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HttpPost(urls string, params map[string]string, token string) ([]uint8, error) {

	DataUrlVal := url.Values{}
	for key, val := range params {
		DataUrlVal.Add(key, val)
	}

	req, err := http.NewRequest("POST", urls, strings.NewReader(DataUrlVal.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("token", token)

	clt := http.Client{}
	//clt.Do(req)

	resp, err := clt.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("bdhttpPost body %s", err.Error())
		return nil, err
	}
	return body, nil
}
