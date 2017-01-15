package redis

func (i *impl) GetSecurityKey(def string) string {
	k := i.Get("system:security_key").Val()
	if k == "" {
		i.Set("system:security_key", def, 0)
		return def
	}
	return k
}
