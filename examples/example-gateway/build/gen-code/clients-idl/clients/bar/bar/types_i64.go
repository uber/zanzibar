package bar

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (v Long) MarshalJSON() ([]byte, error) {
	byteArray := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(byteArray, uint64(v))
	high := int32(binary.BigEndian.Uint32(byteArray[:4]))
	low := int32(binary.BigEndian.Uint32(byteArray[4:]))
	return ([]byte)(fmt.Sprintf("{\"high\":%d,\"low\":%d}", high, low)), nil
}
func (v *Long) UnmarshalJSON(text []byte) error {
	firstByte := text[0]
	if firstByte == byte('{') {
		result := map[string]int32{}
		err := json.Unmarshal(text, &result)
		if err != nil {
			return err
		}
		byteArray := make([]byte, 8, 8)
		binary.BigEndian.PutUint32(byteArray[:4], uint32(result["high"]))
		binary.BigEndian.PutUint32(byteArray[4:], uint32(result["low"]))
		x := binary.BigEndian.Uint64(byteArray)
		*v = Long(int64(x))
	} else {
		x, err := strconv.ParseInt(string(text), 10, 64)
		if err != nil {
			return err
		}
		*v = Long(x)
	}
	return nil
}

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
