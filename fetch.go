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
	var resultsLimit int
	var logQueryStats func(ctx context.Context, stats QueryStats, skip int)
	if db, ok := db.(LoggerDB); ok {
		logQueryStats = db.LogQueryStats
		resultsLimit = db.LimitResults()
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
		if stats.Query == "" && err != nil {
			stats.Query = buf.String() + "%!(error=" + err.Error() + ")"
		}
		if resultsLimit > 0 {
			stats.QueryResults = resultsBuf.String()
		}
		buf.Reset()
		resultsBuf.Reset()
		bufpool.Put(buf)
		bufpool.Put(resultsBuf)
		if logQueryStats == nil {
			return
		}
		stats.Error = err
		stats.RowCount.Valid = true
		stats.RowCount.Int64 = rowCount
		logQueryStats(ctx, stats, skip+1)
	}()
	err = q.AppendSQL(stats.Dialect, buf, &stats.Args, make(map[string][]int), nil)
	if err != nil {
		return 0, err
	}
	stats.Query = buf.String()
	start := time.Now()
	rows, err := db.QueryContext(ctx, stats.Query, stats.Args...)
	stats.TimeTaken = time.Since(start)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if len(dest) == 0 {
		return 0, nil
	}
	RowActivate(r)
	for rows.Next() {
		rowCount++
		// Because dest and r.dest share the same backing array, any change
		// that rows.Scan() makes to dest will be propagated to r.dest and
		// hence be visible in the user's rowmapper. Hooray Go slices!
		err = rows.Scan(dest...)
		if err != nil {
			return rowCount, decorateScanError(stats.Dialect, fields, dest, err)
		}
		if resultsLimit > 0 && rowCount <= int64(resultsLimit) {
			accumulateResults(stats.Dialect, resultsBuf, fields, dest, rowCount)
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
	if resultsLimit > 0 && rowCount > int64(resultsLimit) {
		resultsBuf.WriteString("\n...\n(" + strconv.FormatInt(rowCount-int64(resultsLimit), 10) + " more rows)")
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

func accumulateResults(dialect string, buf *bytes.Buffer, fields []Field, dest []interface{}, rowNumber int64) {
	tmpbuf := bufpool.Get().(*bytes.Buffer)
	tmpargs := argspool.Get().([]interface{})
	defer func() {
		tmpbuf.Reset()
		tmpargs = tmpargs[:0]
		bufpool.Put(tmpbuf)
		argspool.Put(tmpargs)
	}()
	buf.WriteString("\n----[ Row " + strconv.FormatInt(rowNumber, 10) + " ]----")
	for i := range dest {
		buf.WriteString("\n")
		tmpbuf.Reset()
		tmpargs = tmpargs[:0]
		err := fields[i].AppendSQLExclude(dialect, tmpbuf, &tmpargs, make(map[string][]int), nil, nil)
		if err != nil {
			buf.WriteString("%!(error=" + err.Error() + ")")
			continue
		}
		lhs, err := Sprintf(dialect, tmpbuf.String(), tmpargs)
		if err != nil {
			buf.WriteString("%!(error=" + err.Error() + ")")
			continue
		}
		buf.WriteString(lhs + ": ")
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
	var logQueryStats func(ctx context.Context, stats QueryStats, skip int)
	if db, ok := db.(LoggerDB); ok {
		logQueryStats = db.LogQueryStats
	}
	stats.Dialect = q.GetDialect()
	q, err = q.SetFetchableFields([]Field{Literal("1")})
	if err != nil {
		return false, err
	}
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
		if exists {
			stats.RowCount.Valid = true
			stats.RowCount.Int64 = 1
		}
		logQueryStats(ctx, stats, skip+1)
	}()
	buf.WriteString("SELECT EXISTS(")
	err = q.AppendSQL(stats.Dialect, buf, &stats.Args, make(map[string][]int), nil)
	if err != nil {
		return false, err
	}
	buf.WriteString(")")
	stats.Query = buf.String()
	start := time.Now()
	rows, err := db.QueryContext(ctx, stats.Query, stats.Args...)
	stats.TimeTaken = time.Since(start)
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
