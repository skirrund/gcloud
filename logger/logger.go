package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	rotatelogs "github.com/skirrund/gcloud/logger/rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

var sLogger *slog.Logger
var zapL *zap.Logger
var zapLS *zap.SugaredLogger

var once sync.Once

const (
	DEFAULT_FILE = "log.log"
)

func init() {
	slog.Info("[Logger] init default....")
	encoder := getEncoder()
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.DebugLevel
	})
	c := zapcore.AddSync(os.Stderr)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, c, infoLevel),
	)
	zapL = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	zapLS = zapL.Sugar()
	sLogger = slog.New(zapslog.NewHandler(zapL.Core(), zapslog.WithCaller(true)))
	slog.SetDefault(sLogger)
}

// getMessage format with Sprint, Sprintf, or neither.
func GetMessage(template string, fmtArgs []any) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

func Error(args ...any) {
	zapLS.Error(args...)
}

func Fatal(args ...any) {
	zapLS.Fatal(args...)
}

func Infof(template string, args ...any) {
	zapLS.Infof(template, args...)
}

func Errorf(template string, args ...any) {
	zapLS.Errorf(template, args...)
}

func Sync() {
	zapL.Sync()
}

func Warn(args ...any) {
	zapLS.Warn(args...)
}

func Warnf(template string, args ...any) {
	zapLS.Warnf(template, args...)
}

func Panic(args ...any) {
	zapLS.Panic(args...)
}

func Info(args ...any) {
	zapLS.Info(args...)
}

func GetLogStr(needLog string) string {
	if len(needLog) > 1000 {
		return needLog[:1000]
	}
	return needLog
}

func getEncoder() zapcore.Encoder {
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalColorLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
	return encoder
}

func getJSONEncoder(service string) zapcore.Encoder {
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "rest",
		LevelKey:    "severity",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "@timestamp",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
	encoder.AddString("service", service)
	return encoder
}

func initLog(fileDir string, serviceName string, port string, console bool, json bool, maxAge time.Duration) *zap.Logger {
	encoder := getEncoder()
	jsonEncoder := getJSONEncoder(serviceName)
	// 实现两个判断日志等级的interface (其实 zapcore.*Level 自身就是 interface)
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.DebugLevel
	})
	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getWriter(fileDir, serviceName, port, maxAge)
	//	warnWriter := getWriter("log/log.log")
	jsonWriter := getWriterJSON(fileDir, serviceName, port)
	jWriter := zapcore.AddSync(jsonWriter)
	writer := zapcore.AddSync(infoWriter)
	var core zapcore.Core
	if console {
		c := zapcore.AddSync(os.Stdout)
		if json {
			core = zapcore.NewTee(
				zapcore.NewCore(encoder, writer, infoLevel),
				zapcore.NewCore(encoder, c, infoLevel),
				zapcore.NewCore(jsonEncoder, jWriter, infoLevel),
			)
		} else {
			core = zapcore.NewTee(
				zapcore.NewCore(encoder, writer, infoLevel),
				zapcore.NewCore(encoder, c, infoLevel),
			)
		}
	} else {
		if json {
			core = zapcore.NewTee(
				zapcore.NewCore(encoder, writer, infoLevel),
				zapcore.NewCore(jsonEncoder, jWriter, infoLevel),
			)
		} else {
			core = zapcore.NewTee(
				zapcore.NewCore(encoder, writer, infoLevel),
			)
		}
	}
	//core = core.With([]zapcore.Field{zapcore.Field{Key: "service", Type: zapcore.StringType, String: service}})
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func NewLogInstance(fileDir string, serviceName string, port string, console bool, json bool, maxAgeDay uint64) *slog.Logger {
	var logger *slog.Logger
	if maxAgeDay == 0 {
		maxAgeDay = 7
	}
	z := initLog(fileDir, serviceName, port, console, json, time.Duration(maxAgeDay)*time.Hour*24)
	logger = slog.New(zapslog.NewHandler(z.Core(), zapslog.WithCaller(true)))
	return logger
}

func InitLog(fileDir string, serviceName string, port string, console bool, json bool, maxAgeDay uint64) {
	once.Do(func() {
		if maxAgeDay == 0 {
			maxAgeDay = 7
		}
		zapL = initLog(fileDir, serviceName, port, console, json, time.Duration(maxAgeDay)*time.Hour*24)
		zapLS = zapL.Sugar()
		sLogger = slog.New(zapslog.NewHandler(zapL.Core(), zapslog.WithCaller(true)))
		slog.SetDefault(sLogger)
	})
}

func getFileName(serviceName string, port string) string {
	host, _ := os.Hostname()
	return "/" + serviceName + "/" + host + "-" + port // ".log.%Y-%m-%d"
}

func getWriter(fileDir string, serviceName string, port string, maxAgeDay time.Duration) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	//logFile := localApp.BootOptions.LoggerDir + "/" + localApp.BootOptions.ServerName + "-" + localApp.BootOptions.Host + "-" + strconv.FormatUint(localApp.BootOptions.ServerPort, 10) + ".log.%Y-%m-%d"
	fileName := getFileName(serviceName, port)
	p := fileDir + fileName + ".log.%Y-%m-%d"
	slog.Info("[logger]start init textlogger file:" + p)
	hook, err := rotatelogs.New(
		p, // 没有使用go风格反人类的format格式%Y-%m-%d-%H
		//rotatelogs.WithLinkName(fileDir+fileName+".log"),
		rotatelogs.WithMaxAge(maxAgeDay),
		rotatelogs.WithRotationTime(time.Hour),
		rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(e rotatelogs.Event) {
			if e.Type() != rotatelogs.FileRotatedEventType {
				return
			}
		})),
	)

	if err != nil {
		panic(err)
	}
	return hook
}

func getWriterJSON(fileDir string, serviceName string, port string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	//logFile := localApp.BootOptions.LoggerDir + "/" + localApp.BootOptions.ServerName + "-" + localApp.BootOptions.Host + "-" + strconv.FormatUint(localApp.BootOptions.ServerPort, 10) + ".log.%Y-%m-%d"
	fileName := getFileName(serviceName, port)
	p := fileDir + fileName + ".%Y-%m-%d.json"
	slog.Info("[logger]start init jsonlogger file:" + p)
	hook, err := rotatelogs.New(
		p, // 没有使用go风格反人类的format格式%Y-%m-%d-%H
		//rotatelogs.WithLinkName(fileDir+fileName+".json"),
		rotatelogs.WithMaxAge(time.Hour*24*3),
		rotatelogs.WithRotationTime(time.Hour),
		rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(e rotatelogs.Event) {
			if e.Type() != rotatelogs.FileRotatedEventType {
				return
			}
		})),
	)

	if err != nil {
		panic(err)
	}
	return hook
}
