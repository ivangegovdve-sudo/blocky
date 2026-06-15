package querylog

import (
	"time"

	"github.com/0xERR0R/blocky/log"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("LoggerWriter", func() {
	Describe("logger query log", func() {
		When("New log entry was created", func() {
			It("should be logged", func() {
				writer := NewLoggerWriter()
				logger, rec := log.NewRecorder()
				writer.logger = logger

				writer.Write(&LogEntry{
					Start:      time.Now(),
					DurationMs: 20,
				})

				Expect(rec.Records()).Should(HaveLen(1))
				Expect(rec.LastMessage()).Should(Equal("query resolved"))
			})
		})
		When("Cleanup is called", func() {
			It("should do nothing", func() {
				writer := NewLoggerWriter()
				writer.CleanUp()
			})
		})
	})

	Describe("LogEntryFields", func() {
		It("should return log fields", func() {
			entry := LogEntry{
				ClientIP:     "ip",
				DurationMs:   100,
				QuestionType: "qtype",
				ResponseCode: "rcode",
			}

			fields := LogEntryFields(&entry)

			Expect(fields).Should(HaveKeyWithValue("client_ip", entry.ClientIP))
			Expect(fields).Should(HaveKeyWithValue("duration_ms", entry.DurationMs))
			Expect(fields).Should(HaveKeyWithValue("question_type", entry.QuestionType))
			Expect(fields).Should(HaveKeyWithValue("response_code", entry.ResponseCode))

			Expect(fields).ShouldNot(HaveKey("client_names"))
			Expect(fields).ShouldNot(HaveKey("question_name"))
		})
	})

	DescribeTable("withoutZeroes",
		func(value any, isZero bool) {
			fields := withoutZeroes(map[string]any{"a": value})

			if isZero {
				Expect(fields).Should(BeEmpty())
			} else {
				Expect(fields).ShouldNot(BeEmpty())
			}
		},
		Entry("empty string",
			"",
			true),
		Entry("non-empty string",
			"something",
			false),
		Entry("zero int",
			0,
			true),
		Entry("non-zero int",
			1,
			false),
	)
})
