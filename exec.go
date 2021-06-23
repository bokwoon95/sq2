package sq

import (
	"bytes"
	"context"
	"errors"
	"time"
)

func Exec(db Queryer, q Query, execflag int) (rowsAffected, lastInsertID int64, err error) {
	return execContext(context.Background(), db, q, execflag, 1)
}

func ExecContext(ctx context.Context, db Queryer, q Query, execflag int) (rowsAffected, lastInsertID int64, err error) {
	return execContext(ctx, db, q, execflag, 1)
}

func execContext(ctx context.Context, db Queryer, q Query, execflag int, skip int) (rowsAffected, lastInsertID int64, err error) {
	if db == nil {
		return 0, 0, errors.New("sq: db is nil")
	}
	if q == nil {
		return 0, 0, errors.New("sq: query is nil")
	}
	var stats QueryStats
	var logQueryStats func(ctx context.Context, stats QueryStats, skip int)
	if db, ok := db.(QueryerLogger); ok {
		logQueryStats = db.LogQueryStats
	}
	stats.Dialect = q.Dialect()
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		if stats.Query == "" && err != nil {
			stats.Query = buf.String() + "%!(error=" + err.Error() + ")"
		}
		buf.Reset()
		bufpool.Put(buf)
		if logQueryStats == nil {
			return
		}
		stats.Error = err
		if ErowsAffected&execflag != 0 {
			stats.RowsAffected.Valid = true
			stats.RowsAffected.Int64 = rowsAffected
		}
		if ElastInsertID&execflag != 0 {
			stats.LastInsertID.Valid = true
			stats.LastInsertID.Int64 = rowsAffected
		}
		logQueryStats(ctx, stats, skip+2)
	}()
	err = q.AppendSQL(stats.Dialect, buf, &stats.Args, make(map[string][]int))
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
	}
	if result != nil && ElastInsertID&execflag != 0 {
		lastInsertID, err = result.LastInsertId()
		if err != nil {
			return 0, 0, err
		}
	}
	return rowsAffected, lastInsertID, nil
}
