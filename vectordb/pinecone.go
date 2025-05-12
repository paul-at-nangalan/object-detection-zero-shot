package vectordb

import (
	"context"
	"fmt"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
)

type PineconeDB struct {
	host      string
	apiKey    string
	namespace string
}

func NewPineconeDB(host string,
	apiKey string,
	namespace string,
) *PineconeDB {
	if namespace == "" {
		log.Panicln("Namespace cannot be empty")
	}
	return &PineconeDB{
		host:      host,
		apiKey:    apiKey,
		namespace: namespace,
	}
}

// UpsertVector uploads a single vector to Pinecone DB
func (p *PineconeDB) UpsertVector(
	vectorValues []float32,
	vectorID string,
	metadata map[string]interface{},
) error {
	// Create context
	ctx := context.Background()
	// Initialize Pinecone client
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: p.apiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create Pinecone client: %v", err)
	}

	// Create index connection
	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{
		Host:      p.host,
		Namespace: p.namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to create index connection: %v", err)
	}
	// Convert metadata to structpb
	var metadataStruct *structpb.Struct
	if metadata != nil {
		metadataStruct, err = structpb.NewStruct(metadata)
		if err != nil {
			return fmt.Errorf("failed to create metadata struct: %v", err)
		}
	}
	// Create vector
	vector := &pinecone.Vector{
		Id:       vectorID,
		Values:   &vectorValues,
		Metadata: metadataStruct,
	}
	// Upsert vector
	count, err := idxConnection.UpsertVectors(ctx, []*pinecone.Vector{vector})
	if err != nil {
		return fmt.Errorf("failed to upsert vector: %v", err)
	}
	if count != 1 {
		return fmt.Errorf("expected to upsert 1 vector, but upserted %d", count)
	}
	return nil
}

// SearchResult represents a single search result with metadata and score
type SearchResult struct {
	Metadata map[string]interface{} `json:"metadata"`
	Score    float32                `json:"score"`
}

// SearchVectors performs a similarity search in Pinecone DB
func (p *PineconeDB) SearchVectors(
	queryVector []float32,
	topK uint32,
) ([]SearchResult, error) {
	// Create context
	ctx := context.Background()
	// Initialize Pinecone client
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: p.apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone client: %v", err)
	}
	// Create index connection
	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{
		Host:      p.host,
		Namespace: p.namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create index connection: %v", err)
	}
	// Perform the query
	queryResponse, err := idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          queryVector,
		TopK:            topK,
		IncludeValues:   false,
		IncludeMetadata: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %v", err)
	}
	// Process results
	results := make([]SearchResult, 0, len(queryResponse.Matches))
	for _, match := range queryResponse.Matches {
		var metadata map[string]interface{}
		if match.Vector.Metadata != nil {
			metadata = match.Vector.Metadata.AsMap()
		}

		results = append(results, SearchResult{
			Score:    match.Score,
			Metadata: metadata,
		})
	}
	return results, nil
}
