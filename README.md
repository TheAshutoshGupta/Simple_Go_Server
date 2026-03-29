# Go for Backend Developers

A focused guide for developers coming from JavaScript, TypeScript, or C++.
Skips what doesn't matter for backend. Covers what does.

---

## 1. Variables

```go
var foo int = 42   // explicit type
var foo = 42       // type inferred
foo := 42          // shorthand — only inside functions
```

> `:=` only works inside functions. At package level always use `var`.

---

## 2. Functions

```go
// basic function
func greet(name string) string {
    return "Hello " + name
}

// multiple return values — how Go handles errors
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }
    return a / b, nil
}

result, err := divide(10, 2)
if err != nil {
    fmt.Println("error:", err)
    return
}
```

Go has no try/catch. Every function that can fail returns an `error` as the last value. Always check it.

---

## 3. Structs & Methods

Go has no classes. Structs are the building block of everything.

```go
type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
}

// method — (u User) is the receiver, like "this" in JS
func (u User) Greet() string {
    return "Hello " + u.Name
}

u := User{Name: "John", Age: 25}
u.Greet() // Hello John
```

Backtick tags like `` `json:"name"` `` control how the struct serializes to JSON. Without them, keys are capitalized.

---

## 4. Pointers

```go
// without pointer — struct is copied, original unchanged
func (u User) SetName(name string) {
    u.Name = name // changes the copy, not the original
}

// with pointer — modifies the original
func (u *User) SetName(name string) {
    u.Name = name // modifies the original
}

u := User{Name: "John"}
u.SetName("Jane")
fmt.Println(u.Name) // Jane ✅
```

Use pointer receivers `*User` when your method needs to mutate the struct.

---

## 5. Interfaces

Go interfaces are implicit — you don't declare that a type implements an interface. If it has the methods, it qualifies.

```go
type UserStore interface {
    GetUser(id string) (User, error)
    CreateUser(user User) error
}

// PostgresStore automatically satisfies UserStore
// because it has both methods
type PostgresStore struct{ db *sql.DB }

func (p PostgresStore) GetUser(id string) (User, error) { ... }
func (p PostgresStore) CreateUser(user User) error { ... }

// MockStore also satisfies UserStore — for testing
type MockStore struct{}

func (m MockStore) GetUser(id string) (User, error) {
    return User{Name: "test"}, nil
}
func (m MockStore) CreateUser(user User) error { return nil }
```

This is how you write testable Go code — your handler depends on the interface, not the concrete type.

---

## 6. Error Handling

```go
// errors.New for simple errors
return 0, errors.New("something went wrong")

// fmt.Errorf for errors with context
return 0, fmt.Errorf("getUser: id %s not found", id)

// always handle errors immediately
user, err := getUser("123")
if err != nil {
    http.Error(w, "user not found", http.StatusNotFound)
    return // always return after error
}
```

> Never ignore errors with `_` in production code. Always return after handling an error or code keeps running below it.

---

## 7. HTTP Server

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func getUser(w http.ResponseWriter, r *http.Request) {
    // only allow GET
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // read query param — GET /user?id=123
    id := r.URL.Query().Get("id")

    user := User{Name: "John", Age: 25}

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(user)
}

func main() {
    http.HandleFunc("/user", getUser)
    fmt.Println("running on :8080")
    http.ListenAndServe(":8080", nil)
}
```

`w` is where you write your response. `r` contains everything about the incoming request.

---

## 8. Reading Request Body (POST)

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    fmt.Println(user.Name)
}
```

---

## 9. Middleware

A function that wraps a handler and runs before it. Used for logging, auth, rate limiting.

```go
func logger(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("incoming:", r.Method, r.URL.Path)
        next(w, r) // call the actual handler
    }
}

func authCheck(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next(w, r)
    }
}

// wrap your handler
http.HandleFunc("/user", logger(authCheck(getUser)))
```

Request flows like this: `logger → authCheck → getUser → response`

---

## 10. Sending Files

```go
// simplest — Go handles everything
func sendFile(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "files/document.pdf")
}

// force download instead of opening in browser
func downloadFile(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Disposition", "attachment; filename=document.pdf")
    w.Header().Set("Content-Type", "application/pdf")
    http.ServeFile(w, r, "files/document.pdf")
}
```

---

## 11. Goroutines

A goroutine is a function that runs concurrently. Launched with the `go` keyword.

```go
go myFunction() // runs concurrently
```

