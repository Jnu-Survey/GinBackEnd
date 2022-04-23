package public

import "github.com/klauspost/compress/zstd"

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