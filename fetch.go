package sq

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func Fetch(db DB, q Query, rowmapper func(*Row)) (rowCount int64, err error) {
	return fetchContext(context.Background(), db, q, rowmapper, 1)
}

func FetchContext(ctx context.Context, db DB, q Query, rowmapper func(*Row)) (rowCount int64, err error) {
	return fetchContext(ctx, db, q, rowmapper, 1)
}

func fetchContext(ctx context.Context, db DB, q Query, rowmapper func(*Row), skip int) (rowCount int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = r
			default:
				panic(r)
			}
			return
		}
	}()
	if db == nil {
		return 0, errors.New("sq: db is nil")
	}
	if q == nil {
		return 0, errors.New("sq: query is nil")
	}
	if rowmapper == nil {
		return 0, errors.New("sq: cannot call Fetch/FetchContext without a rowmapper")
	}
	var stats QueryStats
	var logSettings LogSettings
	var logQueryStats func(ctx context.Context, stats QueryStats)
	if loggerDB, ok := db.(LoggerDB); ok {
		logSettings, err = loggerDB.GetLogSettings()
		if err != nil {
			if !errors.Is(err, ErrLoggerUnsupported) {
				return rowCount, err
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
	case UpdateQuery:
		stats.Env = q.Env
		stats.QueryType = "UPDATE"
	case DeleteQuery:
		stats.Env = q.Env
		stats.QueryType = "DELETE"
	}
	stats.Dialect = q.GetDialect()
	r := &Row{}
	rowmapper(r)
	fields, dest := RowResult(r)
	q, err = q.SetFetchableFields(fields)
	if err != nil {
		return 0, err
	}
	buf := bufpool.Get().(*bytes.Buffer)
	resultsBuf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		resultsBuf.Reset()
		bufpool.Put(buf)
		bufpool.Put(resultsBuf)
	}()
	defer func() {
		if logQueryStats == nil {
			return
		}
		if stats.Query == "" && err != nil {
			stats.Query = buf.String() + "%!(error=" + err.Error() + ")"
		}
		if logSettings.ResultsLimit > 0 {
			stats.QueryResults = resultsBuf.String()
		}
		stats.Error = err
		stats.RowCount.Valid = true
		stats.RowCount.Int64 = rowCount
		if logSettings.AsyncLogging {
			go logQueryStats(ctx, stats)
		} else {
			logQueryStats(ctx, stats)
		}
	}()
	err = q.AppendSQL(stats.Dialect, buf, &stats.Args, make(map[string][]int), nil)
	if err != nil {
		return 0, err
	}
	stats.Query = buf.String()
	var start time.Time
	if logSettings.TimeQuery {
		start = time.Now()
	}
	rows, err := db.QueryContext(ctx, stats.Query, stats.Args...)
	if logSettings.TimeQuery {
		stats.TimeTaken = time.Since(start)
	}
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if len(dest) == 0 {
		return 0, nil
	}
	RowActivate(r)
	var fieldNames []string
	for rows.Next() {
		rowCount++
		// Because dest and r.dest share the same backing array, any change
		// that rows.Scan() makes to dest will be propagated to r.dest and
		// hence be visible in the user's rowmapper. Hooray Go slices!
		err = rows.Scan(dest...)
		if err != nil {
			return rowCount, decorateScanError(stats.Dialect, fields, dest, err)
		}
		if logSettings.ResultsLimit > 0 && rowCount <= int64(logSettings.ResultsLimit) {
			if len(fieldNames) == 0 {
				fieldNames = computeFieldNames(stats.Dialect, fields)
			}
			accumulateResults(stats.Dialect, resultsBuf, fieldNames, dest, rowCount)
		}
		RowReset(r)
		rowmapper(r)
		err = RowProcessingError(r)
		if err != nil {
			return rowCount, err
		}
		if RowClosed(r) {
			break
		}
	}
	err = rows.Close()
	if err != nil {
		return rowCount, err
	}
	err = rows.Err()
	if err != nil {
		return rowCount, err
	}
	if logSettings.ResultsLimit > 0 && rowCount > int64(logSettings.ResultsLimit) {
		resultsBuf.WriteString("\n...\n(" + strconv.FormatInt(rowCount-int64(logSettings.ResultsLimit), 10) + " more rows)")
	}
	return rowCount, nil
}

