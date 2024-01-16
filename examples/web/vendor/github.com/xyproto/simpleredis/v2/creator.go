package simpleredis

import (
	"github.com/xyproto/pinterface"
)

// For implementing pinterface.ICreator

type RedisCreator struct {
	pool    *ConnectionPool
	dbindex int
}

func NewCreator(pool *ConnectionPool, dbindex int) *RedisCreator {
	return &RedisCreator{pool, dbindex}
}

func (c *RedisCreator) SelectDatabase(dbindex int) {
	c.dbindex = dbindex
}

func (c *RedisCreator) NewList(id string) (pinterface.IList, error) {
	return &List{c.pool, id, c.dbindex}, nil
}

func (c *RedisCreator) NewSet(id string) (pinterface.ISet, error) {
	return &Set{c.pool, id, c.dbindex}, nil
}

func (c *RedisCreator) NewHashMap(id string) (pinterface.IHashMap, error) {
	return &HashMap{c.pool, id, c.dbindex}, nil
}

func (c *RedisCreator) NewKeyValue(id string) (pinterface.IKeyValue, error) {
	return &KeyValue{c.pool, id, c.dbindex}, nil
}
