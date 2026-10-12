// Copyright 2017 The Gitea Authors. All rights reserved.
// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"forgejo.org/modules/avatar"
	"forgejo.org/modules/httpcache"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/storage"
	"forgejo.org/modules/util"
	"forgejo.org/modules/web/routing"

	"golang.org/x/image/draw"
)

// resizingHandler resizes images to one of the supported sizes.
// It expects URLs of the form "{prefix}/{size}/{image_id}"
func resizingHandler(prefix string, imgStore storage.ObjectStorage, allowedSizes []int) http.HandlerFunc {
	whenMissing := func(path string, size int) ([]byte, error) {
		return resizeImageFromStorage(imgStore, path, size, allowedSizes)
	}
	return cachingHandler(prefix, imgStore, whenMissing)
}

// resizeImageFromStorage retrieves an image from an ObjectStorage and resizes it
func resizeImageFromStorage(imgStore storage.ObjectStorage, imgPath string, targetSize int, allowedSizes []int) ([]byte, error) {
	if !slices.Contains(allowedSizes, targetSize) {
		return nil, errors.New("invalid image size requested")
	}

	// Get the original image
	reader, err := imgStore.Open(imgPath)
	if err != nil {
		return nil, err
	}
	// Read the bytes in memory for the purpose of re-using them if the image is small
	originalBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	originalReader := bytes.NewReader(originalBytes)
	// Decode it as an image
	image, _, err := image.Decode(originalReader)
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)

	width := image.Bounds().Dx()
	height := image.Bounds().Dy()

	if width <= targetSize && height <= targetSize {
		// The original image is smaller than the requested size,
		// just return it directly.
		// This will still put a copy of it in the storage for the
		// requested size which we could avoid, but that will also
		// avoid the need for decoding the image next time it is requested.
		return originalBytes, err
	}

	thumbnail := avatar.Scale(image, targetSize, targetSize, draw.BiLinear)
	err = png.Encode(buffer, thumbnail)
	return buffer.Bytes(), err
}

// cachingHandler serves blobs from an ObjectStorage,
// computing them on the fly if they are missing. It falls back on the original image if no size parameter is supplied.
func cachingHandler(prefix string, imgStore storage.ObjectStorage, whenMissing func(string, int) ([]byte, error)) http.HandlerFunc {
	prefix = strings.Trim(prefix, "/")
	funcInfo := routing.GetFuncInfo(cachingHandler, prefix)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" && req.Method != "HEAD" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if !strings.HasPrefix(req.URL.Path, "/"+prefix+"/") {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		routing.UpdateFuncInfo(req.Context(), funcInfo)

		sizeParam := req.URL.Query().Get("size")
		if sizeParam == "" {
			sizeParam = "0"
		}
		size, err := strconv.Atoi(sizeParam)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		rPath := strings.TrimPrefix(req.URL.Path, "/"+prefix+"/")
		rPath = util.PathJoinRelX(rPath)
		if rPath == "" || rPath == "." || strings.Contains(rPath, "/") {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if size != 0 && !slices.Contains(avatar.AllowedResizedAvatarSizes, size) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		originalFile, err := imgStore.Stat(rPath)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// If no size is provided, or if the original file is small enough, fall back on the original image.
		// This is primarily for the purpose of preserving animated images that are small enough, because
		// the image-resizing code would remove the animation from the images otherwise.
		// An alternative could be to only preserve such images if they are actually animated, but it would
		// require reading the image and and inspecting its contents.
		if size == 0 || originalFile.Size() < setting.Avatar.MaxOriginSize {
			reader, err := imgStore.Open(rPath)
			if err != nil {
				log.Error("Error whilst opening %s %s. Error: %v", prefix, rPath, err)
				http.Error(w, fmt.Sprintf("Error whilst opening %s %s", prefix, rPath), http.StatusInternalServerError)
				return
			}
			defer reader.Close()
			httpcache.ServeContentWithCacheControl(w, req, path.Base(rPath), originalFile.ModTime(), reader)
			return
		}

		cachePath := fmt.Sprintf("resized/%d/%s", size, rPath)

		fi, err := imgStore.Stat(cachePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// attempt to compute the missing value via the callback provided
				computed, err := whenMissing(rPath, size)
				if err != nil {
					log.Warn("Unable to compute %s %s", prefix, rPath)
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
				reader := bytes.NewReader(computed)
				_, err = imgStore.Save(cachePath, reader, int64(len(computed)))
				if err != nil {
					// only log the error, still return the computed resource without caching it
					log.Warn("Unable to save %s %s: %s", prefix, cachePath, err)
				}

				reader = bytes.NewReader(computed)
				httpcache.ServeContentWithCacheControl(w, req, path.Base(cachePath), time.Now(), reader)
				return
			}
			log.Error("Error whilst opening %s %s. Error: %v", prefix, rPath, err)
			http.Error(w, fmt.Sprintf("Error whilst opening %s %s", prefix, rPath), http.StatusInternalServerError)
			return
		}

		fr, err := imgStore.Open(cachePath)
		if err != nil {
			log.Error("Error whilst opening %s %s. Error: %v", prefix, cachePath, err)
			http.Error(w, fmt.Sprintf("Error whilst opening %s %s", prefix, cachePath), http.StatusInternalServerError)
			return
		}
		defer fr.Close()
		httpcache.ServeContentWithCacheControl(w, req, path.Base(rPath), fi.ModTime(), fr)
	})
}
