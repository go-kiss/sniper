package sqldb

import "testing"

func TestParseSQL(t *testing.T) {
	cases := [][]string{
		{"select * from foo where", "foo", "select"},
		{"update foo set", "foo", "update"},
		{"insert into foo value", "foo", "insert"},
		{"DELETE from foo where", "foo", "delete"},
	}

	for _, c := range cases {
		table, cmd := parseSQL(c[0])
		if table != c[1] || cmd != c[2] {
			t.Fatal("invalid sql", table, cmd)
		}
	}
}
