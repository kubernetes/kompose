// Package simpleredis provides an easy way to use Redis.
package simpleredis

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	// Version number. Stable API within major version numbers.
	Version = 2.6

	// The default [url]:port that Redis is running at
	defaultRedisServer = ":6379"
)

// Common for each of the Redis data structures used here
type redisDatastructure struct {
	pool    *ConnectionPool
	id      string
	dbindex int
}

type (
	// A pool of readily available Redis connections
	ConnectionPool redis.Pool

	List     redisDatastructure
	Set      redisDatastructure
	HashMap  redisDatastructure
	KeyValue redisDatastructure
)

var (
	// Timeout settings for new connections
	connectTimeout = 7 * time.Second
	readTimeout    = 7 * time.Second
	writeTimeout   = 7 * time.Second
	idleTimeout    = 240 * time.Second

	// How many connections should stay ready for requests, at a maximum?
	// When an idle connection is used, new idle connections are created.
	maxIdleConnections = 3
)

/* --- Helper functions --- */

// Connect to the local instance of Redis at port 6379
func newRedisConnection() (redis.Conn, error) {
	return newRedisConnectionTo(defaultRedisServer)
}

// Connect to host:port, host may be omitted, so ":6379" is valid.
// Will not try to AUTH with any given password (password@host:port).
func newRedisConnectionTo(hostColonPort string) (redis.Conn, error) {
	// Discard the password, if provided
	if _, theRest, ok := twoFields(hostColonPort, "@"); ok {
		hostColonPort = theRest
	}
	hostColonPort = strings.TrimSpace(hostColonPort)
	c, err := redis.Dial("tcp", hostColonPort, redis.DialConnectTimeout(connectTimeout), redis.DialReadTimeout(readTimeout), redis.DialWriteTimeout(writeTimeout))
	if err != nil {
		if c != nil {
			c.Close()
		}
		return nil, err
	}
	return c, nil
}

// Get a string from a list of results at a given position
func getString(bi []interface{}, i int) string {
	return string(bi[i].([]uint8))
}

// Test if the local Redis server is up and running
func TestConnection() (err error) {
	return TestConnectionHost(defaultRedisServer)
}

// Test if a given Redis server at host:port is up and running.
// Does not try to PING or AUTH.
func TestConnectionHost(hostColonPort string) (err error) {
	// Connect to the given host:port
	conn, err := newRedisConnectionTo(hostColonPort)
	defer func() {
		if conn != nil {
			conn.Close()
		}
		if r := recover(); r != nil {
			err = errors.New("Could not connect to redis server: " + hostColonPort)
		}
	}()
	return err
}

/* --- ConnectionPool functions --- */

func copyPoolValues(src *redis.Pool) ConnectionPool {
	return ConnectionPool{
		Dial:            src.Dial,
		DialContext:     src.DialContext,
		TestOnBorrow:    src.TestOnBorrow,
		MaxIdle:         src.MaxIdle,
		MaxActive:       src.MaxActive,
		IdleTimeout:     src.IdleTimeout,
		Wait:            src.Wait,
		MaxConnLifetime: src.MaxConnLifetime,
	}
}

// Create a new connection pool
func NewConnectionPool() *ConnectionPool {
	// The second argument is the maximum number of idle connections
	redisPool := &redis.Pool{
		MaxIdle:     maxIdleConnections,
		IdleTimeout: idleTimeout,
		Dial:        newRedisConnection,
	}

	pool := copyPoolValues(redisPool)
	return &pool
}

// Split a string into two parts, given a delimiter.
// Returns the two parts and true if it works out.
func twoFields(s, delim string) (string, string, bool) {
	if strings.Count(s, delim) != 1 {
		return s, "", false
	}
	fields := strings.Split(s, delim)
	return fields[0], fields[1], true
}

// Create a new connection pool given a host:port string.
// A password may be supplied as well, on the form "password@host:port".
func NewConnectionPoolHost(hostColonPort string) *ConnectionPool {
	// Create a redis Pool
	redisPool := &redis.Pool{
		// Maximum number of idle connections to the redis database
		MaxIdle:     maxIdleConnections,
		IdleTimeout: idleTimeout,
		// Anonymous function for calling new RedisConnectionTo with the host:port
		Dial: func() (redis.Conn, error) {
			conn, err := newRedisConnectionTo(hostColonPort)
			if err != nil {
				return nil, err
			}
			// If a password is given, use it to authenticate
			if password, _, ok := twoFields(hostColonPort, "@"); ok {
				if password != "" {
					if _, err := conn.Do("AUTH", password); err != nil {
						conn.Close()
						return nil, err
					}
				}
			}
			return conn, err
		},
	}
	pool := copyPoolValues(redisPool)
	return &pool
}

