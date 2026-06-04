FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    libxcb1 \
    libgl1 \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender1 \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

# Install provided-model runtime dependencies separately to avoid pip resolver conflicts.
RUN pip install --no-cache-dir "megadetector==10.0.21"
RUN pip install --no-cache-dir "onnx==1.16.2" "onnx2torch" "torch" "torchvision"

RUN python -c "from megadetector.detection.run_detector import load_detector; import torch; import onnx2torch; print('provided model runtime import ok')"

COPY . .

EXPOSE 8080

CMD exec uvicorn app:app --host 0.0.0.0 --port ${PORT:-8080}