package wrapper

import "time"

type IDGenerator func() string

type DateGenerator func() time.Time
