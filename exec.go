package sq

import (
	"bytes"
	"context"
	"errors"
	"time"
)

func Exec(db DB, q Query) (rowsAffected, lastInsertID int64, err error) {
	return execContext(context.Background(), db, q, 1)
}

func ExecContext(ctx context.Context, db DB, q Query) (rowsAffected, lastInsertID int64, err error) {
	return execContext(ctx, db, q, 1)
}

func execContext(ctx context.Context, db DB, q Query, skip int) (rowsAffected, lastInsertID int64, err error) {
	if db == nil {
		return 0, 0, errors.New("sq: db is nil")
	}
	if q == nil {
		return 0, 0, errors.New("sq: query is nil")
	}
	var stats QueryStats
	var logQueryStats func(ctx context.Context, stats QueryStats)
	var logSettings LogSettings
	if loggerDB, ok := db.(LoggerDB); ok {
		logSettings, err = loggerDB.GetLogSettings()
		if err != nil {
			if !errors.Is(err, ErrLoggerUnsupported) {
				return 0, 0, err
			}
		} else {
			logQueryStats = loggerDB.LogQueryStats
		}
	}
	if logQueryStats != nil && logSettings.GetCallerInfo {
		stats.CallerFile, stats.CallerLine, stats.CallerFunction = caller(skip)
	}
	switch q := q.(type) {
	case SelectQuery:
		stats.Env = q.Env
		stats.QueryType = "SELECT"
	case InsertQuery:
		stats.Env = q.Env
		stats.QueryType = "INSERT"
		if table := q.IntoTable; table != nil {
			stats.TableModified[0] = table.GetSchema()
			stats.TableModified[1] = table.GetName()
		}
	case UpdateQuery:
		stats.Env = q.Env
		stats.QueryType = "UPDATE"
		if table := q.UpdateTable; table != nil {
			stats.TableModified[0] = table.GetSchema()
			stats.TableModified[1] = table.GetName()
		}
	case DeleteQuery:
		stats.Env = q.Env
		stats.QueryType = "DELETE"
		if len(q.FromTables) > 0 {
			if table := q.FromTables[0]; table != nil {
				stats.TableModified[0] = table.GetSchema()
				stats.TableModified[1] = table.GetName()
			}
		}
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
		if logSettings.AsyncLogging {
			go logQueryStats(ctx, stats)
		} else {
			logQueryStats(ctx, stats)
		}
	}()
	err = q.AppendSQL(stats.Dialect, buf, &stats.Args, make(map[string][]int), nil)
	if err != nil {
		return 0, 0, err
	}
	stats.Query = buf.String()
	var start time.Time
	if logSettings.TimeQuery {
		start = time.Now()
	}
	result, err := db.ExecContext(ctx, stats.Query, stats.Args...)
	if logSettings.TimeQuery {
		stats.TimeTaken = time.Since(start)
	}
	if err != nil {
		return 0, 0, err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	stats.RowsAffected.Valid = true
	stats.RowsAffected.Int64 = rowsAffected
	if stats.Dialect == DialectSQLite || stats.Dialect == DialectMySQL {
		lastInsertID, err = result.LastInsertId()
		if err != nil {
			return 0, 0, err
		}
		stats.LastInsertID.Valid = true
		stats.LastInsertID.Int64 = lastInsertID
	}
	return rowsAffected, lastInsertID, nil
}
