package common

import (
	"github.com/bits-and-blooms/bitset"
	"sync"
)

const defaultSize = 16                    // 设置哈希数组大小为16
const hashFuncCount = 6                   // 设置哈希函数个数
var seeds = []uint{7, 11, 13, 31, 37, 61} //设置种子

var BloomFilterService *BloomFilter

func init() {
	BloomFilterService = NewBloomFilter()
}

type BloomFilter struct {
	set      *bitset.BitSet //使用第三方库
	hashFunc [hashFuncCount]func(seed uint, value string) uint
	mux      sync.Mutex
}

// NewBloomFilter 构造一个布隆过滤器
func NewBloomFilter() *BloomFilter {
	bf := new(BloomFilter)
	bf.set = bitset.New(defaultSize)
	bf.mux = sync.Mutex{}
	for i := 0; i < len(bf.hashFunc); i++ {
		bf.hashFunc[i] = createHash()
	}
	return bf
}

//构造哈希函数，每个哈希函数有参数seed保证计算方式的不同
func createHash() func(seed uint, value string) uint {
	return func(seed uint, value string) uint {
		var result uint = 0
		for i := 0; i < len(value); i++ {
			result = result*seed + uint(value[i])
		}
		return result & (defaultSize - 1)
	}
}

// Add 添加元素
func (b *BloomFilter) Add(value string) {
	b.mux.Lock()
	defer b.mux.Unlock()
	for i, f := range b.hashFunc {
		b.set.Set(f(seeds[i], value))
	}
}

// Contains 判断元素是否存在
func (b *BloomFilter) Contains(value string) bool {
	for i, f := range b.hashFunc {
		if !b.set.Test(f(seeds[i], value)) {
			return false
		}
	}
	return true
}
