package timeseries

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Clickhouse struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string

	connection driver.Conn
}

func New(host string, port string, database string, username string, password string) (*Clickhouse, error) {
	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	c := Clickhouse{
		Host: host,
		Port: parsedPort,
		Database: database,
		Username: username,
		Password: password,
	}
	return &c, nil
}

func (c *Clickhouse) Connect() error {
	if c.connection != nil {
		return nil
	}

	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
            Addr: []string{fmt.Sprintf("%s:%d", c.Host, c.Port)},
            Auth: clickhouse.Auth{
                Database: c.Database,
                Username: c.Username,
                Password: c.Password,
            },
            ClientInfo: clickhouse.ClientInfo{
                Products: []struct {
                    Name    string
                    Version string
                }{
                    {Name: "smarthome-client", Version: "0.1"},
                },
            },
        })

	if err != nil {
		return err
	}

	if err := conn.Ping(ctx); err != nil {
        if exception, ok := err.(*clickhouse.Exception); ok {
            fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
        }
        return err
    }

	c.connection = conn
	return nil
}

func (c *Clickhouse) Close() error {
	if c.connection != nil {
		err := c.connection.Close()
		return err
	}
	return nil
}

func Query[T any](ctx context.Context, c *Clickhouse, query string, args ...any) ([]T, error) {
	if c.connection == nil {
		return nil, errors.New("no connection")
	}

	rows, err := c.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var item T
		if err := rows.ScanStruct(&item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, rows.Err()
}
