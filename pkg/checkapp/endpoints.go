package checkapp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig"
)

func Endpoint(ctx context.Context, input *Input) (string, int) {
	endpoint := getTestEndpoint(input, input.Index)
	client := endpoint.GetClient()
	body := bytes.NewBufferString(endpoint.Body)

	req, err := http.NewRequestWithContext(ctx, endpoint.Method, endpoint.GetURL(), body)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}

	endpoint.SetHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return err.Error(), http.StatusFailedDependency
	}
	defer resp.Body.Close()

	size, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}

	return resp.Status + " - Response Size: " + strconv.FormatInt(size, mnd.Base10), http.StatusOK
}

func getTestEndpoint(input *Input, index int) *epconfig.Endpoint {
	if len(input.Real.Endpoints) > index {
		return input.Real.Endpoints[index]
	}

	return input.Post.Endpoints[index]
}
