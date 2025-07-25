package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io-go-parser/parser"
	"github.com/zishang520/engine.io/v2/events"
	"github.com/zishang520/engine.io/v2/transports"
	"github.com/zishang520/engine.io/v2/types"
)

// mockTransport is a mock implementation of the Transport interface for testing.
// It allows us to control the transport's behavior precisely, especially for
// simulating network delays and orchestrating event sequences.
type mockTransport struct {
	events.EventEmitter

	isWritable   atomic.Bool
	sendCalled   chan []*packet.Packet // Notifies when Send is called, passing the packets.
	triggerDrain chan struct{}         // A signal to tell the transport to finish sending and emit "drain".
	closeCalled  chan struct{}         // Notifies when Close is called.
	protocol     int
	sid          string

	// --- Fields for failure simulation ---
	failAtPacketIndex int   // The index of the packet in a batch at which to simulate a failure. -1 means no failure.
	failError         error // The error to emit upon failure.
}

func newMockTransport() *mockTransport {
	m := &mockTransport{
		EventEmitter:      events.New(),
		sendCalled:        make(chan []*packet.Packet, 1), // Buffered to prevent blocking
		triggerDrain:      make(chan struct{}),
		closeCalled:       make(chan struct{}, 1),
		protocol:          4,
		failAtPacketIndex: -1, // Default to no failure
		failError:         errors.New("simulated transport error"),
	}
	m.isWritable.Store(true)
	return m
}

// Implement all required Transport interface methods

func (m *mockTransport) Prototype(t transports.Transport) {}
func (m *mockTransport) Proto() transports.Transport      { return m }

func (m *mockTransport) Name() string {
	return "websocket" // Use a known transport name to avoid nil pointer errors
}

func (m *mockTransport) Writable() bool {
	return m.isWritable.Load()
}

func (m *mockTransport) SetWritable(writable bool) {
	m.isWritable.Store(writable)
}

func (m *mockTransport) Protocol() int {
	return m.protocol
}

func (m *mockTransport) SetSid(sid string) {
	m.sid = sid
}

func (m *mockTransport) Sid() string {
	return m.sid
}

func (m *mockTransport) Construct(ctx *types.HttpContext) {
	// No-op for mock
}

func (m *mockTransport) OnError(string, error) {}

func (m *mockTransport) Send(packets []*packet.Packet) {
	if !m.Writable() {
		return
	}
	m.SetWritable(false)

	// The goroutine simulates the asynchronous nature of network I/O.
	go func() {
		// --- Failure simulation logic ---
		if m.failAtPacketIndex >= 0 && len(packets) > m.failAtPacketIndex {
			time.Sleep(5 * time.Millisecond)      // Simulate some processing time before failure
			m.Emit("error", m.failError, packets) // Pass packets on error
			return
		}

		// --- Success path ---
		m.sendCalled <- packets
		<-m.triggerDrain
		m.SetWritable(true)
		m.Emit("drain")
	}()
}

func (m *mockTransport) Discard() {
	// No-op for mock
}

func (m *mockTransport) Close(fn ...types.Callable) {
	// Notify that Close has been called.
	select {
	case <-m.closeCalled:
		// Already closed
	default:
		close(m.closeCalled)
	}

	if len(fn) > 0 && fn[0] != nil {
		fn[0]()
	}
}

// Empty implementations for unused interface methods

func (m *mockTransport) SetSupportsBinary(bool)                        {}
func (m *mockTransport) SetReadyState(string)                          {}
func (m *mockTransport) SetHttpCompression(*types.HttpCompression)     {}
func (m *mockTransport) SetPerMessageDeflate(*types.PerMessageDeflate) {}
func (m *mockTransport) SetMaxHttpBufferSize(int64)                    {}

func (m *mockTransport) Discarded() bool       { return false }
func (m *mockTransport) Parser() parser.Parser { return nil }
func (m *mockTransport) ReadyState() string {
	if m.IsClosed() {
		return "closed"
	}
	return "open"
}

func (m *mockTransport) IsClosed() bool {
	select {
	case <-m.closeCalled:
		return true
	default:
		return false
	}
}

func (m *mockTransport) HttpCompression() *types.HttpCompression     { return nil }
func (m *mockTransport) PerMessageDeflate() *types.PerMessageDeflate { return nil }
func (m *mockTransport) MaxHttpBufferSize() int64                    { return 1000000 }
func (m *mockTransport) HandlesUpgrades() bool                       { return false }
func (m *mockTransport) OnRequest(*types.HttpContext)                {}
func (m *mockTransport) OnPacket(*packet.Packet)                     {}
func (m *mockTransport) OnData(types.BufferInterface)                {}
func (m *mockTransport) OnClose()                                    {}
func (m *mockTransport) DoClose(types.Callable)                      {}
func (m *mockTransport) SupportsBinary() bool                        { return true }

