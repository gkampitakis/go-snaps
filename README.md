# GO Snaps

Jest-like snapshot testing in Golang



## Installation

To install `go-snaps`, use `go get`:

```bash
go get github.com/gkampitakis/go-snaps
```

Import the `go-snaps/snaps` package into your:

```go
package example

import (
  "testing"

	"github.com/gkampitakis/go_snaps/snaps"
)

func TestExample(t *testing.T) {

  snaps.MatchSnapshot(t ,"Hello World")

}
```

### Usage

You can pass multiple parameters to `MatchSnapshot` or call `MatchSnapshot` multiple
times inside the same test. The difference is in the snapshot file the latter will
create multiple entries.

```go

func TestSimple(t *testing.T) {

  t.Run("should make multiple entries in snapshot", func(t *testing.T) {
  
    snaps.MatchSnapshot(t, 5, 10, 20, 25)
    snaps.MatchSnapshot(t, "some value")
  
  })

}

```

By default `go-snaps` saves the snapshots in `__snapshots__` directory and the file
name is the test file name with extension `.snap`. 

So for example if your test is called `add_test.go` when you run your tests at the same
directory a new folder will be created `./__snapshots__/add_test.snaps`. You can 
change the extension or the directory name if you wish.

The example below will create a snapshot at `./mySnaps/<file-name>.txt`.
```go
func TestSimpleConfig(t *testing.T) {
	s := snaps.New(snaps.SnapsDirectory("mySnaps"), snaps.SnapsExtension("txt"))

	s.MatchSnapshot(t, 10)
}
```

Finally you can update your failing snapshots by setting `UPDATE_SNAPS` env variable to true.

```bash
UPDATE_SNAPS=true go test ....
```

You can also see some example usages in `./examples` in this project.

### Acknowledgments

This library used [Jest Snapshoting](https://jestjs.io/docs/snapshot-testing) and [Cupaloy](https://github.com/bradleyjkemp/cupaloy) as inspiration.

- Jest is a full-fledged Javascript testing framework and has robust snapshoting features.
- Cupaloy is a great and simple Golang snapshoting solution.

### Run examples

```bash
go test ./examples/... -v -count=1
```

#### License 

MIT
