from typing import Dict, List, Any
from transformers import CLIPProcessor, CLIPModel
import torch

from transformers import pipeline
from PIL import Image
from io import BytesIO
import base64

class EndpointHandler():
    def __init__(self, path=""):
        self.processor = CLIPProcessor.from_pretrained(path)
        self.model = CLIPModel.from_pretrained(path)

        # if you have CUDA set it to the active device like this
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        # move the model to the device
        self.model.to(self.device)


    def process_image(self, image: Any, candidate_labels: List[str]) -> Dict[str, float]:

        # Prepare text inputs
        input = self.processor(
            images=image,
            text=None,
            return_tensors="pt",
            padding=True
        )['pixel_values'].to(self.device)

        # Get model predictions
        img_emb = self.model.get_image_features(input)
        img_emb = img_emb.detach().cpu().numpy()
        img_emb
        return img_emb

    def get_text_embeddings(self, candidate_labels: List[str]) -> Dict[str, float]:
        # Prepare text inputs
        input = self.processor(
            images=None,
            text=candidate_labels,
            return_tensors="pt",
            padding=True
        )['input_ids'].to(self.device)

        # Get model predictions
        text_emb = self.model.get_text_features(input)
        text_emb = text_emb.detach().cpu().numpy()
        return {'embeddings': text_emb}

    def get_image_embeddings(self, image: Any) -> Dict[str, float]:
        # Prepare text inputs
        input = self.processor(
            images=image,
            text=None,
            return_tensors="pt",
            padding=True
        )['pixel_values'].to(self.device)

        # Get model predictions
        img_emb = self.model.get_image_features(input)
        img_emb = img_emb.detach().cpu().numpy()
        return {'embeddings': img_emb}

    ### Find the main object in the image and return the vector of that object
    def find_main_object(self, image: Any) -> Dict[str, float]:
        # Prepare text inputs
        input = self.processor(
            images=image,
            text=None,
            return_tensors="pt",
            padding=True
        )['pixel_values'].to(self.device)

        # Get model predictions
        img_emb = self.model.get_image_features(input)
        img_emb = img_emb.detach().cpu().numpy()
        return {'embeddings': img_emb}


    def __call__(self, data: Dict[str, Any]) -> List[Dict[str, Any]]:
        """

        """
        inputs = data.pop("inputs", data)

        msgtype = inputs['type']
        ### get embeddings for text or image
        if msgtype == 'get-embeddings':
            # get embeddings
            if inputs['mode'] == 'text':
                results = self.get_text_embeddings(inputs['candidates'])
                return results

            elif inputs['mode'] == 'image':
                image = Image.open(BytesIO(base64.b64decode(inputs['image'])))
                results = self.get_image_embeddings(image)
                return results

            else:
                raise ValueError("Invalid mode. Use 'text' or 'image'.")
        ### find the main object in the image and return the vector of that object
        elif msgtype == 'find-main-object':
            image = Image.open(BytesIO(base64.b64decode(inputs['image'])))
            results = self.find_main_object(image)
            return results

        ### process an image and return the vector of each object detected - this can be used to lookup image candidates

        ##image = Image.open(BytesIO(base64.b64decode(inputs['image'])))
        ##results = self.process_image(image, candidate_labels=inputs["candidates"])
        raise ValueError("Invalid mode. Use 'get-embeddings' or 'find-main-object'.")