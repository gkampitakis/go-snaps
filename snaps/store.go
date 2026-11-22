package snaps

import "errors"

// snapshotStore abstracts where a snapshot lives, so that the compare, create, and update
// flow can be shared between the different kinds of snapshots
type snapshotStore interface {
	// get returns the previous snapshot and the line it is located at,
	// or errSnapNotFound if there is no previous snapshot
	get() (string, int, error)
	add(snapshot string) error
	update(snapshot string) error
	// relPath is the path shown to the user when there is a diff
	relPath() string
}

// snapFileStore stores a snapshot as one of many entries within a .snap file
type snapFileStore struct {
	testID, path, rel string
}

func (s snapFileStore) get() (string, int, error) {
	return getPrevSnapshot(s.testID, s.path)
}

func (s snapFileStore) add(snapshot string) error {
	return addNewSnapshot(s.testID, snapshot, s.path)
}

func (s snapFileStore) update(snapshot string) error {
	return updateSnapshot(s.testID, snapshot, s.path)
}

func (s snapFileStore) relPath() string {
	return s.rel
}

// standaloneStore stores a snapshot as the sole content of its own file
type standaloneStore struct {
	path, rel string
}

func (s standaloneStore) get() (string, int, error) {
	prev, err := getPrevStandaloneSnapshot(s.path)

	return prev, 1, err
}

func (s standaloneStore) add(snapshot string) error {
	return upsertStandaloneSnapshot(snapshot, s.path)
}

func (s standaloneStore) update(snapshot string) error {
	return upsertStandaloneSnapshot(snapshot, s.path)
}

func (s standaloneStore) relPath() string {
	return s.rel
}

// inlineStore stores a snapshot as an argument to the MatchInlineSnapshot call in the test's source
type inlineStore struct {
	file string
	line int
	prev inlineSnapshot
}

func (s inlineStore) get() (string, int, error) {
	if s.prev == nil {
		return "", -1, errSnapNotFound
	}

	// inline snapshots have no meaningful line, which is signalled to prettyDiff with -1
	return *s.prev, -1, nil
}

func (s inlineStore) add(snapshot string) error {
	return upsertInlineSnapshot(s.file, s.line, snapshot)
}

func (s inlineStore) update(snapshot string) error {
	return upsertInlineSnapshot(s.file, s.line, snapshot)
}

func (s inlineStore) relPath() string {
	return ""
}

// compareAndStore compares the snapshot against the previous one held by the store,
// creating or updating it as permitted by the config, and registering the outcome
func compareAndStore(c *Config, t testingT, snapshot string, s snapshotStore, unescape bool) {
	t.Helper()

	prevSnapshot, line, err := s.get()
	if errors.Is(err, errSnapNotFound) {
		if !shouldCreate(c.update) {
			handleError(t, err)
			return
		}

		if err := s.add(snapshot); err != nil {
			handleError(t, err)
			return
		}

		t.Log(addedMsg)
		testEvents.register(added)
		return
	}
	if err != nil {
		handleError(t, err)
		return
	}

	// the escaping is only for storage, so it should not show up in the diff
	received := snapshot
	if unescape {
		prevSnapshot, received = unescapeEndChars(prevSnapshot), unescapeEndChars(snapshot)
	}

	diff := prettyDiff(prevSnapshot, received, s.relPath(), line)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate(c.update) {
		handleError(t, diff)
		return
	}

	if err := s.update(snapshot); err != nil {
		handleError(t, err)
		return
	}

	t.Log(updatedMsg)
	testEvents.register(updated)
}
