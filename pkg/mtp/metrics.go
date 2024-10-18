package mtp

import "sync/atomic"

// The needs of the program used should not be typical metrics,
// they are data that are intended to speed up the program's operation
// e.g. for Prometheus
type Metrics struct {
	// liczba podłączonych klientów
	connectedClients int32
}

func (m *Metrics) ClientConnected() int32 {
	return atomic.AddInt32(&m.connectedClients, 1)
}

func (m *Metrics) ConnectedClients() int32 {
	return atomic.LoadInt32(&m.connectedClients)
}

func (m *Metrics) ClientDisconnected() int32 {
	return atomic.AddInt32(&m.connectedClients, -1)
}
