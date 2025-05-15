from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
import numpy as np
import os
from models.secure_shield_model import SecureShieldModel

router = APIRouter()
model = SecureShieldModel()
model_path = os.path.join(os.path.dirname(__file__), '..', 'models', 'random_forest_webshell_model.joblib')
vectorizer_path = os.path.join(os.path.dirname(__file__), '..', 'models', 'tfidf_vectorizer_webshell.joblib')
model.load_webshell_model(model_path, vectorizer_path)

class WebShellPredictRequest(BaseModel):
    file_content: str

class PredictResponse(BaseModel):
    label: str
    probability: float

@router.post("/", response_model=PredictResponse)
async def predict_webshell(request: WebShellPredictRequest):
    try:
        texts = [request.file_content]
        prediction, proba = model.predict_webshell(texts)
        label = prediction[0]
        probability = float(np.max(proba))
        return PredictResponse(label=label, probability=probability)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
