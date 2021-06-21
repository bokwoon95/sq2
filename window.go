package sq

import "bytes"

type Window struct {
	PartitionByFields Fields
	OrderByFields     Fields
	FrameDefinition   string
}

var _ SQLAppender = Window{}

func (w Window) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("(")
	var written bool
	if len(w.PartitionByFields) > 0 {
		written = true
		buf.WriteString("PARTITION BY ")
		w.PartitionByFields.AppendSQLExclude(dialect, buf, args, params, nil)
	}
	if len(w.OrderByFields) > 0 {
		if written {
			buf.WriteString(" ")
		}
		written = true
		buf.WriteString("ORDER BY ")
		w.OrderByFields.AppendSQLExclude(dialect, buf, args, params, nil)
	}
	if w.FrameDefinition != "" {
		if written {
			buf.WriteString(" ")
		}
		written = true
		buf.WriteString(w.FrameDefinition)
	}
	buf.WriteString(")")
	return nil
}

func PartitionBy(fields ...Field) Window {
	return Window{PartitionByFields: fields}
}

func OrderBy(fields ...Field) Window {
	return Window{OrderByFields: fields}
}

func (w Window) PartitionBy(fields ...Field) Window {
	w.PartitionByFields = fields
	return w
}

func (w Window) OrderBy(fields ...Field) Window {
	w.OrderByFields = fields
	return w
}

func (w Window) Frame(frameDefinition string) Window {
	w.FrameDefinition = frameDefinition
	return w
}
