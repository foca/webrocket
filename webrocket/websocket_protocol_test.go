// This package provides a hybrid of MQ and WebSockets server with
// support for horizontal scalability.
//
// Copyright (C) 2011 by Krzysztof Kowalik <chris@nu7hat.ch>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
package webrocket

import (
	"testing"
	"websocket"
	"time"
	"bytes"
	"log"
)

var (
	ws  *websocket.Conn
	err error
	e   Endpoint
	v   *Vhost
)

func init() {
	go func() {
		ctx := NewContext()
		ctx.SetLog(log.New(bytes.NewBuffer([]byte{}), "", log.LstdFlags))
		e = ctx.NewWebsocketEndpoint("", 9771)
		v, _ = ctx.AddVhost("/test")
		v.OpenChannel("test")
		v.OpenChannel("test2")
		e.ListenAndServe()
	}()
}

func wssend(t *testing.T, data interface{}) {
	wssendto(t, ws, data)
}

func wssendto(t *testing.T, ws *websocket.Conn, data interface{}) {
	err = websocket.JSON.Send(ws, data)
	if err != nil {
		t.Error(err)
	}
}

func wsrecv(t *testing.T) *Message {
	return wsrecvfrom(t, ws)
}

func wsrecvfrom(t *testing.T, ws *websocket.Conn) *Message {
	var resp map[string]interface{}
	err := websocket.JSON.Receive(ws, &resp)
	if err != nil {
		t.Error(err)
		return nil
	}
	msg, err := newMessage(resp)
	if err != nil {
		t.Error(err)
	}
	return msg
}

func wserr(msg *Message) string {
	s, _ := msg.Get("status").(string)
	return s
}

func wsdial() (*websocket.Conn, error) {
	return websocket.Dial("ws://127.0.0.1:9771/test", "ws", "http://127.0.0.1/")
}

func TestWebsocketConnect(t *testing.T) {
	ws, err = wsdial()
	if err != nil {
		t.Error(err)
	}
	resp := wsrecv(t)
	if resp.Event() != "__connected" {
		t.Errorf("Expected to receive the '__connected' event, given '%s'", resp.Event())
	}
}

func TestWebsocketBadRequest(t *testing.T) {
	ws.Write([]byte("foobar"))
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
	ws.Write([]byte("{}"))
	resp = wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
}

func TestWebsocketNotFound(t *testing.T) {
	ws.Write([]byte("{\"hello\": {}}"))
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
}

