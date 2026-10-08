// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"forgejo.org/modules/git"
	"forgejo.org/modules/json"
	"forgejo.org/modules/private"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// Capture what's being written into a standard file descriptor.
func captureOutput(t *testing.T, stdFD *os.File) (finish func() (output string)) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	resetStdout := test.MockVariableValue(stdFD, *w)

	return func() (output string) {
		w.Close()
		resetStdout()

		out, err := io.ReadAll(r)
		require.NoError(t, err)
		return string(out)
	}
}

func TestPktLine(t *testing.T) {
	ctx := t.Context()

	t.Run("Read", func(t *testing.T) {
		s := strings.NewReader("0000")
		r := bufio.NewReader(s)
		result, err := readPktLine(ctx, r, pktLineTypeFlush)
		require.NoError(t, err)
		assert.Equal(t, pktLineTypeFlush, result.Type)

		s = strings.NewReader("0006a\n")
		r = bufio.NewReader(s)
		result, err = readPktLine(ctx, r, pktLineTypeData)
		require.NoError(t, err)
		assert.Equal(t, pktLineTypeData, result.Type)
		assert.Equal(t, []byte("a\n"), result.Data)

		s = strings.NewReader("0004")
		r = bufio.NewReader(s)
		result, err = readPktLine(ctx, r, pktLineTypeData)
		require.Error(t, err)
		assert.Nil(t, result)

		data := strings.Repeat("x", 65516)
		r = bufio.NewReader(strings.NewReader("fff0" + data))
		result, err = readPktLine(ctx, r, pktLineTypeData)
		require.NoError(t, err)
		assert.Equal(t, pktLineTypeData, result.Type)
		assert.Equal(t, []byte(data), result.Data)

		r = bufio.NewReader(strings.NewReader("fff1a"))
		result, err = readPktLine(ctx, r, pktLineTypeData)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Write", func(t *testing.T) {
		w := bytes.NewBuffer([]byte{})
		err := writeFlushPktLine(ctx, w)
		require.NoError(t, err)
		assert.Equal(t, []byte("0000"), w.Bytes())

		w.Reset()
		err = writeDataPktLine(ctx, w, []byte("a\nb"))
		require.NoError(t, err)
		assert.Equal(t, []byte("0007a\nb"), w.Bytes())

		w.Reset()
		data := bytes.Repeat([]byte{0x05}, 288)
		err = writeDataPktLine(ctx, w, data)
		require.NoError(t, err)
		assert.Equal(t, append([]byte("0124"), data...), w.Bytes())

		w.Reset()
		err = writeDataPktLine(ctx, w, nil)
		require.Error(t, err)
		assert.Empty(t, w.Bytes())

		w.Reset()
		data = bytes.Repeat([]byte{0x64}, 65516)
		err = writeDataPktLine(ctx, w, data)
		require.NoError(t, err)
		assert.Equal(t, append([]byte("fff0"), data...), w.Bytes())

		w.Reset()
		err = writeDataPktLine(ctx, w, bytes.Repeat([]byte{0x64}, 65516+1))
		require.Error(t, err)
		assert.Empty(t, w.Bytes())
	})
}

func TestDelayWriter(t *testing.T) {
	// Setup the environment.
	defer test.MockVariableValue(&setting.InternalToken, "Random")()
	defer test.MockVariableValue(&setting.InstallLock, true)()
	defer test.MockVariableValue(&setting.Git.VerbosePush, true)()
	t.Setenv("SSH_ORIGINAL_COMMAND", "true")

	// Setup the Stdin.
	f, err := os.OpenFile(t.TempDir()+"/stdin", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
	require.NoError(t, err)
	_, err = f.Write([]byte("00000000000000000000 00000000000000000001 refs/head/main\n"))
	require.NoError(t, err)
	_, err = f.Seek(0, 0)
	require.NoError(t, err)
	defer test.MockVariableValue(os.Stdin, *f)()

	// Setup the server that processes the hooks.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 600)
	}))
	defer ts.Close()
	defer test.MockVariableValue(&setting.LocalURL, ts.URL+"/")()

	app := cli.Command{}
	app.Commands = []*cli.Command{subcmdHookPreReceive()}

	t.Run("Should delay", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Git.VerbosePushDelay, time.Millisecond*500)()
		finish := captureOutput(t, os.Stdout)

		err = app.Run(t.Context(), []string{"./forgejo", "pre-receive"})
		require.NoError(t, err)
		out := finish()

		require.Contains(t, out, "* Checking 1 references")
		require.Contains(t, out, "Checked 1 references in total")
	})

	t.Run("Shouldn't delay", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Git.VerbosePushDelay, time.Second*5)()
		finish := captureOutput(t, os.Stdout)

		err = app.Run(t.Context(), []string{"./forgejo", "pre-receive"})
		require.NoError(t, err)
		out := finish()

		require.NoError(t, err)
		require.Empty(t, out)
	})
}

