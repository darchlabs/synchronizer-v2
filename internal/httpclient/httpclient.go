package httpclient

import (
	"context"
	"io/ioutil"
	"log"
	http "net/http"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

type BaseHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ValidatorFunc func(*http.Response) (*http.Response, error)

type Client struct {
	httpClient BaseHTTPClient
	rl         *rate.Limiter

	maxRetry int

	validate ValidatorFunc
}

type Options struct {
	MaxRetry        int
	MaxRequest      int
	WindowInSeconds int
}

func NewClient(opts *Options, client BaseHTTPClient) *Client {
	rl := rate.NewLimiter(rate.Limit(opts.MaxRequest), opts.WindowInSeconds)

	return &Client{
		httpClient: client,
		maxRetry:   opts.MaxRetry,
		rl:         rl,
	}
}

func (cl *Client) Do(req *http.Request) (res *http.Response, err error) {
	for i := 0; i < cl.maxRetry; i++ {
		res, err = cl.makeRequest(req)
		if err != nil {
			log.Printf("http: attempt [%d] Client.Do cl.makeRequest error %s", i, err.Error())
			continue
		}

		if cl.validate != nil {
			res, err = cl.validate(res)
			if err != nil {
				defer res.Body.Close()
				body, parseErr := ioutil.ReadAll(res.Body)
				if parseErr != nil {
					return nil, errors.Wrap(parseErr, "http: attempt [%d] Client.Do cl.validate ioutil.ReadAll error")
				}

				log.Printf(
					"http: attempt [%d] Client.Do cl.validate error %s.\n Response payload %s.\n Endpoint: [%s]",
					i,
					err.Error(),
					string(body),
					req.URL.EscapedPath(),
				)
				continue
			}
		}

		return res, nil
	}

	return nil, err
}

func (cl *Client) makeRequest(req *http.Request) (res *http.Response, err error) {
	ctx := context.Background()

	// This is a blocking call
	err = cl.rl.Wait(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "http: Client.makeRequest cl.rl.Wait error")
	}
	res, err = cl.httpClient.Do(req)
	if err != nil {
		log.Printf(
			"Requested endpoint [%s] error %s", req.URL.EscapedPath(), err.Error())
		return nil, errors.Wrap(err, "http: Client.makeRequest cl.httpClient.Do error")
	}

	return res, nil
}

func (cl *Client) RegisterValidate(fn ValidatorFunc) {
	cl.validate = fn
}
