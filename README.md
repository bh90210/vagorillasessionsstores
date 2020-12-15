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
import stores "github.com/bh90210/vagorillasessionsstores"

store, _ := stores.NewBadgerStore("/path/to/data", []byte(os.Getenv("SESSION_KEY")))
```
If `path` is empty data will be stored in system's `tmp` directory.

### Start a store with custom options (see [Badger's docs](https://dgraph.io/docs/badger) for more):
```go
import stores "github.com/bh90210/vagorillasessionsstores"

opts := badger.Options{
		Dir: "/data/dir",
}
store, _ := stores.NewBadgerStoreWithOpts(opts)
```

## Mongo

### Starting a store entails passing credentials and client options (official go mongo driver is necessary):
```go
import (
	stores "github.com/bh90210/vagorillasessionsstores"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cred options.Credential
cred.AuthSource = "YourAuthSource"
cred.Username = "YourUserName"
cred.Password = "YourPassword"
clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI")).SetAuth(cred)
store, _ := stores.NewMongoStore(clientOptions)
```

## Dgraph
_work in progress_

## Help functions
Each store provides two helper functions for direct back-end session manipulation without http request. 

### Edit
_work in progress_

### Delete 
_work in progress_
