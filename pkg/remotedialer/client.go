package remotedialer

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ConnectAuthorizer func(proto, address string) bool

func ClientConnect(wsURL string, headers http.Header, dialer *websocket.Dialer, auth ConnectAuthorizer, onConnect func(context.Context) error) {
	if err := connectToProxy(wsURL, headers, auth, dialer, onConnect); err != nil {
		logrus.WithError(err).Error("Failed to connect to proxy")
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func connectToProxy(proxyURL string, headers http.Header, auth ConnectAuthorizer, dialer *websocket.Dialer, onConnect func(context.Context) error) error {
	logrus.WithField("url", proxyURL).Info("Connecting to proxy")

	if dialer == nil {
		dialer = &websocket.Dialer{}
	}
	ws, _, err := dialer.Dial(proxyURL, headers)
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to proxy")
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := onConnect(ctx); err != nil {
		return err
	}

	session := newClientSession(auth, ws)
	_, err = session.serve()
	session.Close()
	return err
}
