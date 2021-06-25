package sq

import (
	"fmt"
	"path/filepath"
	"time"
)

type ColumnMode int

const (
	ColumnModeInsert ColumnMode = 0
	ColumnModeUpdate ColumnMode = 1
)

type Column struct {
	// mode determines if INSERT or UPDATE
	mode ColumnMode
	// INSERT
	rowStarted    bool
	rowEnded      bool
	firstField    string
	insertColumns Fields
	rowValues     RowValues
	// UPDATE
	assignments Assignments
}

// NOTE: Why did I make the following three functions public functions? Is there some use case that I envisioned for the user?
// NOTE: Oh my god. I wrote those functions because I wanted to allow external packages to use the same *Column for their own functions. One year ago, I had already anticipated the same concerns I have today.
func NewColumn(mode ColumnMode) *Column {
	var col Column
	col.mode = mode
	return &col
}

func ColumnInsertResult(col *Column) (Fields, RowValues) {
	return col.insertColumns, col.rowValues
}

func ColumnUpdateResult(col *Column) Assignments {
	return col.assignments
}

func (col *Column) Set(field Field, value interface{}) {
	if field == nil {
		file, line, _ := caller(1)
		panic(fmt.Errorf("%s:%d: setting a nil field", filepath.Base(file), line))
	}
	// UPDATE mode
	if col.mode == ColumnModeUpdate {
		col.assignments = append(col.assignments, Assign(field, value))
		return
	}
	// INSERT mode
	name := field.GetName()
	if !col.rowStarted {
		col.rowStarted = true
		col.firstField = name
		col.insertColumns = append(col.insertColumns, field)
		col.rowValues = append(col.rowValues, RowValue{value})
		return
	}
	if col.rowStarted && name == col.firstField {
		if !col.rowEnded {
			col.rowEnded = true
		}
		// Start a new RowValue
		col.rowValues = append(col.rowValues, RowValue{value})
		return
	}
	if !col.rowEnded {
		col.insertColumns = append(col.insertColumns, field)
	}
	// Append to last RowValue
	last := len(col.rowValues) - 1
	col.rowValues[last] = append(col.rowValues[last], value)
}

func (col *Column) SetBool(field BooleanField, value bool) { col.Set(field, value) }

func (col *Column) SetFloat64(field NumberField, value float64) { col.Set(field, value) }

func (col *Column) SetInt(field NumberField, value int) { col.Set(field, value) }

func (col *Column) SetInt64(field NumberField, value int64) { col.Set(field, value) }

func (col *Column) SetString(field StringField, value string) { col.Set(field, value) }

func (col *Column) SetTime(field TimeField, value time.Time) { col.Set(field, value) }
