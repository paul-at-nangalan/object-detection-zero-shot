package main

import (
	"flag"
	"fmt"
	"github.com/paul-at-nangalan/errorhandler/handlers"
	"github.com/paul-at-nangalan/json-config/cfg"
	"log"
	"object-detection-zero-shot/service"
	"os"
)

func RunOnce(imagefile string, labels, url, apikey string, mode service.OperationMode) {
	/// First get text embeddings
	payload, err := service.CreateDetectionPayload(imagefile, labels, mode)
	handlers.PanicOnError(err)
	data, err := service.RunDetector(payload, url, apikey)
	handlers.PanicOnError(err)
	embeddings, found := data["embeddings"].([]interface{})
	if !found {
		log.Panicln("No embedding data. ", data)
	}
	for _, embedding := range embeddings {
		fmt.Println()
		fmt.Println("LEN", len(embedding.([]interface{})))
		fmt.Println()
		fmt.Println(embedding)
		fmt.Println()
	}

}

/**
{
	"Items": [
		{"ImageFile": "", "Label": ""},
		{"ImageFile": "", "Label": ""}
	]
}
*/

type Item struct {
	Imagefile string
	Label     string
}

type EmbedCfg struct {
	Items []Item
}

func (e *EmbedCfg) Expand() {
	for i, item := range e.Items {
		e.Items[i].Imagefile = os.ExpandEnv(item.Imagefile)
		e.Items[i].Label = os.ExpandEnv(item.Label)
	}
}

func main() {

	imagepath := ""
	embedding := false
	embeddingcfg := ""

	flag.StringVar(&imagepath, "image-file", "", "The filename with the image to process")
	flag.StringVar(&embeddingcfg, "cfg", "", "Path to cfg dir")
	flag.BoolVar(&embedding, "embed", false, "Generate embeddings for the images or text")
	flag.Parse()

	apikey := os.ExpandEnv("$HF_APITOKEN")
	url := os.ExpandEnv("$HF_OBJ_DETECTION_URL")
	cfg.Setup(embeddingcfg)

	embeddings := EmbedCfg{}
	err := cfg.Read("embeddings", &embeddings)
	handlers.PanicOnError(err)

	if embedding {
		for _, item := range embeddings.Items {
			/// First get text embeddings
			fmt.Println("Text embeddings: ")
			RunOnce("", item.Label, url, apikey, service.OPMODE_TEXT_EMBED)

			/// Then get image embeddings
			/// We must split the image filenames ourselves
			fmt.Println("Image embeddings: ")
			RunOnce(item.Imagefile, "", url, apikey, service.OPMODE_IMAGE_EMBED)
		}
		return
	}

}
