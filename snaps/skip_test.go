package snaps

import (
	"os"
	"sync"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps/internal/test"
)

func TestSkip(t *testing.T) {
	t.Run("should call Skip", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})
		skipArgs := []interface{}{1, 2, 3, 4, 5}

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockSkip: func(args ...interface{}) {
				test.Equal(t, skipArgs, args)
			},
			MockName: func() string {
				return "mock-test"
			},
			MockLog: func(args ...interface{}) {
				test.Equal(t, skippedMsg, args[0])
			},
		}
		Skip(mockT, 1, 2, 3, 4, 5)

		test.Equal(t, []string{"mock-test"}, skippedTests.values)
	})

	t.Run("should call Skipf", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockSkipf: func(format string, args ...interface{}) {
				test.Equal(t, "mock", format)
				test.Equal(t, []interface{}{1, 2, 3, 4, 5}, args)
			},
			MockName: func() string {
				return "mock-test"
			},
			MockLog: func(args ...interface{}) {
				test.Equal(t, skippedMsg, args[0])
			},
		}
		Skipf(mockT, "mock", 1, 2, 3, 4, 5)

		test.Equal(t, []string{"mock-test"}, skippedTests.values)
	})

	t.Run("should call SkipNow", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})

		mockT := test.MockTestingT{
			MockHelper:  func() {},
			MockSkipNow: func() {},
			MockName: func() string {
				return "mock-test"
			},
			MockLog: func(args ...interface{}) {
				test.Equal(t, skippedMsg, args[0])
			},
		}
		SkipNow(mockT)

		test.Equal(t, []string{"mock-test"}, skippedTests.values)
	})

	t.Run("should be concurrent safe", func(t *testing.T) {
		t.Cleanup(func() {
			skippedTests = newSyncSlice()
		})

		mockT := test.MockTestingT{
			MockHelper:  func() {},
			MockSkipNow: func() {},
			MockName: func() string {
				return "mock-test"
			},
			MockLog: func(args ...interface{}) {
				test.Equal(t, skippedMsg, args[0])
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

		test.Equal(t, 1000, len(skippedTests.values))
	})

	t.Run("testSkipped", func(t *testing.T) {
		t.Run("should return true if testID is not part of the 'runOnly'", func(t *testing.T) {
			runOnly := "TestMock"
			testID := "TestSkip/should_call_Skip - 1"

			received := testSkipped(testID, runOnly)
			test.Equal(t, true, received)
		})

		t.Run("should return false if testID is part of 'runOnly'", func(t *testing.T) {
			runOnly := "TestMock"
			testID := "TestMock/Test/should_be_not_skipped - 2"

			received := testSkipped(testID, runOnly)
			test.Equal(t, false, received)
		})

		t.Run(
			"should check if the parent is skipped and mark child tests as skipped",
			func(t *testing.T) {
				t.Cleanup(func() {
					skippedTests = newSyncSlice()
				})

				runOnly := ""
				mockT := test.MockTestingT{
					MockHelper:  func() {},
					MockSkipNow: func() {},
					MockName: func() string {
						return "TestMock/Skip"
					},
					MockLog: func(args ...interface{}) {
						test.Equal(t, skippedMsg, args[0])
					},
				}
				// This is for populating skippedTests.values and following the normal flow
				SkipNow(mockT)

				test.Equal(t, true, testSkipped("TestMock/Skip", runOnly))
				test.Equal(
					t,
					true,
					testSkipped("TestMock/Skip/child_should_also_be_skipped", runOnly),
				)
				test.Equal(t, false, testSkipped("TestAnotherTest", runOnly))
			},
		)

		t.Run("should not mark tests skipped if not not a child", func(t *testing.T) {
			t.Cleanup(func() {
				skippedTests = newSyncSlice()
			})

			runOnly := ""
			mockT := test.MockTestingT{
				MockHelper:  func() {},
				MockSkipNow: func() {},
				MockName: func() string {
					return "Test"
				},
				MockLog: func(args ...interface{}) {
					test.Equal(t, skippedMsg, args[0])
				},
			}
			// This is for populating skippedTests.values and following the normal flow
			SkipNow(mockT)

			test.Equal(t, true, testSkipped("Test", runOnly))
			test.Equal(t, true, testSkipped("Test/chid", runOnly))
			test.Equal(t, false, testSkipped("TestMock", runOnly))
			test.Equal(t, false, testSkipped("TestMock/child", runOnly))
		})

		t.Run("should use regex match for runOnly", func(t *testing.T) {
			test.Equal(t, false, testSkipped("MyTest - 1", "Test"))
			test.Equal(t, true, testSkipped("MyTest - 1", "^Test"))
		})
	})

	t.Run("isFileSkipped", func(t *testing.T) {
		t.Run("should return 'false'", func(t *testing.T) {
			test.Equal(t, false, isFileSkipped("", "", ""))
		})

		t.Run("should return 'true' if test is not included in the test file", func(t *testing.T) {
			dir, _ := os.Getwd()

			test.Equal(
				t,
				true,
				isFileSkipped(dir+"/__snapshots__", "skip_test.snap", "TestNonExistent"),
			)
		})

		t.Run("should return 'false' if test is included in the test file", func(t *testing.T) {
			dir, _ := os.Getwd()

			test.Equal(t, false, isFileSkipped(dir+"/__snapshots__", "skip_test.snap", "TestSkip"))
		})

		t.Run("should use regex match for runOnly", func(t *testing.T) {
			dir, _ := os.Getwd()

			test.Equal(t, false, isFileSkipped(dir+"/__snapshots__", "skip_test.snap", "	TestSkip.*"))
		})
	})
}
