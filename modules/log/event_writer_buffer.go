// Copyright 2025 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package log

import (
	"bytes"
	"sync"
)

type EventWriterBuffer interface {
	EventWriter
	GetString() string
}

type eventWriterBuffer struct {
	*EventWriterBaseImpl
	buffer *bytes.Buffer
	mu     sync.RWMutex
}

var _ EventWriterBuffer = (*eventWriterBuffer)(nil)

func (*eventWriterBuffer) Close() error {
	return nil
}

func (o *eventWriterBuffer) Write(p []byte) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.buffer.Write(p)
}

func (o *eventWriterBuffer) GetString() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	b := o.buffer.Bytes()
	s := make([]byte, len(b))
	copy(s, b)
	return string(s)
}

func NewEventWriterBuffer(name string, mode WriterMode) EventWriter {
	w := &eventWriterBuffer{EventWriterBaseImpl: NewEventWriterBase(name, "buffer", mode)}
	w.buffer = new(bytes.Buffer)
	w.OutputWriteCloser = w
	return w
}
