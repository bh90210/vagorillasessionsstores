# vagorillasessionsstores [![GoDoc](https://godoc.org/github.com/bh90210/va-gorilla-sessions-store?status.svg)](https://godoc.org/github.com/bh90210/va-gorilla-sessions-store)
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
store, _ := stores.NewBadgerStoreWithOpts(opts,[]byte(os.Getenv("SESSION_KEY")))
```
### Help functions
Two helper functions for direct back-end session manipulation without http request. 

## Mongo

### Starting a store entails passing credentials and client options (official go mongo driver is necessary):
```go
import (
	stores "github.com/bh90210/vagorillasessionsstores"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cred options.Credential

cred.AuthSource = "YourAuthSource"
cred.Username = "YourUserName"
cred.Password = "YourPassword"

clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_DB_URI")).SetAuth(cred)

ctx, _ := context.Background()
client, _ := mongo.Connect(ctx, opts)

store, _ := stores.NewMongoStore(client, "databaseName", "collectionName", []byte(os.Getenv("SESSION_KEY")))
```
_If 'databaseName' & 'collectionName' are left empty the defaults are used ('sessions' & 'store')._

## Dgraph

_store uses dgo/v200_

Assumed schema:
```yaml
sessionid: string @index(hash) .
sessionvalue: string . 
type Session {
	sessionid
	sessionvalue
}
```

```go
import (
	stores "github.com/bh90210/vagorillasessionsstores"
	"google.golang.org/grpc"
)

conn, _err_ := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())

store, _ := stores.NewDgraphStore(conn, []byte(os.Getenv("SESSION_KEY")))
```

You can also let schema initiation to the store:
```go
import (
	stores "github.com/bh90210/vagorillasessionsstores"
	"google.golang.org/grpc"
)

conn, _err_ := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())

store, _ := stores.NewDgraphStoreWithSchema(conn, []byte(os.Getenv("SESSION_KEY")))
```