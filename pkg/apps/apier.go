//nolint:wrapcheck
package apps

import (
	"context"
	"io"
	"net/url"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"golift.io/datacounter"
	"golift.io/starr"
)

/* The purpose of this code is to capture data counters to starr apps. */

type starrAPI struct {
	app string
	api starr.APIer
}

func (a *starrAPI) Login(ctx context.Context) error {
	exp.Apps.Add(a.app+"&&Logins", 1)
	return a.api.Login(ctx)
}

// Normal data, returns response body.
func (a *starrAPI) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	body, err := a.api.Get(ctx, path, params)
	if err != nil {
		exp.Apps.Add(a.app+"&&GET Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Received", int64(len(body)))
	exp.Apps.Add(a.app+"&&GET Requests", 1)

	return body, err
}

func (a *starrAPI) Post(ctx context.Context, path string, params url.Values, postBody io.Reader) ([]byte, error) {
	sent := datacounter.NewReaderCounter(postBody)

	body, err := a.api.Post(ctx, path, params, sent)
	if err != nil {
		exp.Apps.Add(a.app+"&&POST Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Received", int64(len(body)))
	exp.Apps.Add(a.app+"&&Bytes Sent", int64(sent.Count()))
	exp.Apps.Add(a.app+"&&POST Requests", 1)

	return body, err
}

func (a *starrAPI) Put(ctx context.Context, path string, params url.Values, putBody io.Reader) ([]byte, error) {
	sent := datacounter.NewReaderCounter(putBody)

	body, err := a.api.Put(ctx, path, params, sent)
	if err != nil {
		exp.Apps.Add(a.app+"&&PUT Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Received", int64(len(body)))
	exp.Apps.Add(a.app+"&&Bytes Sent", int64(sent.Count()))
	exp.Apps.Add(a.app+"&&PUT Requests", 1)

	return body, err
}

func (a *starrAPI) Delete(ctx context.Context, path string, params url.Values) ([]byte, error) {
	body, err := a.api.Delete(ctx, path, params)
	if err != nil {
		exp.Apps.Add(a.app+"&&DELETE Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Received", int64(len(body)))
	exp.Apps.Add(a.app+"&&DELETE Requests", 1)

	return body, err
}

func (a *starrAPI) GetInto(ctx context.Context, path string, params url.Values, output interface{}) (int64, error) {
	rcvd, err := a.api.GetInto(ctx, path, params, output)
	if err != nil {
		exp.Apps.Add(a.app+"&&GET Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Received", rcvd)
	exp.Apps.Add(a.app+"&&GET Requests", 1)

	return rcvd, err
}

func (a *starrAPI) PostInto(
	ctx context.Context,
	path string,
	params url.Values,
	postBody io.Reader,
	output interface{},
) (int64, error) {
	sent := datacounter.NewReaderCounter(postBody)

	rcvd, err := a.api.PostInto(ctx, path, params, sent, output)
	if err != nil {
		exp.Apps.Add(a.app+"&&POST Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Sent", int64(sent.Count()))
	exp.Apps.Add(a.app+"&&Bytes Received", rcvd)
	exp.Apps.Add(a.app+"&&POST Requests", 1)

	return rcvd, err
}

func (a *starrAPI) PutInto(
	ctx context.Context,
	path string,
	params url.Values,
	putBody io.Reader,
	output interface{},
) (int64, error) {
	sent := datacounter.NewReaderCounter(putBody)

	rcvd, err := a.api.PutInto(ctx, path, params, sent, output)
	if err != nil {
		exp.Apps.Add(a.app+"&&PUT Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Sent", int64(sent.Count()))
	exp.Apps.Add(a.app+"&&Bytes Received", rcvd)
	exp.Apps.Add(a.app+"&&PUT Requests", 1)

	return rcvd, err
}

func (a *starrAPI) DeleteInto(ctx context.Context, path string, params url.Values, output interface{}) (int64, error) {
	rcvd, err := a.api.DeleteInto(ctx, path, params, output)
	if err != nil {
		exp.Apps.Add(a.app+"&&DELETE Errors", 1)
	}

	exp.Apps.Add(a.app+"&&Bytes Received", rcvd)
	exp.Apps.Add(a.app+"&&DELETE Requests", 1)

	return rcvd, err
}

func (a *starrAPI) GetBody(ctx context.Context, path string, params url.Values) (io.ReadCloser, int, error) {
	exp.Apps.Add(a.app+"&&GET Requests", 1)

	resp, code, err := a.api.GetBody(ctx, path, params)
	if err != nil {
		exp.Apps.Add(a.app+"&&GET Errors", 1)
		return resp, code, err
	}

	rcvd := datacounter.NewReaderCounter(resp)

	return &FakeCloser{
		App:     a.app,
		Rcvd:    rcvd.Count,
		CloseFn: resp.Close,
		Reader:  rcvd,
	}, code, nil
}

func (a *starrAPI) PostBody(
	ctx context.Context,
	path string,
	params url.Values,
	postBody io.Reader,
) (io.ReadCloser, int, error) {
	exp.Apps.Add(a.app+"&&POST Requests", 1)

	sent := datacounter.NewReaderCounter(postBody)
	resp, code, err := a.api.PostBody(ctx, path, params, sent)
	exp.Apps.Add(a.app+"&&Bytes Sent", int64(sent.Count()))

	if err != nil {
		exp.Apps.Add(a.app+"&&POST Errors", 1)
		return resp, code, err
	}

	rcvd := datacounter.NewReaderCounter(resp)

	return &FakeCloser{
		App:     a.app,
		Rcvd:    rcvd.Count,
		CloseFn: resp.Close,
		Reader:  rcvd,
	}, code, nil
}

func (a *starrAPI) PutBody(
	ctx context.Context,
	path string,
	params url.Values,
	putBody io.Reader,
) (io.ReadCloser, int, error) {
	exp.Apps.Add(a.app+"&&PUT Requests", 1)

	sent := datacounter.NewReaderCounter(putBody)
	resp, code, err := a.api.PutBody(ctx, path, params, sent)
	exp.Apps.Add(a.app+"&&Bytes Sent", int64(sent.Count()))

	if err != nil {
		exp.Apps.Add(a.app+"&&PUT Errors", 1)
		return resp, code, err
	}

	rcvd := datacounter.NewReaderCounter(resp)

	return &FakeCloser{
		App:     a.app,
		Rcvd:    rcvd.Count,
		CloseFn: resp.Close,
		Reader:  rcvd,
	}, code, err
}

func (a *starrAPI) DeleteBody(ctx context.Context, path string, params url.Values) (io.ReadCloser, int, error) {
	exp.Apps.Add(a.app+"&&DELETE Requests", 1)

	resp, code, err := a.api.DeleteBody(ctx, path, params)
	if err != nil {
		exp.Apps.Add(a.app+"&&DELETE Errors", 1)
		return resp, code, err
	}

	rcvd := datacounter.NewReaderCounter(resp)

	return &FakeCloser{
		App:     a.app,
		Rcvd:    rcvd.Count,
		CloseFn: resp.Close,
		Reader:  rcvd,
	}, code, err
}

type FakeCloser struct {
	App     string
	Rcvd    func() uint64
	CloseFn func() error
	io.Reader
}

func (f *FakeCloser) Close() error {
	defer exp.Apps.Add(f.App+"&&Bytes Received", int64(f.Rcvd()))

	return f.CloseFn()
}
