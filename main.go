package main

import (
	"flag"
	"github.com/paul-at-nangalan/errorhandler/handlers"
	"object-detection-zero-shot/service"
	"os"
)

func main() {

	imagepath := ""
	labels := ""

	flag.StringVar(&imagepath, "image-file", "", "The filename with the image to process")
	flag.StringVar(&labels, "labels", "", "Comma seperated list of candidate objects")
	flag.Parse()

	apikey := os.ExpandEnv("$HF_APITOKEN")
	url := os.ExpandEnv("$HF_OBJ_DETECTION_URL")

	payload, err := service.CreateDetectionPayload(imagepath, labels)
	handlers.PanicOnError(err)

	err = service.RunDetector(payload, url, apikey)
	handlers.PanicOnError(err)
}
