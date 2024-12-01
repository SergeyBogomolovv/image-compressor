package controller

import (
	"context"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/SergeyBogomolovv/image-compressor/internal/domain"
	"github.com/SergeyBogomolovv/image-compressor/pkg/utils"
)

type Service interface {
	CompressImage(ctx context.Context, header *multipart.FileHeader) (string, error)
}

type controller struct {
	service   Service
	outputDir string
}

func Register(router *http.ServeMux, service Service, outputDir string) {
	controller := &controller{service: service, outputDir: outputDir}
	router.HandleFunc("POST /upload", controller.UploadHandler)
	router.HandleFunc("GET /download/{name}", controller.DownloadHandler)
}

func (c *controller) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		utils.WriteError(w, utils.NewError("too many images", http.StatusBadRequest))
		return
	}

	images := r.MultipartForm.File["images"]
	if len(images) == 0 {
		utils.WriteError(w, utils.NewError("no images provided", http.StatusBadRequest))
		return
	}

	results := c.processFiles(r.Context(), images)

	response := &domain.ProcessedResponse{Success: []string{}, Errors: []string{}}

	for result := range results {
		if result.Error != nil {
			response.Errors = append(response.Errors, result.Error.Error())
		}
		if result.Path != "" {
			response.Success = append(response.Success, result.Path)
		}
	}
	utils.WriteJSON(w, response, http.StatusCreated)
}

func (c *controller) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	filePath := path.Join(c.outputDir, name)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.WriteError(w, utils.NewError("file not found", http.StatusNotFound))
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		utils.WriteError(w, utils.NewError("failed to open file", http.StatusInternalServerError))
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)

	http.ServeFile(w, r, filePath)
}

func (c *controller) processFiles(ctx context.Context, files []*multipart.FileHeader) <-chan domain.ProcessedImage {
	results := make(chan domain.ProcessedImage, len(files))

	var wg sync.WaitGroup

	for _, file := range files {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				path, err := c.service.CompressImage(ctx, file)
				results <- domain.ProcessedImage{Path: path, Error: err}
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
