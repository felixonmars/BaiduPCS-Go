package panhome

import (
	"github.com/iikira/Baidu-Login/bdcrypto"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
)

type (
	// SignRes 签名结果
	SignRes interface {
		Sign() string
		Timestamp() string
	}

	signRes struct {
		sign      string
		timestamp string
	}
)

func (sr *signRes) Sign() string {
	return sr.sign
}
func (sr *signRes) Timestamp() string {
	return sr.timestamp
}

func Sign2(j, r []rune) []byte {
	var (
		a  = make([]rune, 256)
		p  = make([]rune, 256)
		o  = make([]byte, len(r))
		v  = len(j)
		q  int
		u  rune
		i  int
		k  rune
		dr int
	)
	if v == 0 {
		return o
	}
	for ; q < 256; q++ {
		dr = q % v
		a[q] = j[dr : 1+dr][0]
		p[q] = rune(q)
	}
	for q = 0; q < 256; q++ {
		u = (u + p[q] + a[q]) % 256
		p[q], p[u] = p[u], p[q]
	}
	u = 0
	for q = 0; q < len(r); q++ {
		i = (i + 1) % 256
		u = (u + p[i]) % 256
		p[i], p[u] = p[u], p[i]
		k = p[(p[i]+p[u])%256]
		o[q] = byte(r[q] ^ k)
	}
	return o
}

func (ph *PanHome) Signature() (sign SignRes, err error) {
	err = ph.getSignInfo()
	if err != nil {
		return nil, err
	}

	o := Sign2(ph.sign3, ph.sign1)
	signed := bdcrypto.Base64Encode(o)
	return &signRes{
		sign:      converter.ToString(signed),
		timestamp: ph.timestamp,
	}, nil
}
