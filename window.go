package sq

import "bytes"

type Window struct {
	PartitionByFields Fields
	OrderByFields     Fields
	FrameDefinition   string
}

var _ SQLAppender = Window{}

func (w Window) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	buf.WriteString("(")
	var written bool
	if len(w.PartitionByFields) > 0 {
		written = true
		buf.WriteString("PARTITION BY ")
		w.PartitionByFields.AppendSQLExclude(dialect, buf, args, params, nil, nil)
	}
	if len(w.OrderByFields) > 0 {
		if written {
			buf.WriteString(" ")
		}
		written = true
		buf.WriteString("ORDER BY ")
		w.OrderByFields.AppendSQLExclude(dialect, buf, args, params, nil, nil)
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

func CountOver(field Field, window Window) NumberField {
	return NumberFieldf("COUNT({}) OVER {}", field, window)
}

func CountAllOver(window Window) NumberField {
	return NumberFieldf("COUNT(*) OVER {}", window)
}

func SumOver(field Field, window Window) NumberField {
	return NumberFieldf("SUM({}) OVER {}", field, window)
}

func AvgOver(field Field, window Window) NumberField {
	return NumberFieldf("AVG({}) OVER {}", field, window)
}

func MinOver(field Field, window Window) NumberField {
	return NumberFieldf("MIN({}) OVER {}", field, window)
}

func MaxOver(field Field, window Window) NumberField {
	return NumberFieldf("MAX({}) OVER {}", field, window)
}

func RowNumberOver(window Window) NumberField {
	return NumberFieldf("ROW_NUMBER() OVER {}", window)
}

func RankOver(window Window) NumberField {
	return NumberFieldf("RANK() OVER {}", window)
}

func DenseRankOver(window Window) NumberField {
	return NumberFieldf("DENSE_RANK() OVER {}", window)
}

func PercentRankOver(window Window) NumberField {
	return NumberFieldf("PERCENT_RANK() OVER {}", window)
}

func CumeDistOver(window Window) NumberField {
	return NumberFieldf("CUME_DIST() OVER {}", window)
}

func LeadOver(field Field, offset int, fallback interface{}, window Window) CustomField {
	return Fieldf("LEAD({}, {}, {}) OVER {}", field, offset, fallback, window)
}

func LagOver(field Field, offset int, fallback interface{}, window Window) CustomField {
	return Fieldf("LAG({}, {}, {}) OVER {}", field, offset, fallback, window)
}

func NtileOver(n int, window Window) NumberField {
	return NumberFieldf("NTILE({}) OVER {}", n, window)
}

func FirstValueOver(field Field, window Window) CustomField {
	return Fieldf("FIRST_VALUE({}) OVER {}", field, window)
}

func LastValueOver(field Field, window Window) CustomField {
	return Fieldf("LAST_VALUE({}) OVER {}", field, window)
}

func NthValueOver(field Field, n int, window Window) CustomField {
	return Fieldf("NTH_VALUE({}, {}) OVER {}", field, n, window)
}

func Count(field Field) NumberField { return NumberFieldf("COUNT({})", field) }

func CountAll() NumberField { return NumberFieldf("COUNT(*)") }

func Sum(field Field) NumberField { return NumberFieldf("SUM({})", field) }

func Avg(field Field) NumberField { return NumberFieldf("AVG({})", field) }

func Min(field Field) NumberField { return NumberFieldf("MIN({})", field) }

func Max(field Field) NumberField { return NumberFieldf("MAX({})", field) }
