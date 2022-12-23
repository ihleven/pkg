package log

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalPlain *zap.Logger

var globalSugar *zap.SugaredLogger

func init() {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	writeSyncer := zapcore.AddSync(colorable.NewColorableStdout())
	levelEnabler := zapcore.DebugLevel // zap.NewAtomicLevelAt(zapcore.DebugLevel)

	core := zapcore.NewCore(encoder, writeSyncer, levelEnabler)

	globalPlain = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	globalSugar = globalPlain.Sugar()

}

func NewZapLogger() *zap.Logger {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "msg",

			LevelKey:    "lvl",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "ts",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,

			EncodeDuration: zapcore.StringDurationEncoder,
			// EncodeDuration: zapcore.MillisDurationEncoder,
		},
	}

	l, err := cfg.Build()
	if err != nil {
		return nil
	}

	return l
}
