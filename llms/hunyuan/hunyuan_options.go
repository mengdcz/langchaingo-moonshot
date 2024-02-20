package hunyuan

const (
	huanyuanAppId     = "HUANYUAN_APP_ID"
	huanyuanSecretId  = "HUANYUAN_SECRET_ID"  //nolint:gosec
	huanyuanSecretKey = "HUANYUAN_SECRET_KEY" //nolint:gosec
)

const (
	ModelNameHuanYuanPro = "hunyuan-pro"
	defaultModelName     = ModelNameHuanYuanPro
)

type options struct {
	appId     int64
	secretId  string
	secretKey string
	modelName string
}

type Option func(*options)

func WithAppId(appId int64) Option {
	return func(o *options) {
		o.appId = appId
	}
}

func WithSecretId(secretId string) Option {
	return func(o *options) {
		o.secretId = secretId
	}
}

func WithSecretKey(secretKey string) Option {
	return func(o *options) {
		o.secretKey = secretKey
	}
}

func WithModelName(modelName string) Option {
	return func(o *options) {
		o.modelName = modelName
	}
}
