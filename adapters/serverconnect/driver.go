package serverconnect

import (
	"encoding/json"
	"io"
	"net/http"
)

type Driver struct {
	Url    string
	Client *http.Client
}

// FIXME this is actually just a "Connect to TURN provider" test
/*
This handles the actual connection to the TURN server.
Here we call the TURN provider and consider our client connected
if we receive a list of ICE Servers.
*/
func (d Driver) Connect() (connected bool, err error) {
	res, err := d.Client.Get(d.Url)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var responseServers []map[string]interface{}
	json.Unmarshal([]byte(responseData), &responseServers)

	return len(responseServers) > 0, nil
}
