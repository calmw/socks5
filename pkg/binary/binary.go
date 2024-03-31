package binary

import (
	"bytes"
	"encoding/binary"
	"log"
)

func ReadInt32FromBinary(b []byte) (int32, error) {
	reader := bytes.NewReader(b)
	var number int32

	err := binary.Read(reader, binary.LittleEndian, &number)
	if err != nil {
		log.Println("Error reading from bytes:", err)
		return 0, err
	}

	return number, nil
}
