package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/pkg/vips"
)

func main() {
	vips.Startup(nil)
	defer vips.Shutdown()

	http.HandleFunc("/", imageHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:8088", nil))
}

type Resize struct {
	width  int
	height int
}

func (resize *Resize) init(values []string) (*Resize, error) {
	if len(values) < 2 {
		return nil, fmt.Errorf("invalid resize value")
	}

	var errW, errH error
	resize.width, errW = strconv.Atoi(values[0])
	resize.height, errH = strconv.Atoi(values[1])

	if errW != nil || errH != nil || resize.width <= 0 || resize.height <= 0 {
		return nil, fmt.Errorf("resize values must be positive integers")
	}

	return resize, nil
}

func (resize *Resize) doResize(reader io.Reader, writer io.Writer, errChan chan error) {
	_, _, err := vips.NewTransform().
		Load(reader).
		ResizeStrategy(vips.ResizeStrategyStretch).
		Resize(resize.width, resize.height).
		Output(writer).
		Apply()

	errChan <- err
}

type Crop struct {
	left   int
	top    int
	width  int
	height int
}

func (crop *Crop) init(values []string) (*Crop, error) {
	if len(values) < 4 {
		return nil, fmt.Errorf("invalid resize value")
	}

	var errL, errT, errW, errH error
	crop.left, errL = strconv.Atoi(values[0])
	crop.top, errT = strconv.Atoi(values[1])
	crop.width, errW = strconv.Atoi(values[2])
	crop.height, errH = strconv.Atoi(values[3])

	if errL != nil || errT != nil || errH != nil || errW != nil || crop.left < 0 || crop.top < 0 ||
		crop.width <= 0 || crop.height <= 0 {
		return nil, fmt.Errorf("crop values must be positive integers")
	}

	return crop, nil
}

func (crop *Crop) doCrop(reader io.Reader, writer io.Writer, errChan chan error) {
	_, _, err := vips.NewTransform().
		Load(reader).
		ResizeStrategy(vips.ResizeStrategyCrop).
		CropOffsetY(crop.top).
		CropOffsetX(crop.left).
		Resize(crop.width, crop.height).
		Output(writer).
		Apply()

	errChan <- err
}

func imageHandler(res http.ResponseWriter, r *http.Request) {
	var err error
	query := r.URL.Query()
	querySource := query.Get("source")

	if len(querySource) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte(fmt.Sprintf("source parameter is required")))
		return
	}

	resize := &Resize{}
	resize, err = resize.init(strings.Split(query.Get("resize"), "x"))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte(err.Error()))
		return
	}

	crop := &Crop{}
	crop, err = crop.init(strings.Split(query.Get("crop"), "x"))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte(err.Error()))
		return
	}

	fileReader, err := os.Open(querySource)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(fmt.Sprintf("failed to get %s: %v", querySource, err)))
		return
	}

	pipeReader, pipeWriter := io.Pipe()
	errChan := make(chan error, 2)

	go crop.doCrop(fileReader, pipeWriter, errChan)
	go resize.doResize(pipeReader, res, errChan)

	if err = <-errChan; err != nil {
		_ = pipeWriter.Close()
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(fmt.Sprintf("failed to crop %s: %v", querySource, err)))
		return
	}

	_ = pipeWriter.Close()

	if err = <-errChan; err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(fmt.Sprintf("failed to resize %s: %v", querySource, err)))
		return
	}
}
