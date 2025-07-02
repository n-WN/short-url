package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	// Base62 字符集
	base62Chars       = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	defaultCodeLength = 6
)

// Base62Encoder Base62 编码器
type Base62Encoder struct {
	chars      string
	base       int64
	codeLength int
}

// NewBase62Encoder 创建新的 Base62 编码器
func NewBase62Encoder() *Base62Encoder {
	return &Base62Encoder{
		chars:      base62Chars,
		base:       int64(len(base62Chars)),
		codeLength: defaultCodeLength,
	}
}

// SetCodeLength 设置生成的短码长度
func (e *Base62Encoder) SetCodeLength(length int) {
	if length > 0 {
		e.codeLength = length
	}
}

// Encode 将数字编码为 Base62 字符串
func (e *Base62Encoder) Encode(num int64) string {
	if num == 0 {
		return string(e.chars[0])
	}

	var result strings.Builder
	for num > 0 {
		result.WriteByte(e.chars[num%e.base])
		num /= e.base
	}

	// 反转字符串
	encoded := result.String()
	return reverseString(encoded)
}

// Decode 将 Base62 字符串解码为数字
func (e *Base62Encoder) Decode(encoded string) int64 {
	var result int64
	power := int64(1)

	for i := len(encoded) - 1; i >= 0; i-- {
		char := encoded[i]
		index := strings.IndexByte(e.chars, char)
		if index == -1 {
			return 0 // 无效字符
		}
		result += int64(index) * power
		power *= e.base
	}

	return result
}

// GenerateRandomCode 生成指定长度的随机短码
func (e *Base62Encoder) GenerateRandomCode() (string, error) {
	code := make([]byte, e.codeLength)
	for i := range code {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(e.base))
		if err != nil {
			return "", err
		}
		code[i] = e.chars[randomIndex.Int64()]
	}
	return string(code), nil
}

// GenerateCodeFromID 从 ID 生成短码，如果长度不够则填充随机字符
func (e *Base62Encoder) GenerateCodeFromID(id int64) (string, error) {
	encoded := e.Encode(id)

	// 如果编码后的长度不够，前面填充随机字符
	if len(encoded) < e.codeLength {
		padding := e.codeLength - len(encoded)
		randomPart := make([]byte, padding)

		for i := range randomPart {
			randomIndex, err := rand.Int(rand.Reader, big.NewInt(e.base))
			if err != nil {
				return "", err
			}
			randomPart[i] = e.chars[randomIndex.Int64()]
		}

		encoded = string(randomPart) + encoded
	}

	return encoded, nil
}

// IsValidCode 检查短码是否只包含有效字符
func (e *Base62Encoder) IsValidCode(code string) bool {
	for _, char := range code {
		if !strings.ContainsRune(e.chars, char) {
			return false
		}
	}
	return true
}

// reverseString 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
