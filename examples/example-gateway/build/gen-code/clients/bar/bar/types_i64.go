package bar

import (
	"strconv"
	"time"
)

func (v Timestamp) MarshalJSON() ([]byte, error) {
	x := (int64)(v)
	return ([]byte)("\"" + time.Unix(x/1000, 0).UTC().Format(time.RFC3339) + "\""), nil
}

func (v *Timestamp) UnmarshalJSON(text []byte) error {
	firstByte := text[0]
	if firstByte == byte('"') {
		x, err := time.Parse(time.RFC3339, string(text[1:len(text)-1]))
		if err != nil {
			return err
		}
		*v = Timestamp(x.Unix() * 1000)
	} else {
		x, err := strconv.ParseInt(string(text), 10, 64)
		if err != nil {
			return err
		}
		*v = Timestamp(x)
	}
	return nil
}
