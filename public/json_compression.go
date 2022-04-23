package public

import (
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

// JsonCompress 压缩JSON字符串
func JsonCompress(src []byte) []byte {
	var encoder, _ = zstd.NewWriter(nil)
	return encoder.EncodeAll(src, make([]byte, 0, len(src)))
}

// JsonDecompress 解压JSON字符串
func JsonDecompress(src []byte) ([]byte, error) {
	var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
	return decoder.DecodeAll(src, nil)
}

// JsonDeTool 常用的工具
func JsonDeTool(jsonStr string) ([]byte, error) {
	decompress, err := JsonDecompress(Base64Decoding(jsonStr))
	if err != nil {
		return []byte{}, errors.New("解压错误")
	}
	return decompress, nil
}

func JsonCoTool(strInfo string) string {
	return Base64Encoding(JsonCompress([]byte(strInfo)))
}
