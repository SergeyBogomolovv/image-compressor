package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/SergeyBogomolovv/image-compressor/pkg/utils"
)

type imageService struct {
	log     *slog.Logger
	saveDir string
}

func New(log *slog.Logger, saveDir string) *imageService {
	return &imageService{
		log:     log,
		saveDir: saveDir,
	}
}

func (s *imageService) CompressImage(ctx context.Context, header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	imageName := strings.TrimSuffix(header.Filename, path.Ext(header.Filename))
	log := s.log.With(slog.String("name", imageName))

	log.Info("Compressing image")

	if err != nil {
		log.Error("Failed to open file", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to open file")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Error("Failed to decode image", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to decode image")
	}

	err = os.MkdirAll(s.saveDir, 0755)
	if err != nil {
		log.Error("Failed to create folder", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to create folder")
	}

	archiveName := fmt.Sprintf("%s_%s.zip", imageName, generateHash())
	archiveFile, err := os.Create(path.Join(s.saveDir, archiveName))
	if err != nil {
		log.Error("Failed to create archive file", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to create archive")
	}
	defer archiveFile.Close()

	archive := zip.NewWriter(archiveFile)

	if err := processImage(ctx, archive, img, imageName); err != nil {
		log.Error("Failed to compress image", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to compress image")
	}

	if err := archive.Close(); err != nil {
		log.Error("Failed to close archive", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to close archive")
	}

	log.Info("Image compressed")

	return archiveName, nil
}

func processImage(ctx context.Context, archive *zip.Writer, img image.Image, name string) error {
	type processed struct {
		data    []byte
		quality int
	}
	qualities := []int{50, 70, 90}
	var wg sync.WaitGroup
	errs := make(chan error, 1)
	results := make(chan processed, len(qualities))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, quality := range qualities {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(quality int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				data, err := compressJPEG(img, quality, float64(quality)/110)
				if err != nil {
					select {
					case errs <- err:
					default:
					}
					cancel()
					return
				}
				results <- processed{data: data, quality: quality}
			}
		}(quality)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			return err
		case result, ok := <-results:
			if !ok {
				return nil
			}
			name := fmt.Sprintf("%s_%d.jpg", name, result.quality)
			f, err := archive.Create(name)
			if err != nil {
				return err
			}
			if _, err := f.Write(result.data); err != nil {
				return err
			}
		}
	}
}

func compressJPEG(img image.Image, quality int, scale float64) ([]byte, error) {
	srcBounds := img.Bounds()
	newWidth := int(float64(srcBounds.Dx()) * scale)
	newHeight := int(float64(srcBounds.Dy()) * scale)

	resizedImg := utils.BilinearResize(img, newWidth, newHeight)

	var buff bytes.Buffer
	if err := jpeg.Encode(&buff, resizedImg, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func generateHash() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