// Set the number of maximum *idle* connections standing ready when
// creating new connection pools. When an idle connection is used,
// a new idle connection is created. The default is 3 and should be fine
// for most cases.
func SetMaxIdleConnections(maximum int) {
	maxIdleConnections = maximum
}

// Get one of the available connections from the connection pool, given a database index
func (pool *ConnectionPool) Get(dbindex int) redis.Conn {
	redisPool := (*redis.Pool)(pool)
	conn := redisPool.Get()
	// The default database index is 0
	if dbindex != 0 {
		// SELECT is not critical, ignore the return values
		conn.Do("SELECT", strconv.Itoa(dbindex))
	}
	return conn
}

// Ping the server by sending a PING command
func (pool *ConnectionPool) Ping() error {
	redisPool := (*redis.Pool)(pool)
	conn := redisPool.Get()
	_, err := conn.Do("PING")
	return err
}

// Close down the connection pool
func (pool *ConnectionPool) Close() {
	redisPool := (*redis.Pool)(pool)
	redisPool.Close()
}

/* --- List functions --- */

// Create a new list
func NewList(pool *ConnectionPool, id string) *List {
	return &List{pool, id, 0}
}

// Select a different database
func (rl *List) SelectDatabase(dbindex int) {
	rl.dbindex = dbindex
}

// Returns the element at index index in the list
func (rl *List) Get(index int64) (string, error) {
	conn := rl.pool.Get(rl.dbindex)
	result, err := conn.Do("LINDEX", rl.id, index)
	if err != nil {
		panic(err)
	}
	return redis.String(result, err)
}

// Get the size of the list
func (rl *List) Size() (int64, error) {
	conn := rl.pool.Get(rl.dbindex)
	size, err := conn.Do("LLEN", rl.id)
	if err != nil {
		panic(err)
	}
	return redis.Int64(size, err)
}

// Removes and returns the first element of the list
func (rl *List) PopFirst() (string, error) {
	conn := rl.pool.Get(rl.dbindex)
	result, err := conn.Do("LPOP", rl.id)
	if err != nil {
		panic(err)
	}
	return redis.String(result, err)
}

// Removes and returns the last element of the list
func (rl *List) PopLast() (string, error) {
	conn := rl.pool.Get(rl.dbindex)
	result, err := conn.Do("LPOP", rl.id)
	if err != nil {
		panic(err)
	}
	return redis.String(result, err)
}

// Add an element to the start of the list
func (rl *List) AddStart(value string) error {
	conn := rl.pool.Get(rl.dbindex)
	_, err := conn.Do("RPUSH", rl.id, value)
	return err
}

// Add an element to the end of the list list
func (rl *List) AddEnd(value string) error {
	conn := rl.pool.Get(rl.dbindex)
	_, err := conn.Do("LPUSH", rl.id, value)
	return err
}

// Default Add, aliased to List.AddStart
func (rl *List) Add(value string) error {
	return rl.AddStart(value)
}

// Get all elements of a list
func (rl *List) All() ([]string, error) {
	conn := rl.pool.Get(rl.dbindex)
	result, err := redis.Values(conn.Do("LRANGE", rl.id, "0", "-1"))
	strs := make([]string, len(result))
	for i := 0; i < len(result); i++ {
		strs[i] = getString(result, i)
	}
	return strs, err
}

// Deprecated
func (rl *List) GetAll() ([]string, error) {
	return rl.All()
}

// Get the last element of a list
func (rl *List) Last() (string, error) {
	conn := rl.pool.Get(rl.dbindex)
	result, err := redis.Values(conn.Do("LRANGE", rl.id, "-1", "-1"))
	if len(result) == 1 {
		return getString(result, 0), err
	}
	return "", err
}

// Deprecated
func (rl *List) GetLast() (string, error) {
	return rl.Last()
}

// Get the last N elements of a list
func (rl *List) LastN(n int) ([]string, error) {
	conn := rl.pool.Get(rl.dbindex)
	result, err := redis.Values(conn.Do("LRANGE", rl.id, "-"+strconv.Itoa(n), "-1"))
	strs := make([]string, len(result))
	for i := 0; i < len(result); i++ {
		strs[i] = getString(result, i)
	}
	return strs, err
}