**The main() problem** — main does not wait for goroutines. If main exits, everything dies.

```go
var wg sync.WaitGroup

func main() {
    for i := 0; i < 100; i++ {
        wg.Add(1)      // one more goroutine starting
        go process(i)
    }
    wg.Wait()          // block until all goroutines finish
}

func process(i int) {
    defer wg.Done()    // runs when this goroutine finishes
    fmt.Println(i)
}
```

Three rules of WaitGroup:
- `Add(1)` before launching a goroutine
- `Done()` at the end — always use `defer`
- `Wait()` where you want to block until all finish

Output order is always random. That is correct behavior, not a bug.

---

## 12. Channels

Channels let goroutines send data to each other.

```go
ch := make(chan int)     // unbuffered
ch := make(chan int, 10) // buffered — holds 10 values without blocking

ch <- 42     // send
v := <-ch    // receive
```

**Unbuffered** — both sender and receiver must be ready at the same time. Like a live call.

**Buffered** — sender can keep going until the buffer is full. Like a voicemail.

```go
func main() {
    ch := make(chan int, 100)
    var wg sync.WaitGroup

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            ch <- n  // send into channel
        }(i)
    }

    wg.Wait()
    close(ch) // always close when done sending

    for v := range ch {
        fmt.Println(v) // receive all values
    }
}
```

> For basic backend work you can mostly skip channels. You will need them for background workers, rate limiters, and job queues.

---

## 13. Rate Limiting — Token Bucket Algorithm

This is a real-world rate limiter using the **token bucket algorithm**. Instead of counting requests, it refills tokens over time — so a user who waits gets their quota back naturally.

```go
package main

import (
    "fmt"
    "net/http"
    "time"
)

type Bucket struct {
    tokens         float64
    lastRefillTime time.Time
}

const AllowedTokenPerSecond = 10 // max requests per second
const capacity = 10              // max tokens in the bucket

// package level — persists between requests
var rateLimits = make(map[string]Bucket)

func checkRateLimit(userID string) bool {
    now := time.Now()
    bucket, exists := rateLimits[userID]
    if !exists {
        bucket = Bucket{tokens: capacity - 1, lastRefillTime: now}
    }

    // refill tokens based on how much time has passed
    elapsedTime := now.Sub(bucket.lastRefillTime).Seconds()
    bucket.tokens += elapsedTime * float64(AllowedTokenPerSecond)
    if bucket.tokens > float64(capacity) {
        bucket.tokens = float64(capacity) // cap at max capacity
    }

    if bucket.tokens < 1 {
        rateLimits[userID] = bucket
        return false // no tokens left
    }

    bucket.lastRefillTime = now
    bucket.tokens -= 1 // consume a token for this request
    rateLimits[userID] = bucket
    return true
}

func testUser(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID := r.URL.Query().Get("id")
    if !checkRateLimit(userID) {
        http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
        return
    }

    http.ServeFile(w, r, "index.html")
}

func main() {
    http.HandleFunc("/test", testUser)
    fmt.Println("server running on port 8080")
    http.ListenAndServe(":8080", nil)
}
```

How the token bucket works:
- Every user starts with a full bucket of tokens
- Each request consumes 1 token
- Tokens refill automatically over time based on elapsed seconds
- If the bucket is empty the request is rejected
- A user who slows down naturally gets their tokens back — unlike a fixed counter that resets only at an interval

---

## Common Mistakes Coming From JS/C++

| Mistake | Why it happens | Fix |
|---|---|---|
| `:=` at package level | Habit from inside functions | Use `var` at package level |
| Ignoring errors with `_` | JS has try/catch | Always check `err != nil` |
| Not returning after error | No exceptions to stop flow | Always `return` after error |
| Map inside function | Variable scoping assumption | Declare maps at package level if state needs to persist |
| No mutex on shared map | Single threaded JS habit | Use `sync.Mutex` for any shared state |
| Expecting goroutines to finish | JS async behavior assumption | Use `sync.WaitGroup` |

---

## What to Learn Next

| Topic | Why |
|---|---|
| `chi` or `gorilla/mux` | Proper routing with path params like `/user/{id}` |
| `database/sql` + `sqlx` | Connecting to a real database |
| `context.Context` | Timeouts and cancellation on requests |
| `os.Getenv` + `.env` | Managing secrets and config |
| JWT middleware | Auth in APIs |
| Embedding | Code reuse without inheritance |

---

*Built while learning Go — focused on what actually matters for backend systems.*
