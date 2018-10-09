package clients

import (
	"context"
	"time"

	"github.com/containerum/cherry"
	"github.com/containerum/kube-client/pkg/model"
	utils "github.com/containerum/utils/httputil"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

// EventsClient is an interface to events-api.
type EventsClient interface {
	UserRegistered(ctx context.Context, userName string) error
	UserActivated(ctx context.Context, userName string) error
	UserDeleted(ctx context.Context, userName string) error
	GroupCreated(ctx context.Context, groupName string) error
	GroupDeleted(ctx context.Context, groupName string) error
	UserAddedToGroup(ctx context.Context, userName, groupName string) error
	UserRemovedFromGroup(ctx context.Context, userName, groupName string) error
}

type httpEventsClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

// NewHTTPEventsClient returns client for events-api working via restful api
func NewHTTPEventsClient(serverURL string) EventsClient {
	log := logrus.WithField("component", "events_api_client")
	client := resty.New().
		SetHostURL(serverURL).
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetTimeout(30*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetError(cherry.Err{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpEventsClient{
		rest: client,
		log:  log,
	}
}

func (c *httpEventsClient) UserRegistered(ctx context.Context, userName string) error {
	var event = model.Event{
		Kind:         model.EventInfo,
		Time:         time.Now().Format(time.RFC3339),
		Name:         model.UserRegistered,
		ResourceType: model.TypeUser,
		ResourceName: userName,
	}
	return sendUserEvent(c, ctx, event)
}

func (c *httpEventsClient) UserActivated(ctx context.Context, userName string) error {
	var event = model.Event{
		Kind:         model.EventInfo,
		Time:         time.Now().Format(time.RFC3339),
		Name:         model.UserActivated,
		ResourceType: model.TypeUser,
		ResourceName: userName,
	}
	return sendUserEvent(c, ctx, event)
}

func (c *httpEventsClient) UserDeleted(ctx context.Context, userName string) error {
	var event = model.Event{
		Kind:         model.EventInfo,
		Time:         time.Now().Format(time.RFC3339),
		Name:         model.UserDeleted,
		ResourceType: model.TypeUser,
		ResourceName: userName,
	}
	return sendUserEvent(c, ctx, event)
}

func (c *httpEventsClient) GroupCreated(ctx context.Context, groupName string) error {
	var event = model.Event{
		Kind:         model.EventInfo,
		Time:         time.Now().Format(time.RFC3339),
		Name:         model.GroupCreated,
		ResourceType: model.TypeUser,
		ResourceName: groupName,
	}
	return sendUserEvent(c, ctx, event)
}

func (c *httpEventsClient) GroupDeleted(ctx context.Context, groupName string) error {
	var event = model.Event{
		Kind:         model.EventInfo,
		Time:         time.Now().Format(time.RFC3339),
		Name:         model.GroupDeleted,
		ResourceType: model.TypeUser,
		ResourceName: groupName,
	}
	return sendUserEvent(c, ctx, event)
}

func (c *httpEventsClient) UserAddedToGroup(ctx context.Context, userName, groupName string) error {
	var event = model.Event{
		Kind:         model.EventInfo,
		Time:         time.Now().Format(time.RFC3339),
		Name:         model.UserAddedToGroup,
		ResourceType: model.TypeUser,
		ResourceName: userName,
		Details: map[string]string{
			"groupname": groupName,
		},
	}
	return sendUserEvent(c, ctx, event)
}

func (c *httpEventsClient) UserRemovedFromGroup(ctx context.Context, userName, groupName string) error {
	var event = model.Event{
		Kind:         model.EventInfo,
		Time:         time.Now().Format(time.RFC3339),
		Name:         model.UserRemovedFromGroup,
		ResourceType: model.TypeUser,
		ResourceName: userName,
		Details: map[string]string{
			"groupname": groupName,
		},
	}
	return sendUserEvent(c, ctx, event)
}

func sendUserEvent(c *httpEventsClient, ctx context.Context, event model.Event) error {
	headersMap := utils.RequestHeadersMap(ctx)

	resp, err := c.rest.R().SetContext(ctx).
		SetHeaders(headersMap).
		SetBody(event).
		Post("/events/containerum/users")
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error().(*cherry.Err)
	}
	return nil
}