// Deprecated
func (rl *List) GetLastN(n int) ([]string, error) {
	return rl.LastN(n)
}

// Remove the first occurrence of an element from the list
func (rl *List) RemoveElement(value string) error {
	conn := rl.pool.Get(rl.dbindex)
	_, err := conn.Do("LREM", rl.id, value)
	return err
}

// Set element of list at index n to value
func (rl *List) Set(index int64, value string) error {
	conn := rl.pool.Get(rl.dbindex)
	_, err := conn.Do("LSET", rl.id, index, value)
	return err
}

// Trim an existing list so that it will contain only the specified range of
// elements specified.
func (rl *List) Trim(start, stop int64) error {
	conn := rl.pool.Get(rl.dbindex)
	_, err := conn.Do("LTRIM", rl.id, start, stop)
	return err
}

// Remove this list
func (rl *List) Remove() error {
	conn := rl.pool.Get(rl.dbindex)
	_, err := conn.Do("DEL", rl.id)
	return err
}

// Clear the contents
func (rl *List) Clear() error {
	return rl.Remove()
}

/* --- Set functions --- */

// Create a new set
func NewSet(pool *ConnectionPool, id string) *Set {
	return &Set{pool, id, 0}
}

// Select a different database
func (rs *Set) SelectDatabase(dbindex int) {
	rs.dbindex = dbindex
}

// Add an element to the set
func (rs *Set) Add(value string) error {
	conn := rs.pool.Get(rs.dbindex)
	_, err := conn.Do("SADD", rs.id, value)
	return err
}

// Returns the set cardinality (number of elements) of the set
func (rs *Set) Size() (int64, error) {
	conn := rs.pool.Get(rs.dbindex)
	size, err := conn.Do("SCARD", rs.id)
	if err != nil {
		panic(err)
	}
	return redis.Int64(size, err)
}

// Check if a given value is in the set
func (rs *Set) Has(value string) (bool, error) {
	conn := rs.pool.Get(rs.dbindex)
	retval, err := conn.Do("SISMEMBER", rs.id, value)
	if err != nil {
		panic(err)
	}
	return redis.Bool(retval, err)
}

// Get all elements of the set
func (rs *Set) All() ([]string, error) {
	conn := rs.pool.Get(rs.dbindex)
	result, err := redis.Values(conn.Do("SMEMBERS", rs.id))
	strs := make([]string, len(result))
	for i := 0; i < len(result); i++ {
		strs[i] = getString(result, i)
	}
	return strs, err
}

// Deprecated
func (rs *Set) GetAll() ([]string, error) {
	return rs.All()
}

// Remove a random member from the set
func (rs *Set) Pop() (string, error) {
	conn := rs.pool.Get(rs.dbindex)
	result, err := conn.Do("SPOP", rs.id)
	if err != nil {
		panic(err)
	}
	return redis.String(result, err)
}

// Get a random member of the set
func (rs *Set) Random() (string, error) {
	conn := rs.pool.Get(rs.dbindex)
	result, err := conn.Do("SRANDMEMBER", rs.id)
	if err != nil {
		panic(err)
	}
	return redis.String(result, err)
}

// Remove an element from the set
func (rs *Set) Del(value string) error {
	conn := rs.pool.Get(rs.dbindex)
	_, err := conn.Do("SREM", rs.id, value)
	return err
}

// Remove this set
func (rs *Set) Remove() error {
	conn := rs.pool.Get(rs.dbindex)
	_, err := conn.Do("DEL", rs.id)
	return err
}

// Clear the contents
func (rs *Set) Clear() error {
	return rs.Remove()
}

/* --- HashMap functions --- */

// Create a new hashmap
func NewHashMap(pool *ConnectionPool, id string) *HashMap {
	return &HashMap{pool, id, 0}
}

// Select a different database
func (rh *HashMap) SelectDatabase(dbindex int) {
	rh.dbindex = dbindex
}

// Set a value in a hashmap given the element id (for instance a user id) and the key (for instance "password")
func (rh *HashMap) Set(elementid, key, value string) error {
	conn := rh.pool.Get(rh.dbindex)
	_, err := conn.Do("HSET", rh.id+":"+elementid, key, value)
	return err
}

