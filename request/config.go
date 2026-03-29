package request

type RequestConfig struct {
	categoryId string
	mp         map[string]string
}

func NewRequestConfig() *RequestConfig {
	return &RequestConfig{
		categoryId: "69c637e22cdb56c2fcc5f0df",
		mp: map[string]string{
			"a1": "69c6380e2cdb56c2fcc5f0ed",
			"a2": "69c6533bb47417171cb4642a",
			"b1": "69c65362b47417171cb4643d",
			"b2": "69c65376b47417171cb46443",
			"c1": "69c65396b47417171cb46449",
			"c2": "69c653adb47417171cb4644f",
		},
	}
}

func (rc *RequestConfig) GetCategoryId() string {
	return rc.categoryId
}

func (rc *RequestConfig) GetSeriesId(key string) (string, bool) {
	if val, ok := rc.mp[key]; ok {
		return val, true
	}
	return "", false
}

func (rc *RequestConfig) GetMp() map[string]string {
	return rc.mp
}
