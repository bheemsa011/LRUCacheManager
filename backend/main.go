package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type node struct {
	key        string
	value      interface{}
	prev, next *node
	expiresAt  time.Time
}

type linkedList struct {
	head, tail *node
	size       int
}

func (l *linkedList) add(key string, value interface{}, expiresAt time.Time) {
	newNode := &node{key: key, value: value, expiresAt: expiresAt}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		newNode.next = l.head
		l.head.prev = newNode
		l.head = newNode
	}
	l.size++
}

func (l *linkedList) remove(node *node) {
	if node == nil {
		return
	}
	if node.prev == nil {
		l.head = node.next
	} else {
		node.prev.next = node.next
	}
	if node.next == nil {
		l.tail = node.prev
	} else {
		node.next.prev = node.prev
	}
	l.size--
}

func (l *linkedList) get(key string) (interface{}, bool) {
	for current := l.head; current != nil; current = current.next {
		if current.key == key {
			if time.Now().After(current.expiresAt) {
				l.remove(current)
				return nil, false
			}
			return current.value, true
		}
	}
	return nil, false
}

func (l *linkedList) cleanExpired() {
	for current := l.head; current != nil; {
		next := current.next
		if time.Now().After(current.expiresAt) {
			l.remove(current)
		}
		current = next
	}
}

var lruCache linkedList

func init() {
	lruCache.size = 0
}

func main() {
	router := mux.NewRouter()

	// CORS configuration
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"}),
		handlers.ExposedHeaders([]string{"Content-Length"}),
	//	handlers.AllowCredentials(true),
	//handlers.MaxAge(12*time.Hour),
	)

	// Define API endpoints using Gorilla Mux
	router.HandleFunc("/get", getKeyHandler).Methods("GET")
	router.HandleFunc("/get-all", getAllCacheHandler).Methods("GET")
	router.HandleFunc("/set", setKeyHandler).Methods("POST")

	// Apply CORS middleware
	handler := cors(router)

	// Start the server
	http.ListenAndServe(":8086", handler)

}
func getKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key") // Get the "key" query parameter
	value, found := lruCache.get(key)

	if found {
		w.WriteHeader(http.StatusOK)
		// Encode value as JSON and write to response body
		if err := json.NewEncoder(w).Encode(value); err != nil {
			fmt.Fprintf(w, "Error encoding response: %v", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Key not found: %s", key)
	}
}

func setKeyHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Duration string `json:"duration"`
	}

	// Decode request body as JSON and bind to request struct
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request payload: %v", err)
		return
	}

	var duration time.Duration
	if request.Duration != "" {
		duration, _ = time.ParseDuration(request.Duration)
	} else {
		duration = 60 * time.Second // Default expiration
	}

	if lruCache.size >= 3 {
		lruCache.remove(lruCache.tail) // Remove least recently used when full
	}

	lruCache.add(request.Key, request.Value, time.Now().Add(duration))
	lruCache.cleanExpired()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Set key: %s, value: %s", request.Key, request.Value)
}

func getAllCacheHandler(w http.ResponseWriter, r *http.Request) {
	lruCache.cleanExpired() // Clean expired entries first

	caches := []map[string]interface{}{}
	current := lruCache.head
	for current != nil {
		caches = append(caches, map[string]interface{}{
			"key":   current.key,
			"value": current.value,
		})
		current = current.next
	}

	// Encode cached items as JSON and write to response body
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(caches); err != nil {
		fmt.Fprintf(w, "Error encoding response: %v", err)
		return
	}
}
