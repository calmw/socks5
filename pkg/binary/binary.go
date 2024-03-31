package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func ReadInt32FromBinary(b []byte) (int64, error) {
	// 假设我们有一个字节数组表示一个整数
	//byteArray := []byte{0x00, 0x00, 0x01, 0x02} // 对应整数258

	// 使用bytes.Buffer来提供缓冲区
	buf := bytes.NewBuffer(b)

	// 用于存储转换后的整数
	var number int64

	// 使用binary.Read从buf中读取一个int32类型的数据
	err := binary.Read(buf, binary.BigEndian, &number)
	if err != nil {
		fmt.Println("转换出错:", err)
	}

	fmt.Printf("转换后的整数为: %d\n", number)

	return number, nil
}
