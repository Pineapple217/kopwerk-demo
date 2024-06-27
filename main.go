package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
	"github.com/xitongsys/parquet-go-source/buffer"
	"github.com/xitongsys/parquet-go/reader"

	_ "github.com/joho/godotenv/autoload"
)

var albums = make(map[int64]Album)

func main() {
	e := echo.New()
	e.HideBanner = true

	e.Use(echoMw.RequestLoggerWithConfig(echoMw.RequestLoggerConfig{
		LogStatus:  true,
		LogURI:     true,
		LogMethod:  true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, v echoMw.RequestLoggerValues) error {
			slog.Info("request",
				"method", v.Method,
				"status", v.Status,
				"latency", v.Latency,
				"path", v.URI,
			)
			return nil

		},
	}))
	url := os.Getenv("PARQUIT_URL")
	if url == "" {
		panic("PARQUIT_URL env not set")
	}
	LoadParquet(url)

	e.GET("/health", HandleHealth)
	e.GET("/album/:id", HandleGetAlbum)
	e.GET("/albums", HandleGetAlbums)

	e.Start("0.0.0.0:4000")
}

type Album struct {
	Id     int64  `parquet:"name=Id, type=INT64"`
	Year   int64  `parquet:"name=Year, type=INT64"`
	Album  string `parquet:"name=Album, type=BYTE_ARRAY, convertedtype=UTF8"`
	Artist string `parquet:"name=Artist, type=BYTE_ARRAY, convertedtype=UTF8"`
}

func LoadParquet(url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("failed to fetch parquet file")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	pf, err := buffer.NewBufferFile(b)
	if err != nil {
		panic(err)
	}

	pr, err := reader.NewParquetReader(pf, nil, 4)
	if err != nil {
		panic(err)
	}

	num := int(pr.GetNumRows())
	a := make([]Album, num)
	pr.Read(&a)

	for _, album := range a {
		albums[album.Id] = album
	}

	slog.Info("load parquet file", "row_count", num)
}

type Health struct {
	Status string `json:"status"`
}

func HandleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, Health{
		Status: "OK!",
	})
}

func HandleGetAlbum(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	a, ok := albums[id]
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, a)
}

func HandleGetAlbums(c echo.Context) error {
	return c.JSON(http.StatusOK, albums)
}