func decorateScanError(dialect string, fields []Field, dest []interface{}, err error) error {
	buf := bufpool.Get().(*bytes.Buffer)
	tmpbuf := bufpool.Get().(*bytes.Buffer)
	tmpargs := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		tmpbuf.Reset()
		tmpargs = tmpargs[:0]
		bufpool.Put(buf)
		bufpool.Put(tmpbuf)
		argspool.Put(tmpargs)
	}()
	for i := range dest {
		buf.WriteString("\n" + strconv.Itoa(i) + ") ")
		tmpbuf.Reset()
		tmpargs = tmpargs[:0]
		err2 := fields[i].AppendSQLExclude(dialect, tmpbuf, &tmpargs, make(map[string][]int), nil, nil)
		if err2 != nil {
			buf.WriteString("%!(error=" + err2.Error() + ")")
			continue
		}
		lhs, err2 := Sprintf(dialect, tmpbuf.String(), tmpargs)
		if err2 != nil {
			buf.WriteString("%!(error=" + err2.Error() + ")")
			continue
		}
		buf.WriteString(lhs + " => " + reflect.TypeOf(dest[i]).String())
	}
	return fmt.Errorf("please check if your mapper function is correct:%s\n%w", buf.String(), err)
}

func computeFieldNames(dialect string, fields []Field) []string {
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	var fieldNames []string
	for _, field := range fields {
		if alias := field.GetAlias(); alias != "" {
			fieldNames = append(fieldNames, alias)
			continue
		}
		buf.Reset()
		args = args[:0]
		err := field.AppendSQLExclude(dialect, buf, &args, make(map[string][]int), nil, nil)
		if err != nil {
			fieldNames = append(fieldNames, "%!(error="+err.Error()+")")
			continue
		}
		fieldName, err := Sprintf(dialect, buf.String(), args)
		if err != nil {
			fieldNames = append(fieldNames, "%!(error="+err.Error()+")")
			continue
		}
		fieldNames = append(fieldNames, fieldName)
	}
	return fieldNames
}

func accumulateResults(dialect string, buf *bytes.Buffer, fieldNames []string, dest []interface{}, rowNumber int64) {
	buf.WriteString("\n----[ Row " + strconv.FormatInt(rowNumber, 10) + " ]----")
	for i := range dest {
		buf.WriteString("\n")
		if i < len(fieldNames) {
			buf.WriteString(fieldNames[i])
		}
		buf.WriteString(": ")
		rhs, err := Sprint(dialect, dest[i])
		if err != nil {
			buf.WriteString("%!(error=" + err.Error() + ")")
			continue
		}
		buf.WriteString(rhs)
	}
}

func FetchExists(db DB, q Query) (exists bool, err error) {
	return fetchExistsContext(context.Background(), db, q, 1)
}

func FetchExistsContext(ctx context.Context, db DB, q Query) (exists bool, err error) {
	return fetchExistsContext(context.Background(), db, q, 1)
}

func fetchExistsContext(ctx context.Context, db DB, q Query, skip int) (exists bool, err error) {
	if db == nil {
		return false, errors.New("sq: db is nil")
	}
	if q == nil {
		return false, errors.New("sq: query is nil")
	}
	var stats QueryStats
	var logQueryStats func(ctx context.Context, stats QueryStats)
	var logSettings LogSettings
	if loggerDB, ok := db.(LoggerDB); ok {
		logSettings, err = loggerDB.GetLogSettings()
		if err != nil {
			if !errors.Is(err, ErrLoggerUnsupported) {
				return exists, err
			}
		} else {
			logQueryStats = loggerDB.LogQueryStats
		}
	}
	if logQueryStats != nil && logSettings.GetCallerInfo {
		stats.CallerFile, stats.CallerLine, stats.CallerFunction = caller(skip)
	}
	stats.QueryType = "SELECT"
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
	fields, err := q.GetFetchableFields()
	if err != nil {
		return false, err
	}
	if len(fields) == 0 {
		q, err = q.SetFetchableFields([]Field{Literal("1")})
		if err != nil {
			return false, err
		}
	}
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
		stats.Exists.Valid = true
		stats.Exists.Bool = exists
		if logSettings.AsyncLogging {
			go logQueryStats(ctx, stats)
		} else {
			logQueryStats(ctx, stats)
		}
	}()
	buf.WriteString("SELECT EXISTS (")
	err = q.AppendSQL(stats.Dialect, buf, &stats.Args, make(map[string][]int), nil)
	if err != nil {
		return false, err
	}
	buf.WriteString(")")
	stats.Query = buf.String()
	var start time.Time
	if logSettings.TimeQuery {
		start = time.Now()
	}
	rows, err := db.QueryContext(ctx, stats.Query, stats.Args...)
	if logSettings.TimeQuery {
		stats.TimeTaken = time.Since(start)
	}
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&exists)
		if err != nil {
			return false, err
		}
		break
	}
	err = rows.Close()
	if err != nil {
		return exists, err
	}
	err = rows.Err()
	if err != nil {
		return exists, err
	}
	return exists, nil
}
