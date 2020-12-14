# vaGorillaSessionsStores [![GoDoc](https://godoc.org/github.com/bh90210/va-gorilla-sessions-store?status.svg)](https://godoc.org/github.com/bh90210/va-gorilla-sessions-store)
Collection of various Gorilla Sessions back-end stores.

# Install

```bash
go get github.com/bh90210/vaGorillaSessionsStores
```

# Use

## Badger
_note: Badger will not work in distributed environments. Use it for local testing or single server scenarios._

### Using the store is very simple:
```go
	store, err := badger.NewBadgerStore("/path/to/data", []byte(os.Getenv("SESSION_KEY")))
	if err != nil {
		log.Fatal(err)
	}
```
If `path` is empty data will be stores in system's `tmp` directory.

### Start a store with custom options (see [Badger's docs](https://dgraph.io/docs/badger) for more):
```go
    opts := badger.Options{}
    store, err := badger.NewBadgerStoreWithOpts(opts)
    if err != nil {
		log.Fatal(err)
	}
```

## Mongo

### Starting a store with the default options:
```go
```

### With custom options:
```go
```

## Dgraph

_work in progress_