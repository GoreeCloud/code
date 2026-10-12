// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/url"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/modules/httplib"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/storage"
	"forgejo.org/services/context"
)

// ServeArtifact writes an artifact archive to the response. The given rows
// must all belong to the same logical artifact (same run_id + artifact_name)
// and be in ArtifactStatusUploadConfirmed; callers are responsible for
// validating status and repository ownership beforehand.
//
// When the artifact was produced by the v4 backend (a single zip already
// sitting in storage), it is streamed or redirected to directly. Otherwise
// (v1-v3 backend) the archive is assembled on the fly from the individual
// file rows, applying gzip decompression where needed.
func ServeArtifact(base *context.Base, artifacts []*actions_model.ActionArtifact) error {
	if len(artifacts) == 0 {
		return errors.New("no artifacts to serve")
	}

	if len(artifacts) == 1 && artifacts[0].IsV4() {
		return serveV4Artifact(base, artifacts[0])
	}
	return serveLegacyArtifact(base, artifacts)
}

func serveV4Artifact(base *context.Base, art *actions_model.ActionArtifact) error {
	if setting.Actions.ArtifactStorage.MinioConfig.ServeDirect {
		u, err := storage.ActionsArtifacts.URL(art.StoragePath, art.ArtifactPath, nil)
		if u != nil && err == nil {
			base.Redirect(u.String())
			return nil
		}
	}
	f, err := storage.ActionsArtifacts.Open(art.StoragePath)
	if err != nil {
		return err
	}
	httplib.ServeContentByReadSeeker(base.Req, base.Resp, art.ArtifactName+".zip", new(art.UpdatedUnix.AsTime()), f)
	return nil
}

func serveLegacyArtifact(base *context.Base, artifacts []*actions_model.ActionArtifact) error {
	name := artifacts[0].ArtifactName
	base.Resp.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip; filename*=UTF-8''%s.zip", url.PathEscape(name), name))

	writer := zip.NewWriter(base.Resp)
	defer writer.Close()

	for _, art := range artifacts {
		if err := writeArtifactFile(writer, art); err != nil {
			return err
		}
	}
	return nil
}

func writeArtifactFile(writer *zip.Writer, art *actions_model.ActionArtifact) error {
	f, err := storage.ActionsArtifacts.Open(art.StoragePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var r io.Reader = f
	if art.ContentEncoding == "gzip" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gz.Close()
		r = gz
	}

	w, err := writer.Create(art.ArtifactPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	return err
}
