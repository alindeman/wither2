package minecraft

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Client represents a Minecraft RCON client.
type Client struct {
	sync.Mutex

	network  string
	addr     string
	password string
	conn     net.Conn
}

const (
	typeLogin           uint32 = 3
	typeLoginResponse   uint32 = 2
	typeCommand         uint32 = 2
	typeCommandResponse uint32 = 0
)

type packet struct {
	RequestID uint32
	Type      uint32
	Payload   []byte
}

// 4 bytes for request ID and type. 2 bytes of null padding at the end.
const packetMinimumLength uint32 = uint32(4 + 4 + 2)

// See https://wiki.vg/RCON
func readPacket(r io.Reader) (*packet, error) {
	p := new(packet)

	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &p.RequestID); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &p.Type); err != nil {
		return nil, err
	}

	p.Payload = make([]byte, int(length-packetMinimumLength))
	if _, err := r.Read(p.Payload); err != nil {
		return nil, err
	}

	pad := make([]byte, 2)
	if _, err := r.Read(pad); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *packet) WriteTo(w io.Writer) (n int, err error) {
	buf := new(bytes.Buffer)

	var length uint32 = packetMinimumLength + uint32(len(p.Payload))
	if err := binary.Write(buf, binary.LittleEndian, length); err != nil {
		return 0, err
	}

	if err := binary.Write(buf, binary.LittleEndian, p.RequestID); err != nil {
		return 0, err
	}

	if err := binary.Write(buf, binary.LittleEndian, p.Type); err != nil {
		return 0, err
	}

	if _, err := buf.Write(p.Payload); err != nil {
		return 0, err
	}

	if _, err := buf.Write([]byte{0, 0}); err != nil {
		return 0, err
	}

	return w.Write(buf.Bytes())
}

// New returns a Client.
func New(network string, addr string, password string) *Client {
	return &Client{
		network:  network,
		addr:     addr,
		password: password,
	}
}

func (c *Client) Command(timeout time.Duration, command string) error {
	c.Lock()
	defer c.Unlock()

	deadline := time.Now().Add(timeout)

	if err := c.ensureConn(deadline); err != nil {
		return err
	}

	c.conn.SetDeadline(deadline)

	p := &packet{
		RequestID: 1,
		Type:      typeCommand,
		Payload:   []byte(command),
	}

	if _, err := p.WriteTo(c.conn); err != nil {
		c.disconnect() // reconnect next time
		return fmt.Errorf("failed to write packet: %w", err)
	}

	reply, err := readPacket(c.conn)
	if err != nil {
		c.disconnect() // reconnect next time
		return fmt.Errorf("failed to read response packet: %w", err)
	}

	if typeCommandResponse != reply.Type {
		return fmt.Errorf("expected reply type %v, got %v", typeCommandResponse, reply.Type)
	} else if p.RequestID != reply.RequestID {
		return fmt.Errorf("expected request ID %v, got %v", p.RequestID, reply.RequestID)
	}

	return nil
}

// ensureConn populates c.conn and performs a login, if needed. Must be
// called while holding the lock.
func (c *Client) ensureConn(deadline time.Time) error {
	if c.conn != nil {
		return nil
	}

	dialer := &net.Dialer{
		Deadline: deadline,
	}

	conn, err := dialer.Dial(c.network, c.addr)
	if err != nil {
		return err
	}

	// Login routine
	conn.SetDeadline(deadline)

	p := &packet{
		RequestID: 1,
		Type:      typeLogin,
		Payload:   []byte(c.password),
	}

	if _, err := p.WriteTo(conn); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	reply, err := readPacket(conn)
	if err != nil {
		return fmt.Errorf("failed to read response packet: %w", err)
	}

	if typeLoginResponse != reply.Type {
		return fmt.Errorf("expected reply type %v, got %v", typeLoginResponse, reply.Type)
	} else if p.RequestID != reply.RequestID {
		return fmt.Errorf("expected request ID %v, got %v", p.RequestID, reply.RequestID)
	}

	// Connection is ready to use
	c.conn = conn
	return nil
}

// disconnect disconnects and depopulates c.conn. Must be called while
// locking the lock.
func (c *Client) disconnect() {
	if c.conn == nil {
		return
	}

	_ = c.conn.Close()
	c.conn = nil

	return
}
