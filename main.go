package main

import (
	"flag"
	"fmt"
	"github.com/paul-at-nangalan/errorhandler/handlers"
	"github.com/paul-at-nangalan/json-config/cfg"
	"object-detection-zero-shot/embedding"
	"object-detection-zero-shot/service"
	"object-detection-zero-shot/vectordb"
	"os"
)

func main() {

	imagepath := ""
	emb := false
	embeddingcfg := ""

	flag.StringVar(&imagepath, "image-file", "", "The filename with the image to try and detect")
	flag.StringVar(&embeddingcfg, "cfg", "", "Path to cfg dir")
	flag.BoolVar(&emb, "embed", false, "Generate embeddings for the images or text")
	flag.Parse()

	apikey := os.ExpandEnv("$HF_APITOKEN")
	url := os.ExpandEnv("$HF_OBJ_DETECTION_URL")

	pcapikey := os.ExpandEnv("$PC_APIKEY")
	pchost := os.ExpandEnv("$PC_HOST")
	pcnamespace := os.ExpandEnv("$PC_NAMESPACE")
	cfg.Setup(embeddingcfg)

	/// If embedding from disk data
	embeddings := service.EmbedCfg{}
	err := cfg.Read("embeddings", &embeddings)
	handlers.PanicOnError(err)

	embedder := embedding.NewEmbedder(url, apikey)
	pc := vectordb.NewPineconeDB(pchost, pcapikey, pcnamespace)
	svc := service.NewHandler(embedder, pc)

	if emb {
		svc.EmbedData(&embeddings)
	} else {
		results := svc.ImageDetection(imagepath)
		for _, result := range results {
			fmt.Println()
			fmt.Println(result.Score, "[", result.ID, "] =>", result.Metadata)
			fmt.Println()
		}
	}
}
