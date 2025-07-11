package querator

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/duh-rpc/duh-go"
	v1 "github.com/duh-rpc/duh-go/proto/v1"
	"github.com/kapetan-io/querator/internal"
	pb "github.com/kapetan-io/querator/proto"
	"github.com/kapetan-io/querator/transport"
	"github.com/kapetan-io/tackle/clock"
	"github.com/kapetan-io/tackle/set"
	"google.golang.org/protobuf/proto"
	"net/http"
)

const (
	MsgRequestTimeout    = internal.MsgRequestTimeout
	MsgDuplicateClientID = internal.MsgDuplicateClientID
	MsgServiceInShutdown = internal.MsgServiceInShutdown
	MsgQueueInShutdown   = internal.MsgQueueInShutdown
	MsgQueueOverLoaded   = internal.MsgQueueOverLoaded
)

type ListOptions struct {
	Pivot string
	Limit int
}

type ClientConfig struct {
	// Users can provide their own http client with TLS config if needed
	Client *http.Client
	// The address of endpoint in the format `<scheme>://<host>:<port>`
	Endpoint string
}

type Client struct {
	client *duh.Client
	conf   ClientConfig
}

// NewClient creates a new instance of the Gubernator user client
func NewClient(conf ClientConfig) (*Client, error) {
	set.Default(&conf.Client, &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:     5_000,
			MaxIdleConns:        5_000,
			MaxIdleConnsPerHost: 5_000,
			IdleConnTimeout:     60 * clock.Second,
		},
	})

	if len(conf.Endpoint) == 0 {
		return nil, errors.New("conf.Endpoint is empty; must provide an http endpoint")
	}

	return &Client{
		client: &duh.Client{
			Client: conf.Client,
		},
		conf: conf,
	}, nil
}

func (c *Client) QueueProduce(ctx context.Context, req *pb.QueueProduceRequest) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueueProduce), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

func (c *Client) QueueLease(ctx context.Context, req *pb.QueueLeaseRequest, res *pb.QueueLeaseResponse) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueueLease), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.client.Do(r, res)
}

func (c *Client) QueueComplete(ctx context.Context, req *pb.QueueCompleteRequest) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueueComplete), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

func (c *Client) QueueRetry(ctx context.Context, req *pb.QueueRetryRequest) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueueRetry), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

func (c *Client) QueueClear(ctx context.Context, req *pb.QueueClearRequest) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueueClear), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

// -------------------------------------------------
// API to manage lists of queues
// -------------------------------------------------

func (c *Client) QueuesCreate(ctx context.Context, req *pb.QueueInfo) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueuesCreate), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

func (c *Client) QueuesList(ctx context.Context, res *pb.QueuesListResponse, opts *ListOptions) error {
	var req pb.QueuesListRequest
	if opts != nil {
		req.Limit = int32(opts.Limit)
		req.Pivot = opts.Pivot
	}

	payload, err := proto.Marshal(&req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueuesList), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.client.Do(r, res)
}

func (c *Client) QueuesUpdate(ctx context.Context, req *pb.QueueInfo) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueuesUpdate), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

func (c *Client) QueuesDelete(ctx context.Context, req *pb.QueuesDeleteRequest) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueuesDelete), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

func (c *Client) QueuesInfo(ctx context.Context, req *pb.QueuesInfoRequest, res *pb.QueueInfo) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueuesInfo), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.client.Do(r, res)
}

// TODO: Write an iterator we can use to iterate through list APIs
// TODO(scheduled): Listing enqueued items should NOT include scheduled items

// StorageItemsList lists current items in the queue.
// # NOTE
// If the pivot does not exist when calling StorageItemsList(), the endpoint will return the
// nearest next item in the list to the pivot provided. It is up to the caller to verify the
// list of items returned begins with the id specified in the pivot. This allows users to iterate
// through a constantly moving list without constantly running into "pivot not found" errors.
func (c *Client) StorageItemsList(ctx context.Context, name string, partition int, res *pb.StorageItemsListResponse,
	opts *ListOptions) error {

	if opts == nil {
		opts = &ListOptions{}
	}

	req := pb.StorageItemsListRequest{
		Limit:     int32(opts.Limit),
		Partition: int32(partition),
		Pivot:     opts.Pivot,
		QueueName: name,
	}

	payload, err := proto.Marshal(&req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCStorageItemsList), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.client.Do(r, res)
}

// StorageScheduledList lists scheduled items in the partition.
// # NOTE
// If the pivot does not exist when calling StorageScheduledList(), the endpoint will return the
// nearest next item in the list to the pivot provided. It is up to the caller to verify the
// list of items returned begins with the id specified in the pivot. This allows users to iterate
// through a constantly moving list without constantly running into "pivot not found" errors.
func (c *Client) StorageScheduledList(ctx context.Context, name string, partition int, res *pb.StorageItemsListResponse,
	opts *ListOptions) error {

	if opts == nil {
		opts = &ListOptions{}
	}

	req := pb.StorageItemsListRequest{
		Limit:     int32(opts.Limit),
		Partition: int32(partition),
		Pivot:     opts.Pivot,
		QueueName: name,
	}

	payload, err := proto.Marshal(&req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCStorageScheduledList), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.client.Do(r, res)
}

func (c *Client) StorageItemsImport(ctx context.Context, req *pb.StorageItemsImportRequest,
	res *pb.StorageItemsImportResponse) error {

	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCStorageItemsImport), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.client.Do(r, res)
}

func (c *Client) StorageItemsDelete(ctx context.Context, req *pb.StorageItemsDeleteRequest) error {

	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCStorageItemsDelete), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var res v1.Reply
	return c.client.Do(r, &res)
}

func (c *Client) QueueStats(ctx context.Context, req *pb.QueueStatsRequest,
	res *pb.QueueStatsResponse) error {

	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError("while marshaling request payload: %w", err, nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s%s", c.conf.Endpoint, transport.RPCQueueStats), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError("", err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.client.Do(r, res)
}

// WithNoTLS returns ClientConfig suitable for use with NON-TLS clients
func WithNoTLS(address string) ClientConfig {
	return ClientConfig{
		Endpoint: fmt.Sprintf("http://%s", address),
		Client: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:     2_000,
				MaxIdleConns:        2_000,
				MaxIdleConnsPerHost: 2_000,
				IdleConnTimeout:     60 * clock.Second,
			},
		},
	}
}

// WithTLS returns ClientConfig suitable for use with TLS clients
func WithTLS(tls *tls.Config, address string) ClientConfig {
	return ClientConfig{
		Endpoint: fmt.Sprintf("https://%s", address),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     tls,
				MaxConnsPerHost:     2_000,
				MaxIdleConns:        2_000,
				MaxIdleConnsPerHost: 2_000,
				IdleConnTimeout:     60 * clock.Second,
			},
		},
	}
}

type ItemsWithIDs interface {
	GetId() string
}

func CollectIDs[S ~[]E, E ItemsWithIDs](items S) []string {
	var result []string
	for _, v := range items {
		result = append(result, v.GetId())
	}
	return result
}