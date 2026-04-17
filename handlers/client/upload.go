package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	maxUploadSize = 20 << 20 // 20 MB
	uploadsDir    = "./uploads"
)

var allowedMediaTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
	"video/mp4":  ".mp4",
	"video/webm": ".webm",
	"video/quicktime": ".mov",
}

// UploadMedia godoc
// @Summary Upload a photo or video
// @Description Upload a media file (image or video) to attach to a message. Max 20MB.
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /upload [post]
func UploadMedia(baseURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Файл не найден в запросе", nil, nil)
		}

		if file.Size > maxUploadSize {
			return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Файл слишком большой (максимум 20 МБ)", nil, nil)
		}

		contentType := file.Header.Get("Content-Type")
		// Strip charset/params if present
		if idx := strings.Index(contentType, ";"); idx >= 0 {
			contentType = strings.TrimSpace(contentType[:idx])
		}

		ext, ok := allowedMediaTypes[contentType]
		if !ok {
			return specError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Неподдерживаемый тип файла. Разрешены: JPEG, PNG, GIF, WebP, MP4, WebM, MOV", nil, nil)
		}

		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			return specInternal(c)
		}

		filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixMilli(), ext)
		dst := filepath.Join(uploadsDir, filename)

		src, err := file.Open()
		if err != nil {
			return specInternal(c)
		}
		defer src.Close()

		out, err := os.Create(dst)
		if err != nil {
			return specInternal(c)
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			return specInternal(c)
		}

		url := fmt.Sprintf("%s/uploads/%s", strings.TrimRight(baseURL, "/"), filename)
		return specSuccess(c, fiber.Map{"url": url})
	}
}
