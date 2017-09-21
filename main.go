package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	fb "github.com/huandu/facebook"

	"github.com/go-redis/redis"
)

const defaultPort = "80"
const defaultHost = ""

// HandlerContext comment
type HandlerContext struct {
	redisClient *redis.Client
	fbKey       string
}

func main() {

	// if the server should serve https
	// will be set to false if either the
	// tls key or tls cert files are not found
	shouldUseTLS := true

	// Get the environment variables needed
	// to run
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	tlsKey := os.Getenv("TLSKEY")
	tlsCert := os.Getenv("TLSCERT")
	redisAddr := os.Getenv("REDISADDR")
	fbKey := os.Getenv("FBKEY")

	// if redisAddr is not defined, exit
	if len(redisAddr) == 0 {
		log.Fatal("env REDISADDR not found, exiting...")
	}
	if len(fbKey) == 0 {
		log.Fatal("FBKEY not found, exiting...")
	}

	// get the tls cert and key, if not found do not use TLS
	if len(tlsKey) == 0 {
		fmt.Println("env TLSKEY not found, server will use not https...")
		shouldUseTLS = false
	}
	if len(tlsCert) == 0 {
		fmt.Println("env TLSCERT not found, server will use not https...")
		shouldUseTLS = false
	}

	// Fallback to defaults if the environment variables aren't set
	if len(host) == 0 {
		fmt.Println("ENV HOST not found, defaulting to " + defaultHost)
		host = defaultHost
	}
	if len(port) == 0 {
		fmt.Println("ENV PORT not found, defaulting to " + defaultPort)
		port = defaultPort
	}

	addr := host + ":" + port

	// create the redis client and store it in the handler context
	ropts := redis.Options{
		Addr: redisAddr,
	}
	rclient := redis.NewClient(&ropts)
	hctx := &HandlerContext{
		redisClient: rclient,
		fbKey:       fbKey,
	}

	http.HandleFunc("/get", hctx.FeedHandler)

	if shouldUseTLS {
		fmt.Printf("listening on https://%s...\n", addr)
		log.Fatal(http.ListenAndServeTLS(addr, tlsCert, tlsKey, nil))
	} else {
		fmt.Printf("listening on %s...\n", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

// FeedHandler makes a request to Facebook's Graph API
// to get get event information from the Informatcs Undergraduate
// Association's Facebook Group
func (ctx *HandlerContext) FeedHandler(w http.ResponseWriter, r *http.Request) {
	eventData := "eventdata"

	// Get the existing eventdata from redis if it exists.
	// Handle the error if it exists and it is noot redis.nil.
	data, err := ctx.redisClient.Get(eventData).Bytes()
	if err != nil && err != redis.Nil {
		http.Error(w, "error getting from cache: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}
	// If Redis contains no eventData, query the
	// Facebook API for data
	if err == redis.Nil {
		fmt.Println("querying the FB api...")
		result, err = fb.Get("/232675096843082/events", fb.Params{
			"fields":       "name,description,start_time,id,cover,place,end_time,is_canceled",
			"access_token": ctx.fbKey,
		})
		// Handle an error that is returned by the
		if err != nil {
			http.Error(w, "error fetching event data from facebook: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Filter the results down to just event data.
		result := result["data"]
		data, err = json.Marshal(result)
		if err != nil {
			http.Error(w, "error marshalling json: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Store the retrieved information
		ctx.redisClient.Set(eventData, data, time.Hour*1)
	}

	// Add headers
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