// Given an element id, set a key and a value together with an expiration time
func (rh *HashMap) SetExpire(elementid, key, value string, expire time.Duration) error {
	conn := rh.pool.Get(rh.dbindex)
	if _, err := conn.Do("HSET", rh.id+":"+elementid, key, value); err != nil {
		return err
	}
	// No EXPIRE in Redis for hash keys, as far as I can tell from the documentation.
	// This is the manual way.
	go func() {
		time.Sleep(expire)
		rh.DelKey(elementid, key)
	}()
	// Set the elementid to expire in the given duration (as milliseconds)
	//expireMilliseconds := expire.Nanoseconds() / 1000000
	//if _, err := conn.Do("PEXPIRE", rh.id+":"+elementid, expireMilliseconds); err != nil {
	//	return err
	//}
	return nil
}

// Commented out because this would only return TTL for the elementid, not for the key
// TimeToLive returns how long a key has to live until it expires
// Returns a duration of 0 when the time has passed
//func (rh *HashMap) TimeToLive(elementid string) (time.Duration, error) {
//	conn := rh.pool.Get(rh.dbindex)
//	ttlSecondsInterface, err := conn.Do("TTL", rh.id+":"+elementid)
//	if err != nil || ttlSecondsInterface.(int64) <= 0 {
//		return time.Duration(0), err
//	}
//	ns := time.Duration(ttlSecondsInterface.(int64)) * time.Second
//	return ns, nil
//}

// Get a value from a hashmap given the element id (for instance a user id) and the key (for instance "password")
func (rh *HashMap) Get(elementid, key string) (string, error) {
	conn := rh.pool.Get(rh.dbindex)
	result, err := redis.String(conn.Do("HGET", rh.id+":"+elementid, key))
	if err != nil {
		return "", err
	}
	return result, nil
}

// Check if a given elementid + key is in the hash map
func (rh *HashMap) Has(elementid, key string) (bool, error) {
	conn := rh.pool.Get(rh.dbindex)
	retval, err := conn.Do("HEXISTS", rh.id+":"+elementid, key)
	if err != nil {
		panic(err)
	}
	return redis.Bool(retval, err)
}

// Keys returns the keys of the given elementid.
func (rh *HashMap) Keys(elementid string) ([]string, error) {
	conn := rh.pool.Get(rh.dbindex)
	result, err := redis.Values(conn.Do("HKEYS", rh.id+":"+elementid))
	strs := make([]string, len(result))
	for i := 0; i < len(result); i++ {
		strs[i] = getString(result, i)
	}
	return strs, err
}

// Check if a given elementid exists as a hash map at all
func (rh *HashMap) Exists(elementid string) (bool, error) {
	// TODO: key is not meant to be a wildcard, check for "*"
	return hasKey(rh.pool, rh.id+":"+elementid, rh.dbindex)
}

// Get all elementid's for all hash elements
func (rh *HashMap) All() ([]string, error) {
	conn := rh.pool.Get(rh.dbindex)
	result, err := redis.Values(conn.Do("KEYS", rh.id+":*"))
	strs := make([]string, len(result))
	idlen := len(rh.id)
	for i := 0; i < len(result); i++ {
		strs[i] = getString(result, i)[idlen+1:]
	}
	return strs, err
}

// Deprecated
func (rh *HashMap) GetAll() ([]string, error) {
	return rh.All()
}

// Remove a key for an entry in a hashmap (for instance the email field for a user)
func (rh *HashMap) DelKey(elementid, key string) error {
	conn := rh.pool.Get(rh.dbindex)
	_, err := conn.Do("HDEL", rh.id+":"+elementid, key)
	return err
}

// Remove an element (for instance a user)
func (rh *HashMap) Del(elementid string) error {
	conn := rh.pool.Get(rh.dbindex)
	_, err := conn.Do("DEL", rh.id+":"+elementid)
	return err
}

// Remove this hashmap (all keys that starts with this hashmap id and a colon)
func (rh *HashMap) Remove() error {
	conn := rh.pool.Get(rh.dbindex)
	// Find all hashmap keys that starts with rh.id+":"
	results, err := redis.Values(conn.Do("KEYS", rh.id+":*"))
	if err != nil {
		return err
	}
	// For each key id
	for i := 0; i < len(results); i++ {
		// Delete this key
		if _, err = conn.Do("DEL", getString(results, i)); err != nil {
			return err
		}
	}
	return nil
}

// Clear the contents
func (rh *HashMap) Clear() error {
	return rh.Remove()
}

/* --- KeyValue functions --- */

// Create a new key/value
func NewKeyValue(pool *ConnectionPool, id string) *KeyValue {
	return &KeyValue{pool, id, 0}
}

