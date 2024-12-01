package controller

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/SergeyBogomolovv/image-compressor/internal/domain"
	"github.com/SergeyBogomolovv/image-compressor/pkg/utils"
)

type Service interface {
	CompressImage(ctx context.Context, header *multipart.FileHeader) (string, error)
}

type controller struct {
	service Service
}

func Register(router *http.ServeMux, service Service) {
	controller := &controller{service: service}
	router.HandleFunc("POST /upload", controller.UploadHandler)
	router.HandleFunc("GET /{id}", controller.GetHandler)
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

func (*controller) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println(id)
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
