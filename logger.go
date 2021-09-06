package sq

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const (
	Linterpolate        = 1 << iota // Interpolate the args into the query
	LinterpolateVerbose             // Show the query before and after interpolation
	Lcaller                         // Show caller information i.e. filename, line number, function name
	Lcolor                          // Colorize log output
	Ltime                           // Show time taken
	Lasync                          // log asynchronously
)

var (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorBlue   = "\x1b[34m"
	colorPurple = "\x1b[35m"
	colorCyan   = "\x1b[36m"
	colorGray   = "\x1b[37m"
	colorWhite  = "\x1b[97m"
)

func init() {
	if runtime.GOOS == "windows" {
		colorReset = ""
		colorRed = ""
		colorGreen = ""
		colorYellow = ""
		colorBlue = ""
		colorPurple = ""
		colorCyan = ""
		colorGray = ""
		colorWhite = ""
	}
}

type QueryStats struct {
	Env     map[string]interface{}
	Dialect string

	// TODO: Rethink whether I need these fields. Do I really want post-query
	// auditing code to be put inside the logger function?
	QueryType     string
	TableModified [2]string

	Query          string
	Args           []interface{}
	Error          error
	RowCount       sql.NullInt64
	RowsAffected   sql.NullInt64
	LastInsertID   sql.NullInt64
	Exists         sql.NullBool
	TimeTaken      time.Duration
	QueryResults   string
	CallerFile     string
	CallerLine     int
	CallerFunction string
}

type LogSettings struct {
	ResultsLimit  int
	GetCallerInfo bool
	AsyncLogging  bool
	TimeQuery     bool
}

type Logger interface {
	LogQueryStats(ctx context.Context, stats QueryStats)
	GetLogSettings() (LogSettings, error)
}

var ErrLoggerUnsupported = errors.New("Logger is not supported")

type LoggerDB interface {
	Logger
	DB
}

type loggerDB struct {
	Logger
	DB
}

type logger struct {
	logger       *log.Logger
	logflag      int
	resultsLimit int
}

func NewLogger(out io.Writer, logflag int, resultsLimit int) Logger {
	return logger{
		logger:       log.New(os.Stdout, "", log.LstdFlags),
		logflag:      logflag,
		resultsLimit: resultsLimit,
	}
}

var (
	defaultLogger = NewLogger(os.Stdout, Ltime|Lcaller|Lcolor|Linterpolate, 0)
	verboseLogger = NewLogger(os.Stdout, Ltime|Lcaller|Lcolor|LinterpolateVerbose, 5)
)

func Log(db DB) LoggerDB {
	return loggerDB{Logger: defaultLogger, DB: db}
}

func VerboseLog(db DB) LoggerDB {
	return loggerDB{Logger: verboseLogger, DB: db}
}

func (l logger) GetLogSettings() (LogSettings, error) {
	return LogSettings{
		ResultsLimit:  l.resultsLimit,
		GetCallerInfo: Lcaller&l.logflag != 0,
		AsyncLogging:  Lasync&l.logflag != 0,
		TimeQuery:     Ltime&l.logflag != 0,
	}, nil
}

func (l logger) LogQueryStats(ctx context.Context, stats QueryStats) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	var reset, red, green, blue, purple string
	if Lcolor&l.logflag != 0 {
		reset = colorReset
		red = colorRed
		green = colorGreen
		blue = colorBlue
		purple = colorPurple
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	if stats.Error == nil {
		buf.WriteString(green + "[OK]" + reset)
	} else {
		buf.WriteString(red + "[FAIL]" + reset)
	}
	if LinterpolateVerbose&l.logflag == 0 {
		if Linterpolate&l.logflag != 0 {
			query, err := Sprintf(stats.Dialect, stats.Query, stats.Args)
			if err != nil {
				query += " " + err.Error()
			}
			buf.WriteString(" " + query)
		} else {
			buf.WriteString(" " + stats.Query + " " + fmt.Sprint(stats.Args))
		}
		buf.WriteString(" |")
	}
	if stats.TimeTaken > 0 {
		buf.WriteString(blue + " timeTaken" + reset + "=" + stats.TimeTaken.String())
	}
	if stats.Exists.Valid {
		buf.WriteString(blue + " exists" + reset + "=" + strconv.FormatBool(stats.Exists.Bool))
	}
	if stats.RowCount.Valid {
		buf.WriteString(blue + " rowCount" + reset + "=" + strconv.FormatInt(stats.RowCount.Int64, 10))
	}
	if stats.RowsAffected.Valid {
		buf.WriteString(blue + " rowsAffected" + reset + "=" + strconv.FormatInt(stats.RowsAffected.Int64, 10))
	}
	if stats.LastInsertID.Valid {
		buf.WriteString(blue + " lastInsertID" + reset + "=" + strconv.FormatInt(stats.LastInsertID.Int64, 10))
	}
	if Lcaller&l.logflag != 0 {
		buf.WriteString(blue + " caller" + reset + "=" + stats.CallerFile + ":" + strconv.Itoa(stats.CallerLine) + ":" + filepath.Base(stats.CallerFunction))
	}
	if LinterpolateVerbose&l.logflag != 0 {
		buf.WriteString("\n" + purple + "----[ Executing query ]----" + reset)
		buf.WriteString("\n" + stats.Query + " " + fmt.Sprintf("%#v", stats.Args))
		buf.WriteString("\n" + purple + "----[ with bind values ]----" + reset)
		query, err := Sprintf(stats.Dialect, stats.Query, stats.Args)
		if err != nil {
			query += " " + err.Error()
		}
		buf.WriteString("\n" + query)
	}
	if stats.QueryResults != "" {
		buf.WriteString("\n" + purple + "----[ Fetched result ]----" + reset)
		buf.WriteString(stats.QueryResults)
	}
	if buf.Len() > 0 {
		l.logger.Println(buf.String())
	}
}
