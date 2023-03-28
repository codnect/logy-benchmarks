package benchmarks

import (
	"context"
	"github.com/procyon-projects/logy"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"io"
	"log"
	"testing"

	"go.uber.org/zap"
)

type Syncer struct {
	err    error
	called bool
}

func (s *Syncer) SetError(err error) {
	s.err = err
}

func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

func (s *Syncer) Called() bool {
	return s.called
}

type Discarder struct{ Syncer }

func (d *Discarder) Write(b []byte) (int, error) {
	return io.Discard.Write(b)
}

type discardHandler struct {
	disabled bool
	r        slog.Record
	attrs    []slog.Attr
	groups   []string
}

func (d discardHandler) Enabled(slog.Level) bool { return !d.disabled }
func (d discardHandler) Handle(r slog.Record) error {
	d.r = r
	return nil
}
func (d discardHandler) WithAttrs(as []slog.Attr) slog.Handler {
	c2 := d
	c2.attrs = concat(c2.attrs, as)
	return &c2
}
func (d discardHandler) WithGroup(name string) slog.Handler {
	c2 := d
	c2.groups = append(slices.Clip(c2.groups), name)
	return &c2
}

func concat[T any](s1, s2 []T) []T {
	s := make([]T, len(s1)+len(s2))
	copy(s, s1)
	copy(s[len(s1):], s2)
	return s
}

func BenchmarkDisabledWithoutFields(b *testing.B) {
	b.Logf("Logging at a disabled level without any structured context.")
	b.Run("Logy", func(b *testing.B) {
		logger := logy.Get()
		_ = logy.LoadConfig(&logy.Config{Level: logy.LevelDebug, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard}})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Logy.Formatting", func(b *testing.B) {
		logger := logy.Get()
		_ = logy.LoadConfig(&logy.Config{Level: logy.LevelError, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard}})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info("{} {} {} {} {} {} {} {} {} {}", fakeFmtArgs()...)
			}
		})
	})
	b.Run("exp/slog", func(b *testing.B) {
		logger := slog.New(discardHandler{disabled: true})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.Check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
					m.Write()
				}
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.SugarFormatting", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
	b.Run("apex/log", func(b *testing.B) {
		logger := newDisabledApexLog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog.Formatting", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msgf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
}

func BenchmarkDisabledAccumulatedContext(b *testing.B) {
	b.Logf("Logging at a disabled level with some accumulated context.")
	ctx := logy.WithContextFields(context.Background())

	b.Run("Logy", func(b *testing.B) {
		logger := logy.Get()
		_ = logy.LoadConfig(&logy.Config{Level: logy.LevelError, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard}})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.I(ctx, getMessage(0))
			}
		})
	})
	b.Run("Logy.Formatting", func(b *testing.B) {
		logger := logy.Get()
		_ = logy.LoadConfig(&logy.Config{Level: logy.LevelError, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard}})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.I(ctx, "{} {} {} {} {} {} {} {} {} {}", fakeFmtArgs()...)
			}
		})
	})
	b.Run("exp/slog", func(b *testing.B) {
		logger := slog.New(discardHandler{disabled: true}).With(fakeFmtArgs()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.Check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
					m.Write()
				}
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeFields()...).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.SugarFormatting", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeFields()...).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
	b.Run("apex/log", func(b *testing.B) {
		logger := newDisabledApexLog().WithFields(fakeApexFields())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newDisabledZerolog().With()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog.Formatting", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msgf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
}

func BenchmarkDisabledAddingFields(b *testing.B) {
	b.Logf("Logging at a disabled level, adding context at each log site.")
	b.Run("Logy", func(b *testing.B) {
		logger := logy.Named("")
		logger.SetLevel(logy.LevelError)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0), fakeFields()...)
			}
		})
	})
	b.Run("Zap.Check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
					m.Write(fakeFields()...)
				}
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(getMessage(0), fakeSugarFields()...)
			}
		})
	})
	b.Run("apex/log", func(b *testing.B) {
		logger := newDisabledApexLog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeApexFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				fakeZerologFields(logger.Info()).Msg(getMessage(0))
			}
		})
	})
}