func TestWebsocketAuthWithMissingToken(t *testing.T) {
	wssend(t, map[string]interface{}{
		"auth": map[string]interface{}{"foo": "bar"},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
}

func TestWebsocketAuthWuthInvalidTokenValue(t *testing.T) {
	wssend(t, map[string]interface{}{
		"auth": map[string]interface{}{"tokena": map[string]interface{}{}},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	} 
}

func TestWebsocketAuthWithInvalidToken(t *testing.T) {
	wssend(t, map[string]interface{}{
		"auth": map[string]interface{}{"token": "invalid"},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Unauthorized" {
		t.Errorf("Expected 'Unauthorized' error")
	}
}

func TestWebsocketAuthWithValidToken(t *testing.T) {
	token := v.GenerateSingleAccessToken(".*")
	wssend(t, map[string]interface{}{
		"auth": map[string]interface{}{
			"token": token,
		},
	})
	resp := wsrecv(t)
	if resp.Event() != "__authenticated" {
		t.Errorf("Expected to receive the '__authenticated' event, given '%s'", resp.Event())
	}
}

func TestWebsocketSubscribeInvalidChannelName(t *testing.T) {
	wssend(t, map[string]interface{}{
		"subscribe": map[string]interface{}{"channel": "shit%dfsdf%#"},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Channel not found" {
		t.Errorf("Expected 'Channel not found' error")
	}
}

func TestWebsocketSubscribeEmptyChannelName(t *testing.T) {
	wssend(t, map[string]interface{}{
		"subscribe": map[string]interface{}{"channel": ""},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
}

func TestWebsocketSubscribeNoChannelName(t *testing.T) {
	wssend(t, map[string]interface{}{
		"subscribe": map[string]interface{}{"foo": ""},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
}

func TestWebsocketSubscribeAllowedChannel(t *testing.T) {
	wssend(t, map[string]interface{}{
		"subscribe": map[string]interface{}{"channel": "test"},
	})
	resp := wsrecv(t)
	if resp.Event() != "__subscribed" {
		t.Errorf("Expected to receive the '__subscribed' event, given '%s'", resp.Event())
	}
}

func TestWebsocketUnsubscribeEmptyChannelName(t *testing.T) {
	wssend(t, map[string]interface{}{
		"unsubscribe": map[string]interface{}{"channel": ""},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
}

func TestWebsocketUnsubscribeInvalidChannelName(t *testing.T) {
	wssend(t, map[string]interface{}{
		"unsubscribe": map[string]interface{}{"channel": "shit%dfsdf%#"},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Channel not found" {
		t.Errorf("Expected 'Channel not found' error")
	}
}

func TestWebsocketUnsubscribeNotSubscribedChannel(t *testing.T) {
	wssend(t, map[string]interface{}{
		"unsubscribe": map[string]interface{}{"channel": "test2"},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Not subscribed" {
		t.Errorf("Expected 'Not subscribed' error")
	}
}

func TestWebsocketUnsubscribeValidChannel(t *testing.T) {
	wssend(t, map[string]interface{}{
		"unsubscribe": map[string]interface{}{"channel": "test"},
	})
	resp := wsrecv(t)
	if resp.Event() != "__unsubscribed" {
		t.Errorf("Expected to receive the '__unsubscribed' event, given '%s'", resp.Event())
	}
}

func TestWebsocketBroadcastWhenNotSubscribingTheChannel(t *testing.T) {
	wssend(t, map[string]interface{}{
		"broadcast": map[string]interface{}{
			"channel": "test", "event": "foo", "data": map[string]interface{}{},
		},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Not subscribed" {
		t.Errorf("Expected 'Not subscibed' error")
	}
}

func TestWebsocketBroadcastWithInvalidData(t *testing.T) {
	wssend(t, map[string]interface{}{
		"broadcast": map[string]interface{}{
			"channel": "test", "data": map[string]interface{}{},
		},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
	wssend(t, map[string]interface{}{
		"broadcast": map[string]interface{}{
			"event": "test", "data": map[string]interface{}{},
		},
	})
	resp = wsrecv(t)
	if wserr(resp) != "Bad request" {
		t.Errorf("Expected 'Bad request' error")
	}
}

func TestWebsocketBroadcastWhenInvalidChannelGiven(t *testing.T) {
	wssend(t, map[string]interface{}{
		"broadcast": map[string]interface{}{
			"channel": "notexists", "event": "foo", "data": map[string]interface{}{},
		},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Channel not found" {
		t.Errorf("Expected 'Channel not found' error")
	}
}

func TestWebsocketBroadcastToNotSubscribedChannel(t *testing.T) {
	wssend(t, map[string]interface{}{
		"broadcast": map[string]interface{}{
			"channel": "test2", "event": "foo", "data": map[string]interface{}{},
		},
	})
	resp := wsrecv(t)
	if wserr(resp) != "Not subscribed" {
		t.Errorf("Expected 'Not subscribed' error")
	}	
}

func TestWebsocketBroadcastValidData(t *testing.T) {
	var wss [2]*websocket.Conn
	for i, _ := range wss {
		wss[i], _ = wsdial()
		wsrecvfrom(t, wss[i])
		wssendto(t, wss[i], map[string]interface{}{
			"subscribe": map[string]interface{}{"channel": "test"},
		})
		wsrecvfrom(t, wss[i])
	}
	wssend(t, map[string]interface{}{
			"subscribe": map[string]interface{}{"channel": "test"},
	})
	wsrecv(t)
	wssend(t, map[string]interface{}{"broadcast":
		map[string]interface{}{
			"channel": "test",
			"event": "hello",
			"data": map[string]interface{}{"foo": "bar"},
		},
	})
	time.Sleep(1e6)
	for i, _ := range wss {
		resp := wsrecvfrom(t, wss[i])
		if resp.Event() != "hello" {
			t.Errorf("Expected to broadcast the 'hello' event")
		}
		chanName, _ := resp.Get("channel").(string)
		if chanName != "test" {
			t.Errorf("Expected to broadcast on the 'test' channel")
		}
		sid, _ := resp.Get("sid").(string)
		if sid == "" {
			t.Errorf("Expected to append sender's sid to the broadcasted data")
		}
		foo, _ := resp.Get("foo").(string)
		if foo != "bar" {
			t.Errorf("Expected to broadcast passed data")
		}
	}
}

// TODO: test 'broadcast' with tigger
// TODO: test 'subscribe' and 'broadcast' for protected channels
// TODO: test 'trigger'
// TODO: test 'close'