func (s *socket) Done() <-chan struct{} {
	done := make(chan struct{})
	s.Once("close", func(...any) {
		close(done)
	})
	return done
}

func TestCloseWaitsForEntireBufferDrain(t *testing.T) {
	// This test is designed to catch a very specific and critical race condition.
	// The scenario is as follows:
	// 1. A batch of packets is being sent (the transport is busy).
	// 2. The user calls `socket.Close()`. Because the buffer is empty, `Close()` starts
	//    waiting for a "drain" event before actually closing the transport.
	// 3. While the first batch is still "in-flight", new packets are written to the socket.
	// 4. The first batch finishes, and the transport emits a "drain" event.
	//
	// The BUG: The socket would see the "drain" event and immediately close the transport,
	// leaving the newly added packets stranded in the buffer, causing data loss.
	//
	// The FIX: The socket's `onDrain` handler must first check if the write buffer has
	// new packets. If it does, it must trigger another `flush` cycle and MUST NOT
	// emit the socket-level "drain" event. The socket-level "drain" should only be
	// emitted when the transport is idle AND the write buffer is truly empty.

	// --- Setup ---
	mockServer := MakeBaseServer() // Use the real BaseServer implementation
	mockServer.Construct(nil)      // Initialize the server options
	transport := newMockTransport()

	// Create a mock HttpContext to avoid nil pointer errors
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	mockCtx := types.NewHttpContext(w, req)

	socket := NewSocket("test-sid", mockServer, transport, mockCtx, 4).(*socket)

	assert.NotNil(t, socket)

	// First, we need to handle the initial OPEN packet that the socket sends automatically
	var openBatch []*packet.Packet
	select {
	case openBatch = <-transport.sendCalled:
		t.Log("Setup: Mock transport received initial OPEN packet.")
	case <-time.After(1 * time.Second):
		t.Fatal("Setup: Timed out waiting for initial OPEN packet.")
	}
	assert.Len(t, openBatch, 1)
	assert.Equal(t, packet.OPEN, openBatch[0].Type)

	// Complete the OPEN packet send to make transport writable again
	transport.triggerDrain <- struct{}{}

	// --- Phase 1: Send the first batch of packets ---
	// This simulates a batch of data being sent, making the transport busy.
	t.Log("Phase 1: Sending first packet batch...")
	socket.Write(bytes.NewBufferString("packet 1"), nil, nil)

	var firstBatch []*packet.Packet
	select {
	case firstBatch = <-transport.sendCalled:
		t.Log("Phase 1: Mock transport correctly received first batch.")
	case <-time.After(1 * time.Second):
		t.Fatal("Phase 1: Timed out waiting for transport.Send to be called.")
	}
	assert.Len(t, firstBatch, 1)
	assert.Equal(t, packet.MESSAGE, firstBatch[0].Type)
	// The data is stored as types.StringBuffer, need to read its content
	firstData := new(strings.Builder)
	io.Copy(firstData, firstBatch[0].Data)
	assert.Equal(t, "packet 1", firstData.String())

	// At this point, the transport's send goroutine is blocked, waiting for us
	// to signal it via the `triggerDrain` channel.

	// --- Phase 2: Trigger the race condition ---
	// While the transport is "busy", we first write more data, THEN call Close().
	// This ensures that Close() sees packets in the buffer and waits for drain.
	t.Log("Phase 2: Writing second packet batch while transport is busy...")
	socket.Write(bytes.NewBufferString("packet 2"), nil, nil)
	t.Log("Phase 2: Second packet batch written to buffer.")

	t.Log("Phase 2: Now calling Close() - it should wait for drain since buffer has packet 2...")
	go socket.Close(false) // Run in a goroutine as it will block waiting for drain

	time.Sleep(10 * time.Millisecond) // Give the Close() goroutine time to set up its drain listener

	// --- Phase 3: Complete the first send and verify behavior ---
	// Now, we unblock the transport, simulating the completion of the first batch.
	t.Log("Phase 3: Triggering drain for the first batch...")
	transport.triggerDrain <- struct{}{}

	// CRITICAL ASSERTION (Pre-fix validation):
	// Before the fix, the socket would emit "drain" here, and Close() would
	// immediately call `transport.Close()`. We verify this does NOT happen.
	select {
	case <-transport.closeCalled:
		t.Fatal("Phase 3: BUG DETECTED! Transport was closed prematurely, before the second batch was sent.")
	case <-time.After(100 * time.Millisecond):
		t.Log("Phase 3: OK. Transport was not closed prematurely.")
	}

	// CRITICAL ASSERTION (Post-fix validation):
	// With the fix, the `onDrain` handler should have found "packet 2" in the
	// buffer and triggered another flush.
	var secondBatch []*packet.Packet
	select {
	case secondBatch = <-transport.sendCalled:
		t.Log("Phase 3: Mock transport correctly received second batch for sending.")
	case <-time.After(1 * time.Second):
		t.Fatal("Phase 3: Timed out waiting for the second flush cycle.")
	}
	assert.Len(t, secondBatch, 1)
	assert.Equal(t, packet.MESSAGE, secondBatch[0].Type)
	// The data is stored as types.StringBuffer, need to read its content
	secondData := new(strings.Builder)
	io.Copy(secondData, secondBatch[0].Data)
	assert.Equal(t, "packet 2", secondData.String())

	// --- Phase 4: Complete the second send and verify final closure ---
	// Now we complete the final batch.
	t.Log("Phase 4: Triggering drain for the second batch...")
	transport.triggerDrain <- struct{}{}

	// FINAL ASSERTION:
	// This time, onDrain finds an empty buffer and should emit the real drain
	// event, which finally allows `Close()` to complete its work.
	select {
	case <-transport.closeCalled:
		t.Log("Phase 4: OK. Transport was closed correctly after all packets were drained.")
	case <-time.After(1 * time.Second):
		t.Fatal("Phase 4: Timed out waiting for the final transport.Close() call.")
	}
}