func BenchmarkWithoutFields(b *testing.B) {
	logger := logy.Get()
	_ = logy.LoadConfig(&logy.Config{Level: logy.LevelDebug, IncludeCaller: false, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard, Enabled: true, Format: "%d %p %c : %m%s%n", Json: &logy.JsonConfig{
		Enabled: true,
	}}})

	b.Logf("Logging without any structured context.")
	b.Run("Logy", func(b *testing.B) {

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Logy.Formatting", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info("{} {} {} {} {} {} {} {} {} {}", fakeFmtArgs()...)
			}
		})
	})
	b.Run("exp/slog", func(b *testing.B) {
		logger := slog.New(discardHandler{disabled: false})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Error(getMessage(0))
			}
		})
	})
	b.Run("Zap.Check", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if ce := logger.Check(zap.InfoLevel, getMessage(0)); ce != nil {
					ce.Write()
				}
			}
		})
	})
	b.Run("Zap.CheckSampled", func(b *testing.B) {
		logger := newSampledLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				if ce := logger.Check(zap.InfoLevel, getMessage(i)); ce != nil {
					ce.Write()
				}
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.SugarFormatting", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
	b.Run("apex/log", func(b *testing.B) {
		logger := newApexLog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("go-kit/kit/log", func(b *testing.B) {
		logger := newKitLog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Log(getMessage(0), getMessage(1))
			}
		})
	})
	b.Run("inconshreveable/log15", func(b *testing.B) {
		logger := newLog15()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("stdlib.Println", func(b *testing.B) {
		logger := log.New(io.Discard, "", log.LstdFlags)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Println(getMessage(0))
			}
		})
	})
	b.Run("stdlib.Printf", func(b *testing.B) {
		logger := log.New(io.Discard, "", log.LstdFlags)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Printf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog.Formatting", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msgf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
	b.Run("rs/zerolog.Check", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if e := logger.Info(); e.Enabled() {
					e.Msg(getMessage(0))
				}
			}
		})
	})
}

func BenchmarkWithContext(b *testing.B) {

	b.Logf("Logging with some accumulated context.")
	b.Run("Logy console", func(b *testing.B) {
		logger := logy.Get()
		_ = logy.LoadConfig(&logy.Config{Level: logy.LevelDebug, IncludeCaller: false, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard, Enabled: true}})

		ctx := logy.WithContextFields(context.Background())

		ctx = logy.WithValue(ctx, "int", _tenInts[0])
		ctx = logy.WithValue(ctx, "ints", _tenInts)
		ctx = logy.WithValue(ctx, "string", _tenStrings[0])
		ctx = logy.WithValue(ctx, "strings", _tenStrings)
		ctx = logy.WithValue(ctx, "time", _tenTimes[0])
		ctx = logy.WithValue(ctx, "times", _tenTimes)
		ctx = logy.WithValue(ctx, "user1", _oneUser)
		ctx = logy.WithValue(ctx, "user2", _oneUser)
		ctx = logy.WithValue(ctx, "users", _tenUsers)
		ctx = logy.WithValue(ctx, "error", errExample)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.I(ctx, getMessage(0))
			}
		})
	})

	b.Run("Logy", func(b *testing.B) {
		logger := logy.Get()
		_ = logy.LoadConfig(&logy.Config{Level: logy.LevelDebug, IncludeCaller: false, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard, Enabled: true, Format: "%d %p %c : %m%s%n", Json: &logy.JsonConfig{
			Enabled: true,
		}}})

		ctx := logy.WithContextFields(context.Background())

		ctx = logy.WithValue(ctx, "int", _tenInts[0])
		ctx = logy.WithValue(ctx, "ints", _tenInts)
		ctx = logy.WithValue(ctx, "string", _tenStrings[0])
		ctx = logy.WithValue(ctx, "strings", _tenStrings)
		ctx = logy.WithValue(ctx, "time", _tenTimes[0])
		ctx = logy.WithValue(ctx, "times", _tenTimes)
		ctx = logy.WithValue(ctx, "user1", _oneUser)
		ctx = logy.WithValue(ctx, "user2", _oneUser)
		ctx = logy.WithValue(ctx, "users", _tenUsers)
		ctx = logy.WithValue(ctx, "error", errExample)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.I(ctx, getMessage(0))
			}
		})
	})

	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
	b.Run("Logy.Formatting", func(b *testing.B) {
		logger := logy.Get()
		_ = logy.LoadConfig(&logy.Config{Level: logy.LevelDebug, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard, Enabled: true, Format: "%d %p %c : %m%s%n"}})

		ctx := logy.WithContextFields(context.Background())

		ctx = logy.WithValue(ctx, "int", _tenInts[0])
		ctx = logy.WithValue(ctx, "ints", _tenInts)
		ctx = logy.WithValue(ctx, "string", _tenStrings[0])
		ctx = logy.WithValue(ctx, "strings", _tenStrings)
		ctx = logy.WithValue(ctx, "time", _tenTimes[0])
		ctx = logy.WithValue(ctx, "times", _tenTimes)
		ctx = logy.WithValue(ctx, "user1", _oneUser)
		ctx = logy.WithValue(ctx, "user2", _oneUser)
		ctx = logy.WithValue(ctx, "users", _tenUsers)
		ctx = logy.WithValue(ctx, "error", errExample)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.I(ctx, "{} {} {} {} {} {} {} {} {} {}", fakeFmtArgs()...)
			}
		})
	})

	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.Check", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if ce := logger.Check(zap.InfoLevel, getMessage(0)); ce != nil {
					ce.Write()
				}
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.SugarFormatting", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
	b.Run("rs/zerolog.Formatting", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msgf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})

	b.Run("exp/slog", func(b *testing.B) {
		logger := slog.New(slog.NewTextHandler(io.Discard)).With("int", _tenInts[0], "ints", _tenInts, "string", _tenStrings[0], "strings", _tenStrings, "time", _tenTimes[0], "times", _tenTimes,
			"user1", _oneUser, "user2", _oneUser, "users", _tenUsers, "error", errExample)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.Check", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if ce := logger.Check(zap.InfoLevel, getMessage(0)); ce != nil {
					ce.Write()
				}
			}
		})
	})
	b.Run("Zap.CheckSampled", func(b *testing.B) {
		logger := newSampledLogger(zap.DebugLevel).With(fakeFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				if ce := logger.Check(zap.InfoLevel, getMessage(i)); ce != nil {
					ce.Write()
				}
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("Zap.SugarFormatting", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeFields()...).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
	b.Run("apex/log", func(b *testing.B) {
		logger := newApexLog().WithFields(fakeApexFields())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("go-kit/kit/log", func(b *testing.B) {
		logger := newKitLog(fakeSugarFields()...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Log(getMessage(0), getMessage(1))
			}
		})
	})
	b.Run("inconshreveable/log15", func(b *testing.B) {
		logger := newLog15().New(fakeSugarFields())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog.Check", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if e := logger.Info(); e.Enabled() {
					e.Msg(getMessage(0))
				}
			}
		})
	})
	b.Run("rs/zerolog.Formatting", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msgf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
	})
}

