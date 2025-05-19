# Object Detection Zero-Shot
A zero-shot image classification and detection service using OpenAI's ViT-CLIP model through Hugging Face's inference engine, with Pinecone vector database integration for efficient similarity search.
**Note, this is for demonstration purposes only.**

## API Endpoints
### 1. Image Embedding (`/image/embed`)
Creates vector embeddings for images and their associated text labels, storing them in Pinecone.
**Request Format:**
```
http
POST /image/embed
Content-Type: multipart/form-data

image: <image_file>
text: <text_description>
```

**Response:**
```
json
{
    "status": "success",
    "id": "<generated_id>"
}
```

The endpoint:
- Generates both image and text embeddings using the CLIP model
- Stores embeddings in Pinecone with unique IDs
- Maintains separate vectors for image and text with prefixes "img-" and "text-"
- Rate limited to 30 requests per 24 hours per IP

### 2. Image Detection (`/image/detect`)
Performs zero-shot object detection on images by comparing them against stored embeddings.
**Request Format:**
```
http
POST /image/detect
Content-Type: multipart/form-data

image: <image_file>

```

**Response:**
```
json
{
    "score": <similarity_score>,
    "id": "<vector_id>",
    "label": "<matched_label>"
}
```

The endpoint:
- Generates embeddings for the input image
- Searches Pinecone for similar vectors
- Returns the best match
- Rate limited to 30 requests per 24 hours per IP

### Curl examples
# 1. Image Embed Endpoint
```
curl -X POST https://nesasia.io/image/embed \
-H "Content-Type: multipart/form-data" \
-F "image=@/path/to/your/image.jpg" \
-F "text=description of the image"
```
# Example Response:
# {
#     "status": "success",
#     "id": "generated-id"
# }

# 2. Image Detection Endpoint
```
curl -X POST https://nesasia.io/image/detect \
-H "Content-Type: multipart/form-data" \
-F "image=@/path/to/your/image.jpg"
```
# Example Response:
# {
#     "found": true,
#     "score": 0.85,
#     "label": "matched label",
# }

## Architecture

### Hugging Face Integration

The service uses a custom handler for the Hugging Face inference API, supporting three operation modes:
- `OPMODE_TEXT_EMBED`: Generates text embeddings
- `OPMODE_IMAGE_EMBED`: Generates image embeddings
- `OPMODE_MAINOBJECT`: Detects main objects in images


### Vector Database

Pinecone serves as the vector database, storing and searching high-dimensional embeddings:
- Namespace-based organization
- Efficient similarity search
- Metadata storage for labels
- Upsert operations for vector management

## Further Reading
For more information about zero-shot image classification using CLIP:
[Zero-Shot Image Classification with CLIP](https://www.pinecone.io/learn/series/image-search/zero-shot-image-classification-clip/)
See [handler.py](https://github.com/paul-at-nangalan/object-detection-zero-shot/blob/main/handler.py) for the huggingface custom handler.
For details on creating a custom handler for the inference endpoint, visit:
[HuggingFace create custom handler](https://huggingface.co/docs/inference-endpoints/en/guides/custom_handler).

## Environment Variables
Required environment variables:
- `HF_APITOKEN`: Hugging Face API token
- `HF_OBJ_DETECTION_URL`: Hugging Face model endpoint URL
- `PC_APIKEY`: Pinecone API key
- `PC_HOST`: Pinecone host
- `PC_NAMESPACE`: Pinecone namespace


## License
MIT License - See LICENSE file for details