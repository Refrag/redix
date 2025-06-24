package postgresql

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Refrag/redix/internals/datastore/contract"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Engine represents the contract.Engine implementation.
type Engine struct {
	conn *pgxpool.Pool
}

// Open opens the database.
func (e *Engine) Open(dsn string) (err error) {
	e.conn, err = pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return err
	}

	if _, err := e.conn.Exec(
		context.Background(),
		fmt.Sprintf(
			`
				%s
				%s
				%s
				%s
				%s
			`,
			createExtensionQuery,
			createTableQuery,
			createUniqueIndexQuery,
			createTrgmIndexQuery,
			createExpiresAtIndexQuery,
		),
	); err != nil {
		return err
	}

	go (func() {
		for {
			now := time.Now().UnixNano()

			if _, err := e.conn.Exec(
				context.Background(),
				deleteExpiredKeysQuery,
				now,
			); err != nil {
				panic(err)
			}

			time.Sleep(time.Second * 1)
		}
	})()

	return nil
}

func (e *Engine) handleDeleteOperations(input *contract.WriteInput) error {
	if input.Key == nil {
		_, err := e.conn.Exec(context.Background(), deleteAllKeysQuery)
		return err
	}

	if input.Value == nil {
		_, err := e.conn.Exec(
			context.Background(),
			deleteMatchingKeysQuery,
			append(input.Key, '%'),
		)
		return err
	}

	return nil
}

func (e *Engine) processValue(input *contract.WriteInput) (interface{}, bool, error) {
	val := interface{}(string(input.Value))
	isNumber := false

	if fval, err := strconv.ParseFloat(string(input.Value), 64); err == nil {
		isNumber = true
		val = fval
	}

	if input.Increment && !isNumber {
		return nil, false, errors.New("the specified value is not a number")
	}

	return val, isNumber, nil
}

func (e *Engine) buildInsertQuery(input *contract.WriteInput) ([]string, bool) {
	var query []string
	appending := false

	if input.OnlyIfNotExists {
		query = []string{insertQuery, onConflictDoNothing}
	} else if input.Increment {
		query = []string{incrementInsertQuery}
		appending = true
	} else if input.Append {
		query = []string{appendInsertQuery}
		appending = true
	} else {
		// Use INSERT with ON CONFLICT DO UPDATE for upsert behavior
		query = []string{insertQuery, "ON CONFLICT (_key) DO UPDATE SET _value = EXCLUDED._value"}
		appending = true
	}

	if appending && !input.KeepTTL {
		query = append(query, expiresAtUpdateQuery)
	}

	query = append(query, returningQuery)
	return query, appending
}

// Write writes into the database.
func (e *Engine) Write(input *contract.WriteInput) (*contract.WriteOutput, error) {
	if input == nil {
		return nil, errors.New("empty input specified")
	}

	if err := e.handleDeleteOperations(input); err != nil {
		return nil, err
	}

	if input.Key == nil || input.Value == nil {
		return nil, nil
	}

	val, _, err := e.processValue(input)
	if err != nil {
		return nil, err
	}

	insertQuery, _ := e.buildInsertQuery(input)

	ttl := int64(0)
	if input.TTL > 0 {
		ttl = time.Now().Add(input.TTL).UnixNano()
	}

	jsonVal, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	var retVal []byte
	var retExpiresAt int64

	if err := e.conn.QueryRow(
		context.Background(),
		strings.Join(insertQuery, " "),
		input.Key, string(jsonVal), ttl,
	).Scan(&retVal, &retExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &contract.WriteOutput{
		Value: retVal,
		TTL:   time.Since(time.Unix(0, retExpiresAt)),
	}, nil
}

// Read reads from the database.
func (e *Engine) Read(input *contract.ReadInput) (*contract.ReadOutput, error) {
	if input == nil {
		return nil, errors.New("empty input specified")
	}

	var retQueryVal []byte
	var retVal interface{}
	var retExpiresAt int64

	if err := e.conn.QueryRow(
		context.Background(),
		selectQuery,
		input.Key,
	).Scan(&retQueryVal, &retExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &contract.ReadOutput{}, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(retQueryVal, &retVal); err != nil {
		return nil, err
	}

	readOutput := contract.ReadOutput{
		Key:    input.Key,
		Value:  []byte(fmt.Sprintf("%v", retVal)),
		TTL:    0,
		Exists: true,
	}

	if retExpiresAt != 0 {
		readOutput.TTL = time.Until(time.Unix(0, retExpiresAt))
	}

	deleter := func() {
		// TODO report any expected error?
		e.conn.Exec(context.Background(), deleteQuery, input.Key)
	}

	if readOutput.TTL < 0 {
		go (func() {
			deleter()
		})()
		return &contract.ReadOutput{}, nil
	}

	if input.Delete {
		go (func() {
			deleter()
		})()
	}

	return &readOutput, nil
}

// Iterate iterates on the whole database stops if the IteratorOpts returns an error.
func (e *Engine) Iterate(opts *contract.IteratorOpts) error {
	if opts == nil {
		return errors.New("empty options specified")
	}

	if opts.Callback == nil {
		return errors.New("you must specify the callback")
	}

	iter, err := e.conn.Query(
		context.Background(),
		selectWhereQuery,
		append(opts.Prefix, '%'),
	)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.Next() {
		var key, value []byte
		var expiresAt int64

		if err := iter.Scan(&key, &value, &expiresAt); err != nil {
			return err
		}

		var parsedVal interface{}

		if err := json.Unmarshal(value, &parsedVal); err != nil {
			return err
		}

		readOutput := contract.ReadOutput{
			Key:   key,
			Value: []byte(fmt.Sprintf("%v", parsedVal)),
			TTL:   0,
		}

		if expiresAt != 0 {
			readOutput.TTL = time.Until(time.Unix(0, expiresAt))
		}

		// expired
		if readOutput.TTL < 0 {
			continue
		}

		if err := opts.Callback(&readOutput); err != nil {
			return err
		}
	}

	return iter.Err()
}

// Close closes the connection.
func (e *Engine) Close() error {
	e.conn.Close()
	return nil
}

// Publish submits the payload to the specified channel.
func (e *Engine) Publish(channel []byte, payload []byte) error {
	channelEncoded := fmt.Sprintf("%x", md5.Sum(channel))
	if _, err := e.conn.Exec(context.Background(), selectNotifyQuery, channelEncoded, payload); err != nil {
		return err
	}

	return nil
}

// Subscribe listens for the incoming payloads on the specified channel.
func (e *Engine) Subscribe(channel []byte, cb func([]byte) error) error {
	if cb == nil {
		return errors.New("you must specify a callback (cb)")
	}

	conn, err := e.conn.Acquire(context.Background())
	if err != nil {
		return err
	}

	channelEncoded := fmt.Sprintf("\"%x\"", md5.Sum(channel))
	if _, err := conn.Exec(context.Background(), fmt.Sprintf(listenQuery, channelEncoded)); err != nil {
		return fmt.Errorf("database::listen::err %s", err.Error())
	}

	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			return fmt.Errorf("database::notification::err %s", err.Error())
		}

		if err := cb([]byte(notification.Payload)); err != nil {
			return fmt.Errorf("unable to process notification due to: %s", err.Error())
		}
	}
}
