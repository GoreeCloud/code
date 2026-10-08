// Copyright 2025 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package log_test

import (
	"testing"

	"forgejo.org/modules/log"

	"github.com/stretchr/testify/assert"
)

func TestBufferLogger(t *testing.T) {
	prefix := "TestPrefix "
	level := log.INFO
	expected := "something"

	bufferWriter := log.NewEventWriterBuffer("test-buffer", log.WriterMode{
		Level:      level,
		Prefix:     prefix,
		Expression: expected,
	})

	logger := log.NewLoggerWithWriters(t.Context(), "test", bufferWriter)

	logger.SendLogEvent(&log.Event{
		Level:         log.INFO,
		MsgSimpleText: expected,
	})
	logger.Close()
	assert.Contains(t, bufferWriter.Buffer.String(), expected)
}

func TestBufferLoggerWithExclusion(t *testing.T) {
	prefix := "ExclusionPrefix "
	level := log.INFO
	message := "something"

	bufferWriter := log.NewEventWriterBuffer("test-buffer", log.WriterMode{
		Level:     level,
		Prefix:    prefix,
		Exclusion: message,
	})

	logger := log.NewLoggerWithWriters(t.Context(), "test", bufferWriter)

	logger.SendLogEvent(&log.Event{
		Level:         log.INFO,
		MsgSimpleText: message,
	})
	logger.Close()
	assert.NotContains(t, bufferWriter.Buffer.String(), message)
}

func TestBufferLoggerWithExpressionAndExclusion(t *testing.T) {
	prefix := "BothPrefix "
	level := log.INFO
	expression := ".*foo.*"
	exclusion := ".*bar.*"

	bufferWriter := log.NewEventWriterBuffer("test-buffer", log.WriterMode{
		Level:      level,
		Prefix:     prefix,
		Expression: expression,
		Exclusion:  exclusion,
	})

	logger := log.NewLoggerWithWriters(t.Context(), "test", bufferWriter)

	logger.SendLogEvent(&log.Event{Level: log.INFO, MsgSimpleText: "foo expression"})
	logger.SendLogEvent(&log.Event{Level: log.INFO, MsgSimpleText: "bar exclusion"})
	logger.SendLogEvent(&log.Event{Level: log.INFO, MsgSimpleText: "foo bar both"})
	logger.SendLogEvent(&log.Event{Level: log.INFO, MsgSimpleText: "none"})
	logger.Close()

	assert.Contains(t, bufferWriter.Buffer.String(), "foo expression")
	assert.NotContains(t, bufferWriter.Buffer.String(), "bar")
}
