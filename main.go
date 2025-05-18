package main

import (
	"flag"
	"fmt"
	"github.com/paul-at-nangalan/errorhandler/handlers"
	"github.com/paul-at-nangalan/json-config/cfg"
	"log"
	"net/http"
	"object-detection-zero-shot/embedding"
	"object-detection-zero-shot/service"
	"object-detection-zero-shot/vectordb"
	"object-detection-zero-shot/webfront"
	"os"
)

func main() {

	imagepath := ""
	emb := false
	embeddingcfg := ""
	runservice := false

	flag.StringVar(&imagepath, "image-file", "", "The filename with the image to try and detect")
	flag.StringVar(&embeddingcfg, "cfg", "", "Path to cfg dir")
	flag.BoolVar(&emb, "embed", false, "Generate embeddings for the images or text")
	flag.BoolVar(&runservice, "service", false, "Run as a service")
	flag.Parse()

	apikey := os.ExpandEnv("$HF_APITOKEN")
	url := os.ExpandEnv("$HF_OBJ_DETECTION_URL")

	pcapikey := os.ExpandEnv("$PC_APIKEY")
	pchost := os.ExpandEnv("$PC_HOST")
	pcnamespace := os.ExpandEnv("$PC_NAMESPACE")
	cfg.Setup(embeddingcfg)

	if runservice {

		// Get required environment variables
		apikey := os.Getenv("HF_APITOKEN")
		url := os.Getenv("HF_OBJ_DETECTION_URL")
		pcapikey := os.Getenv("PC_APIKEY")
		pchost := os.Getenv("PC_HOST")
		pcnamespace := os.Getenv("PC_NAMESPACE")
		uploadDir := os.Getenv("UPLOAD_DIR")
		certfile := os.Getenv("CERTFILE")
		keyfile := os.Getenv("KEYFILE")
		if apikey == "" || url == "" || pcapikey == "" || pchost == "" || pcnamespace == "" || uploadDir == "" || certfile == "" || keyfile == "" {
			log.Fatal("Missing required environment variables")
		}
		// Create the embedder
		embedder := embedding.NewEmbedder(url, apikey)
		// Create the Pinecone DB connection
		pc := vectordb.NewPineconeDB(pchost, pcapikey, pcnamespace)
		// Create the service handler
		svc := service.NewHandler(embedder, pc)
		// Create the web frontend handler
		_ = webfront.NewHandler(svc, uploadDir)
		// Start the HTTPS server
		port := os.Getenv("PORT")
		if port == "" {
			port = "443"
		}
		fmt.Printf("Starting server on port %s...\n", port)
		err := http.ListenAndServeTLS(":"+port, certfile, keyfile, nil)
		handlers.PanicOnError(err)
		return
	}

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
