package log

import (
	"context"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
)

// These benchmarks isolate the per-request logging plumbing that runs on the
// DNS hot path, independent of the underlying logging library. They are meant
// to be re-run unchanged after a logging-backend migration to compare
// allocations/op before and after.
//
// Run with:
//
//	go test -run=^$ -bench=. -benchmem ./log/

// benchSetup configures the global logger at the given level with a discarded
// output, so we measure the plumbing cost (entry/field/context building) and
// not the cost of formatting and writing bytes.
func benchSetup(b *testing.B, level logrus.Level) {
	b.Helper()

	Configure(&Config{Level: level, Format: FormatTypeText, Timestamp: true})
	Log().SetOutput(io.Discard)

	b.ReportAllocs()
}

// fields mirror what server.newRequest attaches to every incoming request.
func requestFields() logrus.Fields {
	return logrus.Fields{
		"req_id":    "b8f4c1e2-0000-1111-2222-333344445555",
		"question":  "A example.com.",
		"client_ip": "192.168.178.1",
	}
}

// BenchmarkCtxWithFields measures attaching the per-request fields to the
// context, as done once per request in server.newRequest.
func BenchmarkCtxWithFields(b *testing.B) {
	benchSetup(b, logrus.InfoLevel)

	ctx := context.Background()

	for b.Loop() {
		_, _ = CtxWithFields(ctx, requestFields())
	}
}

// BenchmarkFromCtx measures retrieving the request logger from the context,
// as done by every resolver hop and by util.LogOnError etc.
func BenchmarkFromCtx(b *testing.B) {
	ctx, _ := CtxWithFields(context.Background(), requestFields())

	benchSetup(b, logrus.InfoLevel)

	for b.Loop() {
		_ = FromCtx(ctx)
	}
}

// BenchmarkWrapCtxWithPrefix measures one resolver hop's logging plumbing:
// WrapCtx + WithPrefix, which is what resolver.typed.log(ctx) does at the top
// of each resolver's Resolve, on every request, regardless of log level.
func BenchmarkWrapCtxWithPrefix(b *testing.B) {
	ctx, _ := CtxWithFields(context.Background(), requestFields())

	benchSetup(b, logrus.InfoLevel)

	for b.Loop() {
		_, _ = WrapCtx(ctx, func(e *logrus.Entry) *logrus.Entry {
			return WithPrefix(e, "blocking")
		})
	}
}

// chainPrefixes is a representative resolver chain depth: roughly the number of
// resolvers a typical query traverses (client-names, query-log, blocking,
// caching, conditional, custom-dns, hosts-file, ... , upstream).
var chainPrefixes = []string{
	"client_names", "query_logging", "metrics", "blocking", "caching",
	"conditional", "custom_dns", "hosts_file", "dns64", "upstream",
}

// benchmarkRequestChain simulates the full per-request logging path: attach the
// request fields once, then walk the resolver chain, with each hop doing its
// WrapCtx+WithPrefix plumbing and emitting one Debug line (which is a no-op at
// info level but still pays the plumbing cost).
func benchmarkRequestChain(b *testing.B, level logrus.Level) {
	benchSetup(b, level)

	base := context.Background()

	for b.Loop() {
		ctx, _ := CtxWithFields(base, requestFields())

		for _, prefix := range chainPrefixes {
			var logger *logrus.Entry

			ctx, logger = WrapCtx(ctx, func(e *logrus.Entry) *logrus.Entry {
				return WithPrefix(e, prefix)
			})

			logger.Debugf("resolving in %s", prefix)
		}
	}
}

// BenchmarkRequestChain_InfoLevel is the common production case: debug/trace
// lines are disabled, yet the plumbing still allocates on every hop.
func BenchmarkRequestChain_InfoLevel(b *testing.B) {
	benchmarkRequestChain(b, logrus.InfoLevel)
}

// BenchmarkRequestChain_DebugLevel shows the cost when debug lines are actually
// emitted (formatting + writing is discarded, so this is still plumbing-bound).
func BenchmarkRequestChain_DebugLevel(b *testing.B) {
	benchmarkRequestChain(b, logrus.DebugLevel)
}
