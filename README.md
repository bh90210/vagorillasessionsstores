# vaGorillaSessionsStores [![GoDoc](https://godoc.org/github.com/bh90210/va-gorilla-sessions-store?status.svg)](https://godoc.org/github.com/bh90210/va-gorilla-sessions-store)
Collection of various Gorilla Sessions back-end stores.

# Install

```bash
go get github.com/bh90210/vaGorillaSessionsStores
```

# Use
Errors are excluded for brevity.

## Badger
_note: Badger will not work in distributed environments. Use it for local testing or single server scenarios._

### Using the store is very simple:
```go
import "github.com/bh90210/vaGorillaSessionsStores/badger"

store, _ := badger.NewBadgerStore("/path/to/data", []byte(os.Getenv("SESSION_KEY")))
```
If `path` is empty data will be stored in system's `tmp` directory.

### Start a store with custom options (see [Badger's docs](https://dgraph.io/docs/badger) for more):
```go
import "github.com/bh90210/vaGorillaSessionsStores/badger"

opts := badger.Options{
		Dir: "/data/dir",
}
store, _ := badger.NewBadgerStoreWithOpts(opts)
```

## Mongo

### Starting a store with the default options:
```go
import "github.com/bh90210/vaGorillaSessionsStores/mongo"

store, _ := mongo.NewMongoStore("mongodb://localhost:27017")
```

### With custom options:
```go
import (
	mongoStore "github.com/bh90210/vaGorillaSessionsStores/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

client, _ := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
store, _ := mongoStore.NewMongoStoreWithOpts(client)
```

## Dgraph
_work in progress_

## Help function
Each store provides two helper functions for direct back-end session manipulation without http request. 

### Edit
_work in progress_

### Delete 
_work in progress_
