package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/xyproto/simpleredis/v2"
)

var (
	leaderPool   *simpleredis.ConnectionPool
	replicaPools []*simpleredis.ConnectionPool
)

func ListRangeHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	var members []string
	var err error
	for _, replicaPool := range replicaPools {
		list := simpleredis.NewList(replicaPool, key)
		members, err = list.GetAll()
		if err == nil {
			break // Found a replica with data, exit loop
		}
	}
	if err != nil {
		http.Error(rw, "Failed to retrieve data from replicas", http.StatusInternalServerError)
		return
	}

	membersJSON, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		http.Error(rw, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	rw.Write(membersJSON)
}

func ListPushHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	value := mux.Vars(req)["value"]
	list := simpleredis.NewList(leaderPool, key)
	HandleError(nil, list.Add(value))
	ListRangeHandler(rw, req)
}

func InfoHandler(rw http.ResponseWriter, req *http.Request) {
	info := HandleError(leaderPool.Get(0).Do("INFO")).([]byte)
	rw.Write(info)
}

func EnvHandler(rw http.ResponseWriter, req *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		environment[key] = val
	}

	envJSON := HandleError(json.MarshalIndent(environment, "", "  ")).([]byte)
	rw.Write(envJSON)
}

func HandleError(result interface{}, err error) (r interface{}) {
	if err != nil {
		panic(err)
	}
	return result
}

func getReplicaPool() *simpleredis.ConnectionPool {
	// Use the first replica as the primary replica for read operations
	return replicaPools[0]
}

func main() {
	// Read the Redis replica addresses from an environment variable
	replicaAddresses := os.Getenv("REDIS_REPLICAS")
	if replicaAddresses == "" {
		// Use default values if not set
		replicaAddresses = "redis-replica"
	}

	// Read the Redis port number from an environment variable
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		// Use default port if not set
		redisPort = "6379"
	}

	// Read the Redis leader address from an environment variable
	redisLeaderAddress := os.Getenv("REDIS_LEADER")
	if redisLeaderAddress == "" {
		// Use default leader address if not set
		redisLeaderAddress = "redis-leader"
	}

	// Read the server port number from an environment variable
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		// Use default port if not set
		serverPort = "8080"
	}

	// Construct the Redis leader and replica addresses
	leaderAddress := redisLeaderAddress + ":" + redisPort
	replicaAddressesArr := strings.Split(replicaAddresses, ",")
	replicaPools = make([]*simpleredis.ConnectionPool, len(replicaAddressesArr))
	for i, addr := range replicaAddressesArr {
		replicaPools[i] = simpleredis.NewConnectionPoolHost(addr + ":" + redisPort)
		defer replicaPools[i].Close()
	}

	// Create a connection pool for the leader using the constructed address
	leaderPool = simpleredis.NewConnectionPoolHost(leaderAddress)
	defer leaderPool.Close()

	r := mux.NewRouter()
	r.Path("/lrange/{key}").Methods("GET").HandlerFunc(ListRangeHandler)
	r.Path("/rpush/{key}/{value}").Methods("GET").HandlerFunc(ListPushHandler)
	r.Path("/info").Methods("GET").HandlerFunc(InfoHandler)
	r.Path("/env").Methods("GET").HandlerFunc(EnvHandler)

	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":" + serverPort)
}
