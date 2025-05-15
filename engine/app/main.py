import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from fastapi import FastAPI, WebSocket, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import uvicorn
from .predict_ddos import router as ddos_router
from .predict_sqli import router as sqli_router
from .predict_webshell import router as webshell_router
import logging
import asyncio

app = FastAPI()

logging.basicConfig(filename='engine.log', level=logging.INFO, format='%(asctime)s %(message)s')

detected_events = []
websocket_connections = set()
websocket_lock = asyncio.Lock()

engine_monitoring_active = False

@app.post("/api/engine/activate")
async def activate_engine():
    global engine_monitoring_active
    engine_monitoring_active = True
    return {"status": "engine monitoring activated"}

@app.post("/api/engine/deactivate")
async def deactivate_engine():
    global engine_monitoring_active
    engine_monitoring_active = False
    return {"status": "engine monitoring deactivated"}

class Event(BaseModel):
    type: str
    timestamp: str
    content: str
    source: str

@app.get("/status")
async def status():
    return {"status": "SecureShield Engine running"}

@app.post("/api/events")
async def receive_event(event: Event):
    logging.info(f"Received event: {event}")
    detected_events.append(event.dict())

    # Broadcast to all connected websocket clients
    async with websocket_lock:
        to_remove = set()
        for ws in websocket_connections:
            try:
                await ws.send_json(event.dict())
            except Exception as e:
                logging.error(f"Error sending websocket message: {e}")
                to_remove.add(ws)
        websocket_connections.difference_update(to_remove)

    return JSONResponse(content={"status": "success", "message": "event received"})

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    await websocket.accept()
    async with websocket_lock:
        websocket_connections.add(websocket)
    try:
        while True:
            data = await websocket.receive_text()
            logging.info(f"WebSocket received data: {data}")
            await websocket.send_text(f"Received: {data}")
    except Exception as e:
        logging.error(f"WebSocket connection error: {e}")
    finally:
        async with websocket_lock:
            websocket_connections.discard(websocket)

app.include_router(ddos_router, prefix="/api/predict/ddos")
app.include_router(sqli_router, prefix="/api/predict/sqli")
app.include_router(webshell_router, prefix="/api/predict/webshell")

from fastapi.responses import StreamingResponse
import io
import csv
import json

@app.get("/api/report")
async def get_report(format: str = "json"):
    """
    Endpoint to get detected events report in JSON or CSV format.
    Query param 'format' can be 'json' or 'csv'.
    """
    if format not in ("json", "csv"):
        return JSONResponse(status_code=400, content={"error": "Invalid format. Use 'json' or 'csv'."})

    if format == "json":
        return JSONResponse(content=detected_events)

    # CSV format
    output = io.StringIO()
    writer = csv.DictWriter(output, fieldnames=["type", "timestamp", "content", "source"])
    writer.writeheader()
    for event in detected_events:
        writer.writerow(event)
    output.seek(0)
    return StreamingResponse(iter([output.getvalue()]), media_type="text/csv", headers={"Content-Disposition": "attachment; filename=report.csv"})

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
