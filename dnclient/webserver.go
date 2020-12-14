package dnclient

import (
	"encoding/json"
	"net/http"
)

func (c *Client) Run() {
	http.HandleFunc("/", c.Handle)
	err := http.ListenAndServe(c.BindAddr, nil)
	if err != nil && err != http.ErrServerClosed {
		//
	}
}

func (c *Client) Handle(w http.ResponseWriter, r *http.Request) {
	payload := &IncomingPayload{}

	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("rekt"))

		return
	}

	if c.Config.APIKey != payload.Key {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

}
