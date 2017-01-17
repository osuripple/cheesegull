package redis

import (
	"fmt"
	"time"
)

func (i *impl) GetSecurityKey(def string) string {
	k := i.Get("system:security_key").Val()
	if k == "" {
		i.Set("system:security_key", def, 0)
		return def
	}
	return k
}

func (i *impl) UpdateInLast5Minutes(b int) (bool, error) {
	v := i.Exists(fmt.Sprintf("updates:5m:%d", b))
	if v.Err() != nil {
		return false, v.Err()
	}
	if !v.Val() {
		i.Set(fmt.Sprintf("updates:5m:%d", b), nil, time.Minute*5)
	}
	return v.Val(), nil
}
