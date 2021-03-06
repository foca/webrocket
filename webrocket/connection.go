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
	"log"
	"sync"
)

// Base connection wrapper.
type connection struct {
	vhost   *Vhost
	ctx     *Context
	log     *log.Logger
	mtx     sync.Mutex
	isAlive bool
}

// newConnection creates and initializes new connection.
func newConnection(vhost *Vhost) (c *connection) {
	c = &connection{isAlive: true}
	c.vhost = vhost
	c.ctx = vhost.ctx
	c.log = c.ctx.log
	return
}

// Returns true if this connection is alive.
func (c *connection) IsAlive() bool {
	return c.isAlive
}

// Turns off the connection's alive state.
func (c *connection) kill() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.isAlive = false
}
