// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"forgejo.org/models/asymkey"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
)

var logger = log.GetManager().GetLogger("ssh")

func Init(ctx context.Context) error {
	if setting.SSH.Disabled {
		builtinUnused()
		return nil
	}

	if setting.SSH.StartBuiltinServer {
		Listen(setting.SSH.ListenHost, setting.SSH.ListenPort, setting.SSH.ServerCiphers, setting.SSH.ServerKeyExchanges, setting.SSH.ServerMACs)
		log.Info("SSH server started on %s. Cipher list (%v), key exchange algorithms (%v), MACs (%v)",
			net.JoinHostPort(setting.SSH.ListenHost, strconv.Itoa(setting.SSH.ListenPort)),
			setting.SSH.ServerCiphers, setting.SSH.ServerKeyExchanges, setting.SSH.ServerMACs,
		)
		return nil
	}

	builtinUnused()

	// FIXME: why 0o644 for a directory .....
	if err := os.MkdirAll(setting.SSH.KeyTestPath, 0o644); err != nil {
		return fmt.Errorf("failed to create directory %q for ssh key test: %w", setting.SSH.KeyTestPath, err)
	}

	if len(setting.SSH.TrustedUserCAKeys) > 0 && setting.SSH.AuthorizedPrincipalsEnabled {
		caKeysFileName := setting.SSH.TrustedUserCAKeysFile
		caKeysFileDir := filepath.Dir(caKeysFileName)

		err := os.MkdirAll(caKeysFileDir, 0o700) // SSH.RootPath by default (That is `~/.ssh` in most cases)
		if err != nil {
			return fmt.Errorf("failed to create directory %q for ssh trusted ca keys: %w", caKeysFileDir, err)
		}

		if err := os.WriteFile(caKeysFileName, []byte(strings.Join(setting.SSH.TrustedUserCAKeys, "\n")), 0o600); err != nil {
			return fmt.Errorf("failed to write ssh trusted ca keys to %q: %w", caKeysFileName, err)
		}
	}

	if !setting.SSH.AllowUnexpectedAuthorizedKeys {
		findings, err := asymkey.InspectPublicKeys(ctx)
		if err != nil {
			return fmt.Errorf("inspect authorized_keys failed: %w", err)
		}

		unexpectedKeys := []string{}
		for _, finding := range findings {
			switch finding.Type {
			case asymkey.InspectionResultFileMissing:
				err := asymkey.RewriteAllPublicKeys(ctx)
				if err != nil {
					return fmt.Errorf("rewrite authorized_keys failed: %w", err)
				}
			case asymkey.InspectionResultUnexpectedKey:
				unexpectedKeys = append(unexpectedKeys, finding.Comment)

			case asymkey.InspectionResultMissingExpectedKey:
				// MissingExpectingKey can happen at the same time as UnexpectedKey -- so while it might seem to make
				// sense to regenerate the key file automatically in this case, it could overwrite keys that someone
				// wants present there when they want SSH_ALLOW_UNEXPECTED_AUTHORIZED_KEYS=true but haven't set it yet.
				// So, just warn as this isn't an insecure situation.
				log.Warn("authorized_keys is missing a key from the database; regenerate authorized_keys from the admin panel to repair this")
			}
		}

		if len(unexpectedKeys) > 0 {
			detailConcat := strings.Join(unexpectedKeys, "\n\t")
			log.Fatal("An unexpected ssh public key was discovered. Forgejo will shutdown to require this to be fixed. Fix by either:\n"+
				"Option 1: Delete the file %s, and Forgejo will recreate it with only expected ssh public keys.\n"+
				"Option 2: Permit unexpected keys by setting [server].SSH_ALLOW_UNEXPECTED_AUTHORIZED_KEYS=true in Forgejo's config file.\n"+
				"Option 3: If unused, disable SSH support by setting [server].DISABLE_SSH=true in Forgejo's config file.\n"+
				"\t"+
				detailConcat, filepath.Join(setting.SSH.RootPath, "authorized_keys"))
		}
	}

	return nil
}