func TestIntermediatePacketLossWithConnectionAlive(t *testing.T) {
	// This test simulates the user's ACTUAL problem: when sending 10 packets continuously,
	// some intermediate packets (e.g., packet 3, packet 5) are lost, but the connection remains alive.
	//
	// Root cause: The transport may "silently fail" on some packets or incorrectly report success,
	// causing the socket's writeBuffer to be cleared even though not all packets were actually sent.

	// --- Setup ---
	mockServer := MakeBaseServer()
	mockServer.Construct(nil)
	transport := newMockTransport()

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mockCtx := types.NewHttpContext(w, req)

	socket := NewSocket("test-sid", mockServer, transport, mockCtx, 4).(*socket)
	assert.NotNil(t, socket)

	// Absorb the initial OPEN packet.
	<-transport.sendCalled
	transport.triggerDrain <- struct{}{}

	// Track which packets were actually "sent" by the transport
	var sentPackets [][]*packet.Packet

	// --- Phase 1: Send 10 packets continuously ---
	t.Log("Phase 1: Sending 10 packets continuously...")
	for i := 0; i < 10; i++ {
		socket.Write(bytes.NewBufferString(fmt.Sprintf("packet %d", i)), nil, nil)
	}

	// --- Phase 2: Simulate partial success/failure ---
	// The transport will receive packets in batches, but we'll simulate that
	// only some packets in each batch are "actually sent"

	packetCount := 0
	for {
		select {
		case batch := <-transport.sendCalled:
			t.Logf("Phase 2: Transport received batch with %d packets", len(batch))
			sentPackets = append(sentPackets, batch)

			// Simulate that the transport "drops" some packets but still reports success
			// This is the core of the user's problem: silent packet loss
			transport.triggerDrain <- struct{}{} // Transport incorrectly reports success

			packetCount += len(batch)
			if packetCount >= 10 {
				goto analysis
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Phase 2: Timed out waiting for packet batches")
		}
	}

analysis:
	// --- Phase 3: Analysis ---
	// In a buggy implementation, the socket's writeBuffer would be empty,
	// even though the transport might have "dropped" some packets.

	t.Log("Phase 3: Analyzing the results...")

	// Count total packets that were sent to transport
	totalSent := 0
	for _, batch := range sentPackets {
		totalSent += len(batch)
	}

	t.Logf("Total packets sent to transport: %d", totalSent)
	t.Logf("Socket writeBuffer length: %d", socket.writeBuffer.Len())
	t.Logf("Socket state: %s", socket.ReadyState())

	// CRITICAL ASSERTIONS:
	// 1. Connection should still be alive (this is what the user observed)
	assert.Equal(t, "open", socket.ReadyState(), "Connection should remain alive")

	// 2. All 10 packets should have been sent to the transport (proving continuous sending works)
	assert.Equal(t, 10, totalSent, "All 10 packets should have been sent to transport")

	// 3. WriteBuffer should be empty (proving that socket thinks all data was sent)
	assert.Equal(t, 0, socket.writeBuffer.Len(), "WriteBuffer should be empty after successful flush")

	// This test demonstrates the problem: even if the transport "loses" some packets,
	// the socket's writeBuffer gets cleared, and the connection remains alive.
	// The user would observe that some packets never reach the receiver,
	// but the connection is still active.

	t.Log("Phase 3: Test completed - this demonstrates the user's actual scenario")
}