func BenchmarkAddingFields(b *testing.B) {
	logger := logy.Get()
	_ = logy.LoadConfig(&logy.Config{Level: logy.LevelDebug, Console: &logy.ConsoleConfig{Target: logy.TargetDiscard, Enabled: true, Format: "%d %p %x{int} %c : %m%s%n", Json: &logy.JsonConfig{
		Enabled: true,
	}}})

	b.Logf("Logging with additional context at each log site.")

	b.Run("Logy.Formatting", func(b *testing.B) {

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0), fakeFmtArgs()...)
			}
		})
	})

	b.Run("Zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0), fakeFields()...)
			}
		})
	})
	b.Run("Zap.Check", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if ce := logger.Check(zap.InfoLevel, getMessage(0)); ce != nil {
					ce.Write(fakeFields()...)
				}
			}
		})
	})
	b.Run("Zap.CheckSampled", func(b *testing.B) {
		logger := newSampledLogger(zap.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				if ce := logger.Check(zap.InfoLevel, getMessage(i)); ce != nil {
					ce.Write(fakeFields()...)
				}
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(getMessage(0), fakeSugarFields()...)
			}
		})
	})
	b.Run("apex/log", func(b *testing.B) {
		logger := newApexLog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeApexFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("go-kit/kit/log", func(b *testing.B) {
		logger := newKitLog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Log(fakeSugarFields()...)
			}
		})
	})
	b.Run("inconshreveable/log15", func(b *testing.B) {
		logger := newLog15()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0), fakeSugarFields()...)
			}
		})
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				fakeZerologFields(logger.Info()).Msg(getMessage(0))
			}
		})
	})
	b.Run("rs/zerolog.Check", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if e := logger.Info(); e.Enabled() {
					fakeZerologFields(e).Msg(getMessage(0))
				}
			}
		})
	})
}
