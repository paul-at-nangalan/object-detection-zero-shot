package service

import (
	"fmt"
	"github.com/paul-at-nangalan/errorhandler/handlers"
	"log"
	"object-detection-zero-shot/embedding"
	"object-detection-zero-shot/vectordb"
	"os"
)

type Handler struct {
	clipmodel  *embedding.Embedder
	pineconedb *vectordb.PineconeDB
}

func NewHandler(clipmodel *embedding.Embedder, pineconedb *vectordb.PineconeDB) *Handler {
	return &Handler{
		clipmodel:  clipmodel,
		pineconedb: pineconedb,
	}
}

func (h *Handler) getEmbedding(imagefile string, labels string, mode embedding.OperationMode) []float32 {
	/// First get text embeddings
	payload, err := embedding.CreateDetectionPayload(imagefile, labels, mode)
	handlers.PanicOnError(err)
	data, err := h.clipmodel.Do(payload)
	handlers.PanicOnError(err)

	emb := data["embeddings"].([]interface{})

	vector := make([]float32, len(emb))
	for i, val := range emb {
		vector[i] = float32(val.(float64))
	}
	return vector
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
	ID        string
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

func (h *Handler) EmbedData(embeddings *EmbedCfg) {

	for _, item := range embeddings.Items {
		if item.ID == "" {
			log.Panicln("Each item must have a non empty ID")
		}
		/// First get text embeddings
		fmt.Println("Text embeddings: ")
		txtembedding := h.getEmbedding("", item.Label, embedding.OPMODE_TEXT_EMBED)

		/// Then get image embeddings
		/// We must split the image filenames ourselves
		fmt.Println("Image embeddings: ")
		imgembedding := h.getEmbedding(item.Imagefile, "", embedding.OPMODE_IMAGE_EMBED)
		txtid := "text-" + item.ID
		metadata := map[string]interface{}{
			"value": item.Label, //// don't store image data here
		}
		err := h.pineconedb.UpsertVector(txtembedding, txtid, metadata)
		handlers.PanicOnError(err)

		imgid := "img-" + item.ID
		err = h.pineconedb.UpsertVector(imgembedding, imgid, metadata)
		handlers.PanicOnError(err)
	}

}

func (h *Handler) ImageDetection(imagefile string) {

	payload, err := embedding.CreateDetectionPayload(imagefile, "", embedding.OPMODE_MAINOBJECT)
	handlers.PanicOnError(err)

	data, err := h.clipmodel.Do(payload)
	handlers.PanicOnError(err)

	emb := data["embeddings"].([]interface{})

	vector := make([]float32, len(emb))
	for i, val := range emb {
		vector[i] = float32(val.(float64))
	}
	results, err := h.pineconedb.SearchVectors(vector, 20)
	handlers.PanicOnError(err)

	for _, result := range results {
		fmt.Println()
		fmt.Println(result.Score, "=>", result.Metadata)
		fmt.Println()
	}
}
