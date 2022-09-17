package snaps

func MatchJSON(t testingT, input interface{}, matchers ...interface{}) {
	t.Helper()

	switch tp := input.(type) {
	case string:
		t.Error("string not implemented")
	case []byte:
		t.Error("[]byte not implemented")
	default:
		t.Errorf("type: %T not supported\n", tp)
	}
}
