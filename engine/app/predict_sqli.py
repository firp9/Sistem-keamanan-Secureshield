from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
import pandas as pd
import numpy as np
from models.secure_shield_model import SecureShieldModel

router = APIRouter()
import os

model = SecureShieldModel()
model_path = os.path.join(os.path.dirname(__file__), '..', 'models', 'model_sqli.joblib')
model.load_sqli_model(model_path)

class SQLiPredictRequest(BaseModel):
    features: dict

class PredictResponse(BaseModel):
    label: str
    probability: float

@router.post("/", response_model=PredictResponse)
async def predict_sqli(request: SQLiPredictRequest):
    try:
        df = pd.DataFrame([request.features])
        prediction, proba = model.predict_sqli(df)
        label = prediction[0]
        probability = float(np.max(proba))
        return PredictResponse(label=label, probability=probability)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
