package snaps

import (
	"os"
	"sync"
	"testing"
)

type MockTestingT struct {
	mockHelper  func()
	mockName    func() string
	mockSkip    func(args ...interface{})
	mockSkipf   func(format string, args ...interface{})
	mockSkipNow func()
	mockError   func(args ...interface{})
	mockLog     func(args ...interface{})
}

func (m MockTestingT) Error(args ...interface{}) {
	m.mockError(args...)
}

func (m MockTestingT) Helper() {
	m.mockHelper()
}

func (m MockTestingT) Skip(args ...interface{}) {
	m.mockSkip(args...)
}

func (m MockTestingT) Skipf(format string, args ...interface{}) {
	m.mockSkipf(format, args...)
}

func (m MockTestingT) SkipNow() {
	m.mockSkipNow()
}

func (m MockTestingT) Name() string {
	return m.mockName()
}

func (m MockTestingT) Log(args ...interface{}) {
	m.mockLog(args...)
}

func TestSkip(t *testing.T) {
	t.Run("should call Skip", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})
		skipArgs := []interface{}{1, 2, 3, 4, 5}

		mockT := MockTestingT{
			mockSkip: func(args ...interface{}) {
				Equal(t, skipArgs, args)
			},
			mockHelper: func() {},
			mockName: func() string {
				return "mock-test"
			},
		}
		Skip(mockT, 1, 2, 3, 4, 5)

		Equal(t, []string{"mock-test"}, skippedTests.values)
	})

	t.Run("should call Skipf", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})

		mockT := MockTestingT{
			mockSkipf: func(format string, args ...interface{}) {
				Equal(t, "mock", format)
				Equal(t, []interface{}{1, 2, 3, 4, 5}, args)
			},
			mockHelper: func() {},
			mockName: func() string {
				return "mock-test"
			},
		}
		Skipf(mockT, "mock", 1, 2, 3, 4, 5)

		Equal(t, []string{"mock-test"}, skippedTests.values)
	})

	t.Run("should call SkipNow", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})

		mockT := MockTestingT{
			mockSkipNow: func() {},
			mockHelper:  func() {},
			mockName: func() string {
				return "mock-test"
			},
		}
		SkipNow(mockT)

		Equal(t, []string{"mock-test"}, skippedTests.values)
	})

	t.Run("should be concurrent safe", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})

		mockT := MockTestingT{
			mockSkipNow: func() {},
			mockHelper:  func() {},
			mockName: func() string {
				return "mock-test"
			},
		}

		wg := sync.WaitGroup{}

		for i := 0; i < 1000; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				SkipNow(mockT)
			}()
		}

		wg.Wait()

		Equal(t, 1000, len(skippedTests.values))
	})

	t.Run("testSkipped", func(t *testing.T) {
		t.Run("should return true if testID is not part of the 'runOnly'", func(t *testing.T) {
			runOnly := "TestMock"
			testID := "TestSkip/should_call_Skip - 1"

			received := testSkipped(testID, runOnly)
			Equal(t, true, received)
		})

		t.Run("should return false if testID is part of 'runOnly'", func(t *testing.T) {
			runOnly := "TestMock"
			testID := "TestMock/Test/should_be_not_skipped - 2"

			received := testSkipped(testID, runOnly)
			Equal(t, false, received)
		})

		t.Run(
			"should check if the parent is skipped and mark child tests as skipped",
			func(t *testing.T) {
				t.Cleanup(func() {
					skippedTests = newSyncSlice()
				})

				runOnly := ""
				mockT := MockTestingT{
					mockSkipNow: func() {},
					mockHelper:  func() {},
					mockName: func() string {
						return "TestMock/Skip"
					},
				}
				// This is for populating skippedTests.values and following the normal flow
				SkipNow(mockT)

				Equal(t, true, testSkipped("TestMock/Skip", runOnly))
				Equal(t, true, testSkipped("TestMock/Skip/child_should_also_be_skipped", runOnly))
				Equal(t, false, testSkipped("TestAnotherTest", runOnly))
			},
		)

		t.Run("should not mark tests skipped if not not a child", func(t *testing.T) {
			t.Cleanup(func() {
				skippedTests = newSyncSlice()
			})

			runOnly := ""
			mockT := MockTestingT{
				mockSkipNow: func() {},
				mockHelper:  func() {},
				mockName: func() string {
					return "Test"
				},
			}
			// This is for populating skippedTests.values and following the normal flow
			SkipNow(mockT)

			Equal(t, true, testSkipped("Test", runOnly))
			Equal(t, true, testSkipped("Test/chid", runOnly))
			Equal(t, false, testSkipped("TestMock", runOnly))
			Equal(t, false, testSkipped("TestMock/child", runOnly))
		})

		t.Run("should use regex match for runOnly", func(t *testing.T) {
			Equal(t, false, testSkipped("MyTest - 1", "Test"))
			Equal(t, true, testSkipped("MyTest - 1", "^Test"))
		})
	})

	t.Run("isFileSkipped", func(t *testing.T) {
		t.Run("should return 'false'", func(t *testing.T) {
			Equal(t, false, isFileSkipped("", "", ""))
		})

		t.Run("should return 'true' if test is not included in the test file", func(t *testing.T) {
			dir, _ := os.Getwd()

			Equal(t, true, isFileSkipped(dir+"/__snapshots__", "skip_test.snap", "TestNonExistent"))
		})

		t.Run("should return 'false' if test is included in the test file", func(t *testing.T) {
			dir, _ := os.Getwd()

			Equal(t, false, isFileSkipped(dir+"/__snapshots__", "skip_test.snap", "TestSkip"))
		})
	})
}
