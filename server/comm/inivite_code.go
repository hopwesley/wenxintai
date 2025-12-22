package comm

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
)

const (
	tier       = "B"                                // B=基础版, P=专业版, C=校园版
	expireDays = 30                                 // 0 表示不过期
	alphabet   = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Crockford Base32
)

// ================== 邀请码生成 ==================

func randBase32(n int) (string, error) {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			return "", err
		}
		out[i] = alphabet[int(b[0])&31]
	}
	return string(out), nil
}

func checksumChar(s string) byte {
	h := sha1.Sum([]byte(s))
	return alphabet[int(h[0])&31]
}

func MakeInviteCode() (string, error) {
	b1, err := randBase32(4)
	if err != nil {
		return "", err
	}
	b2, err := randBase32(4)
	if err != nil {
		return "", err
	}
	b3, err := randBase32(4)
	if err != nil {
		return "", err
	}

	body := fmt.Sprintf("%s-%s-%s-%s", tier, b1, b2, b3)
	c := checksumChar(body)
	return fmt.Sprintf("%s-%c", body, c), nil
}
