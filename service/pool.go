package service

// ConnectionPool maintains the set of active clients and broadcasts messages to the
// clients.
type ConnectionPool struct {
	// Registered clients.
	clients map[chan []byte]bool

	// Register requests from the clients.
	Register chan chan []byte

	// Unregister requests from clients.
	Unregister chan chan []byte
}

func NewPool() *ConnectionPool {
	return &ConnectionPool{
		Register:   make(chan chan []byte),
		Unregister: make(chan chan []byte),
		clients:    make(map[chan []byte]bool),
	}
}

func (h *ConnectionPool) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if !IsClosed(client) {
					close(client)
				}
			}
		}
	}
}
