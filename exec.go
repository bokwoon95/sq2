package sq

import (
	"bytes"
	"context"
	"errors"
	"time"
)

func Exec(db DB, q Query, execflag int) (rowsAffected, lastInsertID int64, err error) {
	return execContext(context.Background(), db, q, execflag, 1)
}

func ExecContext(ctx context.Context, db DB, q Query, execflag int) (rowsAffected, lastInsertID int64, err error) {
	return execContext(ctx, db, q, execflag, 1)
}

func execContext(ctx context.Context, db DB, q Query, execflag int, skip int) (rowsAffected, lastInsertID int64, err error) {
	if db == nil {
		return 0, 0, errors.New("sq: db is nil")
	}
	if q == nil {
		return 0, 0, errors.New("sq: query is nil")
	}
	var stats QueryStats
	var logQueryStats func(ctx context.Context, stats QueryStats)
	var logSettings LogSettings
	if db, ok := db.(LoggerDB); ok {
		logQueryStats = db.LogQueryStats
		logSettings = db.GetLogSettings()
	}
	if logSettings.GetCallerInfo {
		stats.CallerFile, stats.CallerLine, stats.CallerFunction = caller(skip)
	}
	switch q := q.(type) {
	case SelectQuery:
		stats.Env = q.Env
	case InsertQuery:
		stats.Env = q.Env
	case UpdateQuery:
		stats.Env = q.Env
	case DeleteQuery:
		stats.Env = q.Env
	}
	stats.Dialect = q.GetDialect()
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	defer func() {
		if logQueryStats == nil {
			return
		}
		if stats.Query == "" && err != nil {
			stats.Query = buf.String() + "%!(error=" + err.Error() + ")"
		}
		stats.Error = err
		go logQueryStats(ctx, stats)
	}()
	err = q.AppendSQL(stats.Dialect, buf, &stats.Args, make(map[string][]int), nil)
	if err != nil {
		return 0, 0, err
	}
	stats.Query = buf.String()
	start := time.Now()
	result, err := db.ExecContext(ctx, stats.Query, stats.Args...)
	stats.TimeTaken = time.Since(start)
	if err != nil {
		return 0, 0, err
	}
	if result != nil && ErowsAffected&execflag != 0 {
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return 0, 0, err
		}
		stats.RowsAffected.Valid = true
		stats.RowsAffected.Int64 = rowsAffected
	}
	if result != nil && ElastInsertID&execflag != 0 {
		lastInsertID, err = result.LastInsertId()
		if err != nil {
			return 0, 0, err
		}
		stats.LastInsertID.Valid = true
		stats.LastInsertID.Int64 = lastInsertID
	}
	return rowsAffected, lastInsertID, nil
}
