package sqldb

import (
	"database/sql/driver"
	"regexp"
	"strings"
)

func values(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, 0, len(args))
	for _, a := range args {
		values = append(values, a.Value)
	}
	return values
}

var sqlreg = regexp.MustCompile(`(?i)` +
	`(?P<cmd>select)\s+.+?from\s+(?P<table>\w+)\s+|` +
	`(?P<cmd>update)\s+(?P<table>\w+)\s+|` +
	`(?P<cmd>delete)\s+from\s+(?P<table>\w+)\s+|` +
	`(?P<cmd>insert)\s+into\s+(?P<table>\w+)`)

// 提取 sql 的表名和指令
//
// "select * from foo ..." => foo,select
func parseSQL(sql string) (table, cmd string) {
	matches := sqlreg.FindStringSubmatch(sql)

	results := map[string]string{}
	names := sqlreg.SubexpNames()
	for i, match := range matches {
		if match != "" {
			results[names[i]] = match
		}
	}

	table = strings.ToLower(results["table"])
	cmd = strings.ToLower(results["cmd"])
	return
}
