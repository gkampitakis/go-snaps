package snaps

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// MatchSnapshot verifies the values match the most recent snap file
//
// You can pass multiple values
// 	MatchSnapshot(t, 10, "hello world")
// or call MatchSnapshot multiples times inside a test
//  MatchSnapshot(t, 10)
//  MatchSnapshot(t, "hello world")
// The difference is the latter will create multiple entries.
func MatchSnapshot(t testingT, value interface{}) {
	t.Helper()

	dir, snapPath := snapDirAndName()
	testID := testsRegistry.getTestID(t.Name(), snapPath)
	snapshot := takeSnapshot(value)
	// Load snapshot
	loadSnapshotFile(snapPath)
	prevSnapshot, err := getPrevSnapshot(testID, snapPath)

	if errors.Is(err, errSnapNotFound) {
		if isCI {
			t.Error(err)
			return
		}

		// NOTE: issue with extraneous newline at the last entry
		err := addNewSnapshot(testID, snapshot, dir, snapPath)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log(greenText(arrowPoint + "New snapshot written.\n"))
		return
	}
	if err != nil {
		t.Error(err)
		return
	}

	diff := prettyDiff(prevSnapshot, snapshot)
	if diff == "" {
		return
	}

	if !shouldUpdate {
		t.Error(diff)
		return
	}

	// if err = updateSnapshot(testID, snapshot, snapPath); err != nil {
	// 	t.Error(err)
	// 	return
	// }

	t.Log(greenText(arrowPoint + "Snapshot updated.\n"))
}

func getPrevSnapshot(testID, snapPath string) (string, error) {
	snapshot, exists := snapshotMap.get(snapPath, testID)
	if !exists {
		return snapshot, errSnapNotFound
	}

	return snapshot, nil
}

func stringToSnapshotFile(snap, name string) error {
	return os.WriteFile(name, []byte(snap), os.ModePerm)
}

func addNewSnapshot(testID, snapshot, dir, snapPath string) error {
	// _m.Lock()
	// defer _m.Unlock()
	// BUG: <- this creates race conditions
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(snapPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	e.Encode(map[string]string{
		testID: snapshot,
	})
	err = e.Close()
	if err != nil {
		return err
	}

	_, err = f.WriteString(newLine)
	return err
}

/*
	Returns the dir for snapshots [where the tests run + /snapsDirName]
	and the name [dir + /snapsDirName + /<test-name>.snapsExtName]
*/
func snapDirAndName() (dir, name string) {
	callerPath, _ := baseCaller()
	base := filepath.Base(callerPath)

	dir = filepath.Join(filepath.Dir(callerPath), snapsDir)
	name = filepath.Join(dir, strings.TrimSuffix(base, filepath.Ext(base))+snapsExt)

	return
}

func loadSnapshotFile(snapPath string) error {
	if _, err := os.Stat(snapPath); err != nil {
		return errSnapNotFound
	}

	if snapshotMap.fileLoaded(snapPath) {
		return nil
	}

	snapshots := map[string]string{}

	f, err := os.Open(snapPath)
	if err != nil {
		return err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.Decode(snapshots)

	snapshotMap.setFile(snapPath, snapshots)
	return nil
}

// func updateSnapshot(testID, snapshot, snapPath string) error {
// 	// When t.Parallel a test can override another snapshot as we dump
// 	// all snapshots
// 	_m.Lock()
// 	defer _m.Unlock()
// 	f, err := snapshotFileToString(snapPath, testID)
// 	if err != nil {
// 		return err
// 	}

// 	updatedSnapFile := dynamicTestIDRegexp(testID).(f, fmt.Sprintf("%s\n%s---", testID, snapshot))

// 	return stringToSnapshotFile(updatedSnapFile, snapPath)
// }
