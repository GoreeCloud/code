// Copyright 2014 The Gogs Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package avatar

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"

	_ "image/gif"  // for processing gif images
	_ "image/jpeg" // for processing jpeg images

	"forgejo.org/modules/avatar/identicon"
	"forgejo.org/modules/setting"

	"golang.org/x/image/draw"

	_ "golang.org/x/image/webp" // for processing webp images
)

// DefaultAvatarSize is the target CSS pixel size for avatar generation. It is
// multiplied by setting.Avatar.RenderedSizeFactor and the resulting size is the
// usual size of avatar image saved on server, unless the original file is smaller
// than the size after resizing.
const DefaultAvatarSize = 256

// Sizes to which we allow resizing an avatar down to. They must be specified in increasing order.
var AllowedResizedAvatarSizes = []int{64, 128}

// RandomImageSize generates and returns a random avatar image unique to input data
// in custom size (height and width).
func RandomImageSize(size int, data []byte) (image.Image, error) {
	// we use white as background, and use dark colors to draw blocks
	imgMaker, err := identicon.New(size, color.White, identicon.DarkColors...)
	if err != nil {
		return nil, fmt.Errorf("identicon.New: %w", err)
	}
	return imgMaker.Make(data), nil
}

// RandomImage generates and returns a random avatar image unique to input data
// in default size (height and width).
func RandomImage(data []byte) (image.Image, error) {
	return RandomImageSize(DefaultAvatarSize*setting.Avatar.RenderedSizeFactor, data)
}

// processAvatarImage process the avatar image data, crop and resize it if necessary.
// the returned data could be the original image if no processing is needed.
func processAvatarImage(data []byte, maxOriginSize int64) ([]byte, image.Image, error) {
	imgCfg, imgType, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("image.DecodeConfig: %w", err)
	}

	// for safety, only accept known types explicitly
	if imgType != "png" && imgType != "jpeg" && imgType != "gif" && imgType != "webp" {
		return nil, nil, errors.New("unsupported avatar image type")
	}

	// do not process image which is too large, it would consume too much memory
	if imgCfg.Width > setting.Avatar.MaxWidth {
		return nil, nil, fmt.Errorf("image width is too large: %d > %d", imgCfg.Width, setting.Avatar.MaxWidth)
	}
	if imgCfg.Height > setting.Avatar.MaxHeight {
		return nil, nil, fmt.Errorf("image height is too large: %d > %d", imgCfg.Height, setting.Avatar.MaxHeight)
	}

	// If the origin is small enough, just use it, then APNG could be supported,
	// otherwise, if the image is processed later, APNG loses animation.
	// And one more thing, webp is not fully supported, for animated webp, image.DecodeConfig works but Decode fails.
	// So for animated webp, if the uploaded file is smaller than maxOriginSize, it will be used, if it's larger, there will be an error.
	if len(data) < int(maxOriginSize) {
		//nolint:nilnil
		return data, nil, nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("image.Decode: %w", err)
	}

	// try to crop and resize the origin image if necessary
	img = cropSquare(img)

	targetSize := DefaultAvatarSize * setting.Avatar.RenderedSizeFactor
	img = Scale(img, targetSize, targetSize, draw.BiLinear)

	// try to encode the cropped/resized image to png
	bs := bytes.Buffer{}
	if err = png.Encode(&bs, img); err != nil {
		return nil, nil, err
	}
	resized := bs.Bytes()

	// usually the png compression is not good enough, use the original image (no cropping/resizing) if the origin is smaller
	if len(data) <= len(resized) {
		return data, img, nil
	}

	return resized, img, nil
}

// ProcessAvatarImage process the avatar image data, crop and resize it if necessary.
// the returned data could be the original image if no processing is needed.
func ProcessAvatarImage(data []byte) ([]byte, image.Image, error) {
	return processAvatarImage(data, setting.Avatar.MaxOriginSize)
}

// Scale resizes the image to width x height using the given scaler.
func Scale(src image.Image, width, height int, scale draw.Scaler) image.Image {
	rect := image.Rect(0, 0, width, height)
	dst := image.NewRGBA(rect)
	scale.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

// cropSquare crops the largest square image from the center of the image.
// If the image is already square, it is returned unchanged.
func cropSquare(src image.Image) image.Image {
	bounds := src.Bounds()
	if bounds.Dx() == bounds.Dy() {
		return src
	}

	var rect image.Rectangle
	if bounds.Dx() > bounds.Dy() {
		// width > height
		size := bounds.Dy()
		rect = image.Rect((bounds.Dx()-size)/2, 0, (bounds.Dx()+size)/2, size)
	} else {
		// width < height
		size := bounds.Dx()
		rect = image.Rect(0, (bounds.Dy()-size)/2, size, (bounds.Dy()+size)/2)
	}

	dst := image.NewRGBA(rect)
	draw.Draw(dst, rect, src, rect.Min, draw.Src)
	return dst
}

// BestAvatarCachedSize computes the size at which the avatar should be downloaded, to display it at the desired size.
// When it returns 0, it means that the original should be downloaded.
func BestAvatarCachedSize(size int) int {
	if size == 0 {
		return 0
	}
	for i := 0; i < len(AllowedResizedAvatarSizes); i++ {
		if AllowedResizedAvatarSizes[i] >= size {
			return AllowedResizedAvatarSizes[i]
		}
	}
	return 0
}
