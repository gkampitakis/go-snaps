package examples

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
)

type myMatcher struct {
	age int
}

func Matcher() *myMatcher {
	return &myMatcher{}
}

func (m *myMatcher) AgeGreater(a int) *myMatcher {
	m.age = a
	return m
}

func (m *myMatcher) JSON(s []byte) ([]byte, []match.MatcherError) {
	var v struct {
		User  string
		Age   int
		Email string
	}

	err := json.Unmarshal(s, &v)
	if err != nil {
		return nil, []match.MatcherError{
			{
				Reason:  err,
				Matcher: "my matcher",
				Path:    "",
			},
		}
	}

	if v.Age < m.age {
		return nil, []match.MatcherError{
			{
				Reason:  fmt.Errorf("%d is >= from %d", m.age, v.Age),
				Matcher: "my matcher",
				Path:    "age",
			},
		}
	}

	// the second string is the formatted error message
	return []byte(`{"value":"blue"}`), nil
}

func TestMatchJSON(t *testing.T) {
	t.Run("should make a json object snapshot", func(t *testing.T) {
		m := map[string]any{
			"mock-0": "value",
			"mock-1": 2,
			"mock-2": struct{ Msg string }{"Hello World"},
			"mock-3": float32(10.4),
		}

		snaps.MatchJSON(t, m)
	})

	t.Run("should create a prettyJSON snap", func(t *testing.T) {
		value := `{"user":"mock-user","age":10,"email":"mock@email.com"}`
		snaps.MatchJSON(t, value)
		snaps.MatchJSON(t, []byte(value))
	})

	t.Run("should marshal struct", func(t *testing.T) {
		type User struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Keys  []int  `json:"keys"`
		}

		u := User{
			Name:  "mock-name",
			Email: "mock@email.com",
			Keys:  []int{1, 2, 3, 4, 5},
		}

		snaps.MatchJSON(t, u)
	})

	t.Run("should create a prettyJSON snap following by config", func(t *testing.T) {
		value := `{"user":"mock-user","age":10,"email":"mock@email.com"}`
		s := snaps.WithConfig(snaps.JSON(snaps.JSONConfig{
			Indent:   "    ",
			SortKeys: false,
		}))
		s.MatchJSON(t, value)
		s.MatchJSON(t, []byte(value))
	})
}

func TestMatchers(t *testing.T) {
	t.Run("Custom matcher", func(t *testing.T) {
		t.Run("struct marshalling", func(t *testing.T) {
			type User struct {
				Name  string `json:"name"`
				Email string `json:"email"`
				Keys  []int  `json:"keys"`
			}

			u := User{
				Name:  "mock-user",
				Email: "mock-user@email.com",
				Keys:  []int{1, 2, 3, 4, 5},
			}

			snaps.MatchJSON(t, u, match.Custom("keys", func(val any) (any, error) {
				keys, ok := val.([]any)
				if !ok {
					return nil, fmt.Errorf("expected []any but got %T", val)
				}

				if len(keys) > 5 {
					return nil, fmt.Errorf("expected less than 5 keys")
				}

				return val, nil
			}))
		})

		t.Run("JSON string validation", func(t *testing.T) {
			value := `{"user":"mock-user","age":2,"email":"mock@email.com"}`

			snaps.MatchJSON(t, value, match.Custom("age", func(val any) (any, error) {
				if valInt, ok := val.(float64); !ok || valInt >= 5 {
					return nil, fmt.Errorf("expecting number less than 5")
				}

				return "<less than 5 age>", nil
			}))
		})
	})

	t.Run("Any matcher", func(t *testing.T) {
		t.Run("should ignore fields", func(t *testing.T) {
			value := fmt.Sprintf(
				`{"user":"mock-user","age":10,"nested":{"now":["%s"]}}`,
				time.Now(),
			)
			snaps.MatchJSON(t, value, match.Any("nested.now.0"))
		})

		t.Run("http response", func(t *testing.T) {
			// mock server returning a json object
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				payload := fmt.Sprintf(
					`{"data":{"message":"hello world","createdAt":"%s"}}`,
					time.Now().UTC(),
				)
				w.Write([]byte(payload))
			}))

			res, err := http.Get(s.URL)
			if err != nil {
				t.Errorf("unexpected error %s", err)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("unexpected error %s", err)
				return
			}

			snaps.MatchJSON(t, body, match.Any("data.createdAt"))
		})

		t.Run("should handle nested json arrays", func(t *testing.T) {
			j := []byte(`{
					"repositories": [
						{
							"name": "repo1",
							"commits": [
								{"sha": "abc123", "files": [{"path": "a.js", "checksum": "c1"}, {"path": "b.js", "checksum": "c2"}]},
								{"sha": "def456", "files": [{"path": "c.js", "checksum": "c3"}]}
							]
						},
						{
							"name": "repo2",
							"commits": [
								{
									"sha": "ghi789", 
									"files": [
										{"path": "d.js", "checksum": "c4"}, 
										{"path": "e.js", "checksum": "c5"}, 
										{"path": "f.js", "checksum": "c6"}
									]
								}
							]
						}
					]
				}`)

			a := match.Any("repositories.#.commits.#.files.#.checksum").HandleNestedJSONArrays()

			res, errs := a.JSON(j)
			if len(errs) != 0 {
				t.Errorf("unexpected errors %v", errs)
				return
			}

			snaps.MatchJSON(t, res)
		})
	})

	t.Run("my matcher", func(t *testing.T) {
		t.Run("should allow using your matcher", func(t *testing.T) {
			value := `{"user":"mock-user","age":10,"email":"mock@email.com"}`

			snaps.MatchJSON(t, value, Matcher().AgeGreater(5))
		})
	})

	t.Run("Type matcher", func(t *testing.T) {
		t.Run("should create snapshot with type placeholder", func(t *testing.T) {
			snaps.MatchJSON(t, `{"data":10}`, match.Type[float64]("data"))
			snaps.MatchJSON(
				t,
				`{"metadata":{"timestamp":"1687108093142"}}`,
				match.Type[map[string]any]("metadata"),
			)
		})
	})
}
