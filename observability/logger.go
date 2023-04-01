package observability

import (
	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {
	configBuilder := zap.NewProductionConfig()
	configBuilder.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config, err := configBuilder.Build()

	if err != nil {
		panic(err)
	}

	return config.Sugar()
}
