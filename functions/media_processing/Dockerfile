FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
RUN python -c "from megadetector.detection.run_detector import load_detector; print('megadetector import ok')"

COPY . .

EXPOSE 8080

CMD exec uvicorn app:app --host 0.0.0.0 --port ${PORT:-8080}
