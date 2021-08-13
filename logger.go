package sq

import (
	"bytes"
	"context"
	"database/sql"
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
	// TODO: bruh I need a better naming system than Linterpolate vs Lbeforeafter
	Linterpolate = 0b1     // Interpolate the args into the query
	Lbeforeafter = 0b10    // Show the query before and after interpolation
	Lcaller      = 0b100   // Show caller information i.e. filename, line number, function name
	Lresults     = 0b1000  // Show the first 5 results if applicable. Lmultiline must be enabled.
	Lcolor       = 0b10000 // Colorize log output
)

const (
	ExecActive    = 0b1   // Used by Logger to discern between Fetch and Exec queries
	ElastInsertID = 0b10  // Get last inserted ID
	ErowsAffected = 0b100 // Get number of rows affected
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
	Dialect        string
	Query          string
	Args           []interface{}
	Error          error
	RowCount       sql.NullInt64
	RowsAffected   sql.NullInt64
	LastInsertID   sql.NullInt64
	TimeTaken      time.Duration
	QueryResults   string
	CallerFile     string
	CallerLine     int
	CallerFunction string
}

type LogSettings struct {
	ResultsLimit  int
	GetCallerInfo bool
}

type Logger interface {
	LogQueryStats(ctx context.Context, stats QueryStats, skip int)
	LogResults() (shouldLogResults bool, limit int)
	GetLogSettings() LogSettings
}

type LoggerDB struct {
	Logger
	DB
}

type logger struct {
	log          *log.Logger
	logflag      int
	resultsLimit int
}

func NewLogger(out io.Writer, logflag int, resultsLimit int) Logger {
	return logger{
		log:          log.New(os.Stdout, "", log.LstdFlags),
		logflag:      logflag,
		resultsLimit: resultsLimit,
	}
}

var (
	defaultLogger = NewLogger(os.Stdout, Linterpolate|Lcaller|Lcolor, 5) // Lcaller rationale: logging is for debugging, so we should provide caller info by default
	verboseLogger = NewLogger(os.Stdout, Lbeforeafter|Lcaller|Lcolor|Lresults, 5)
)

func Log(db DB) LoggerDB {
	return LoggerDB{Logger: defaultLogger, DB: db}
}

func VerboseLog(db DB) LoggerDB {
	return LoggerDB{Logger: verboseLogger, DB: db}
}

func (l logger) LogResults() (shouldLogResults bool, limit int) { return true, l.resultsLimit }

func (l logger) GetLogSettings() LogSettings {
	return LogSettings{
		ResultsLimit:  l.resultsLimit,
		GetCallerInfo: l.logflag&Lcaller != 0,
	}
}

func (l logger) LogQueryStats(ctx context.Context, stats QueryStats, skip int) {
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
	if Lbeforeafter&l.logflag == 0 {
		// Log one-liner
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
	buf.WriteString(blue + " timeTaken" + reset + "=" + stats.TimeTaken.String())
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
		file, line, function := caller(skip + 1)
		buf.WriteString(blue + " caller" + reset + "=" + file + ":" + strconv.Itoa(line) + ":" + filepath.Base(function))
	}
	if Lbeforeafter&l.logflag != 0 {
		// Log multiline
		buf.WriteString("\n" + purple + "----[ Executing query ]----" + reset)
		buf.WriteString("\n" + stats.Query + " " + fmt.Sprint(stats.Args))
		buf.WriteString("\n" + purple + "----[ with bind values ]----" + reset)
		query, err := Sprintf(stats.Dialect, stats.Query, stats.Args)
		if err != nil {
			query += " " + err.Error()
		}
		buf.WriteString("\n" + query)
	}
	if Lresults&l.logflag != 0 && stats.QueryResults != "" {
		buf.WriteString("\n" + purple + "----[ Fetched result ]----" + reset)
		buf.WriteString(stats.QueryResults)
	}
	if buf.Len() > 0 {
		l.log.Printf(buf.String())
	}
}
