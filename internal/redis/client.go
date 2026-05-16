package redis

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Client is a minimal Redis client using raw TCP (no external deps)
type Client struct {
	conn net.Conn
	rw   *bufio.ReadWriter
}

// New connects to Redis at addr (e.g. "localhost:6379")
func New(addr, password string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to Redis at %s: %w", addr, err)
	}

	c := &Client{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	if password != "" {
		resp, err := c.Do("AUTH", password)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("Redis AUTH failed: %w", err)
		}
		if resp != "OK" {
			conn.Close()
			return nil, fmt.Errorf("Redis AUTH rejected: %s", resp)
		}
	}

	// Ping to confirm
	pong, err := c.Do("PING")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Redis PING failed: %w", err)
	}
	if pong != "PONG" {
		conn.Close()
		return nil, fmt.Errorf("unexpected PING response: %s", pong)
	}

	return c, nil
}

// Close closes the connection.
func (c *Client) Close() {
	c.conn.Close()
}

// Do sends a Redis command and returns the response as a string.
func (c *Client) Do(args ...string) (string, error) {
	if err := c.writeCommand(args...); err != nil {
		return "", err
	}
	return c.readReply()
}

// Keys returns all keys matching a pattern.
func (c *Client) Keys(pattern string) ([]string, error) {
	if err := c.writeCommand("KEYS", pattern); err != nil {
		return nil, err
	}
	return c.readArray()
}

// DBSize returns the number of keys in the current DB.
func (c *Client) DBSize() (int, error) {
	resp, err := c.Do("DBSIZE")
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(resp)
	if err != nil {
		return 0, fmt.Errorf("unexpected DBSIZE response: %s", resp)
	}
	return n, nil
}

// Info returns the Redis INFO string for a section.
func (c *Client) Info(section string) (string, error) {
	if section == "" {
		return c.Do("INFO")
	}
	return c.Do("INFO", section)
}

// FlushDB flushes the current database.
func (c *Client) FlushDB() error {
	resp, err := c.Do("FLUSHDB", "ASYNC")
	if err != nil {
		// Fallback: some versions don't support ASYNC
		resp, err = c.Do("FLUSHDB")
		if err != nil {
			return err
		}
	}
	if resp != "OK" {
		return fmt.Errorf("FLUSHDB returned: %s", resp)
	}
	return nil
}

// Del deletes specific keys, returns number deleted.
func (c *Client) Del(keys ...string) (int, error) {
	args := append([]string{"DEL"}, keys...)
	resp, err := c.Do(args...)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(resp)
	if err != nil {
		return 0, fmt.Errorf("unexpected DEL response: %s", resp)
	}
	return n, nil
}

// writeCommand writes a RESP array command.
func (c *Client) writeCommand(args ...string) error {
	cmd := fmt.Sprintf("*%d\r\n", len(args))
	for _, arg := range args {
		cmd += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	_, err := c.rw.WriteString(cmd)
	if err != nil {
		return err
	}
	return c.rw.Flush()
}

// readReply reads a single RESP reply and returns it as string.
func (c *Client) readReply() (string, error) {
	line, err := c.rw.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	line = strings.TrimSuffix(line, "\r\n")

	switch line[0] {
	case '+': // Simple string
		return line[1:], nil
	case '-': // Error
		return "", fmt.Errorf("redis error: %s", line[1:])
	case ':': // Integer
		return line[1:], nil
	case '$': // Bulk string
		n, err := strconv.Atoi(line[1:])
		if err != nil {
			return "", fmt.Errorf("invalid bulk length: %s", line)
		}
		if n == -1 {
			return "(nil)", nil
		}
		buf := make([]byte, n+2) // +2 for \r\n
		_, err = c.rw.Read(buf)
		if err != nil {
			return "", err
		}
		return string(buf[:n]), nil
	case '*': // Array — flatten to comma-separated
		count, _ := strconv.Atoi(line[1:])
		parts := make([]string, 0, count)
		for i := 0; i < count; i++ {
			val, err := c.readReply()
			if err != nil {
				return "", err
			}
			parts = append(parts, val)
		}
		return strings.Join(parts, ", "), nil
	default:
		return line, nil
	}
}

// readArray reads a RESP array and returns as []string.
func (c *Client) readArray() ([]string, error) {
	line, err := c.rw.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(line, "\r\n")

	if line[0] == '-' {
		return nil, fmt.Errorf("redis error: %s", line[1:])
	}
	if line[0] != '*' {
		return nil, fmt.Errorf("expected array, got: %s", line)
	}

	count, _ := strconv.Atoi(line[1:])
	if count <= 0 {
		return []string{}, nil
	}

	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		val, err := c.readReply()
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}
