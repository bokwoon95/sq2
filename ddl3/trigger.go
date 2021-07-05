package ddl3

import "io"

type TriggerSource interface {
	GetSource() io.Reader
}