// Select a different database
func (rkv *KeyValue) SelectDatabase(dbindex int) {
	rkv.dbindex = dbindex
}

// Set a key and value
func (rkv *KeyValue) Set(key, value string) error {
	conn := rkv.pool.Get(rkv.dbindex)
	_, err := conn.Do("SET", rkv.id+":"+key, value)
	return err
}

// Set a key and value, with expiry
func (rkv *KeyValue) SetExpire(key, value string, expire time.Duration) error {
	conn := rkv.pool.Get(rkv.dbindex)
	// Convert from nanoseconds to milliseconds
	expireMilliseconds := expire.Nanoseconds() / 1000000
	// Set the value, together with an expiry time, given in milliseconds
	_, err := conn.Do("SET", rkv.id+":"+key, value, "PX", expireMilliseconds)
	return err
}

// TimeToLive returns how long a key has to live until it expires
// Returns a duration of 0 when the time has passed
func (rkv *KeyValue) TimeToLive(key string) (time.Duration, error) {
	conn := rkv.pool.Get(rkv.dbindex)
	ttlSecondsInterface, err := conn.Do("TTL", rkv.id+":"+key)
	if err != nil || ttlSecondsInterface.(int64) <= 0 {
		return time.Duration(0), err
	}
	ns := time.Duration(ttlSecondsInterface.(int64)) * time.Second
	return ns, nil
}

// Get a value given a key
func (rkv *KeyValue) Get(key string) (string, error) {
	conn := rkv.pool.Get(rkv.dbindex)
	result, err := redis.String(conn.Do("GET", rkv.id+":"+key))
	if err != nil {
		return "", err
	}
	return result, nil
}

// Remove a key
func (rkv *KeyValue) Del(key string) error {
	conn := rkv.pool.Get(rkv.dbindex)
	_, err := conn.Do("DEL", rkv.id+":"+key)
	return err
}

// Increase the value of a key, returns the new value
// Returns an empty string if there were errors,
// or "0" if the key does not already exist.
func (rkv *KeyValue) Inc(key string) (string, error) {
	conn := rkv.pool.Get(rkv.dbindex)
	result, err := redis.Int64(conn.Do("INCR", rkv.id+":"+key))
	if err != nil {
		return "0", err
	}
	return strconv.FormatInt(result, 10), nil
}

// Remove this key/value
func (rkv *KeyValue) Remove() error {
	conn := rkv.pool.Get(rkv.dbindex)
	// Find all keys that starts with rkv.id+":"
	results, err := redis.Values(conn.Do("KEYS", rkv.id+":*"))
	if err != nil {
		return err
	}
	// For each key id
	for i := 0; i < len(results); i++ {
		// Delete this key
		if _, err = conn.Do("DEL", getString(results, i)); err != nil {
			return err
		}
	}
	return nil
}

// Clear the contents
func (rkv *KeyValue) Clear() error {
	return rkv.Remove()
}

// --- Generic redis functions ---

// Check if a key exists. The key can be a wildcard (ie. "user*").
func hasKey(pool *ConnectionPool, wildcard string, dbindex int) (bool, error) {
	conn := pool.Get(dbindex)
	result, err := redis.Values(conn.Do("KEYS", wildcard))
	if err != nil {
		return false, err
	}
	return len(result) > 0, nil
}

// --- Related to setting and retrieving timeout values

// SetConnectTimeout sets the connect timeout for new connections
func SetConnectTimeout(t time.Duration) {
	connectTimeout = t
}

// SetReadTimeout sets the read timeout for new connections
func SetReadTimeout(t time.Duration) {
	readTimeout = t
}

// SetWriteTimeout sets the write timeout for new connections
func SetWriteTimeout(t time.Duration) {
	writeTimeout = t
}

// SetIdleTimeout sets the idle timeout for new connections
func SetIdleTimeout(t time.Duration) {
	idleTimeout = t
}

// ConnectTimeout returns the current connect timeout for new connections
func ConnectTimeout() time.Duration {
	return connectTimeout
}

// ReadTimeout returns the current read timeout for new connections
func ReadTimeout() time.Duration {
	return readTimeout
}

// WriteTimeout returns the current write timeout for new connections
func WriteTimeout() time.Duration {
	return writeTimeout
}

// IdleTimeout returns the current idle timeout for new connections
func IdleTimeout() time.Duration {
	return idleTimeout
}