func TestRunHookPrePostReceive(t *testing.T) {
	// Setup the environment.
	defer test.MockVariableValue(&setting.InternalToken, "Random")()
	defer test.MockVariableValue(&setting.InstallLock, true)()
	defer test.MockVariableValue(&setting.Git.VerbosePush, true)()
	t.Setenv("SSH_ORIGINAL_COMMAND", "true")

	tests := []struct {
		name        string
		inputLine   string
		oldCommitID string
		newCommitID string
		refFullName string
	}{
		{
			name:        "base case",
			inputLine:   "00000000000000000000 00000000000000000001 refs/head/main\n",
			oldCommitID: "00000000000000000000",
			newCommitID: "00000000000000000001",
			refFullName: "refs/head/main",
		},
		{
			name:        "nbsp case",
			inputLine:   "00000000000000000000 00000000000000000001 refs/head/ma\u00A0in\n",
			oldCommitID: "00000000000000000000",
			newCommitID: "00000000000000000001",
			refFullName: "refs/head/ma\u00A0in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the Stdin.
			f, err := os.OpenFile(t.TempDir()+"/stdin", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
			require.NoError(t, err)
			_, err = f.Write([]byte(tt.inputLine))
			require.NoError(t, err)
			_, err = f.Seek(0, 0)
			require.NoError(t, err)
			defer test.MockVariableValue(os.Stdin, *f)()

			// Setup the server that processes the hooks.
			var serverError error
			var hookOpts *private.HookOptions

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					serverError = err
					w.WriteHeader(500)
					return
				}

				err = json.Unmarshal(body, &hookOpts)
				if err != nil {
					serverError = err
					w.WriteHeader(500)
					return
				}

				w.WriteHeader(200)

				resp := &private.HookPostReceiveResult{}
				bytes, err := json.Marshal(resp)
				if err != nil {
					serverError = err
					return
				}

				_, err = w.Write(bytes)
				if err != nil {
					serverError = err
					return
				}
			}))
			defer ts.Close()
			defer test.MockVariableValue(&setting.LocalURL, ts.URL+"/")()

			t.Run("pre-receive", func(t *testing.T) {
				app := cli.Command{}
				app.Commands = []*cli.Command{subcmdHookPreReceive()}

				finish := captureOutput(t, os.Stdout)
				err = app.Run(t.Context(), []string{"./forgejo", "pre-receive"})
				require.NoError(t, err)
				out := finish()
				require.Empty(t, out)

				require.NoError(t, serverError)
				require.NotNil(t, hookOpts)

				require.Len(t, hookOpts.OldCommitIDs, 1)
				assert.Equal(t, tt.oldCommitID, hookOpts.OldCommitIDs[0])
				require.Len(t, hookOpts.NewCommitIDs, 1)
				assert.Equal(t, tt.newCommitID, hookOpts.NewCommitIDs[0])
				require.Len(t, hookOpts.RefFullNames, 1)
				assert.Equal(t, git.RefName(tt.refFullName), hookOpts.RefFullNames[0])
			})

			// seek stdin back to beginning
			_, err = f.Seek(0, 0)
			require.NoError(t, err)
			// reset state from prev test
			serverError = nil
			hookOpts = nil

			t.Run("post-receive", func(t *testing.T) {
				app := cli.Command{}
				app.Commands = []*cli.Command{subcmdHookPostReceive()}

				finish := captureOutput(t, os.Stdout)
				err = app.Run(t.Context(), []string{"./forgejo", "post-receive"})
				require.NoError(t, err)
				out := finish()
				require.Empty(t, out)

				require.NoError(t, serverError)
				require.NotNil(t, hookOpts)

				require.Len(t, hookOpts.OldCommitIDs, 1)
				assert.Equal(t, tt.oldCommitID, hookOpts.OldCommitIDs[0])
				require.Len(t, hookOpts.NewCommitIDs, 1)
				assert.Equal(t, tt.newCommitID, hookOpts.NewCommitIDs[0])
				require.Len(t, hookOpts.RefFullNames, 1)
				assert.Equal(t, git.RefName(tt.refFullName), hookOpts.RefFullNames[0])
			})
		})
	}
}
