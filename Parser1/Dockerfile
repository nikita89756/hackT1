FROM python:3.9-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .

RUN pip install -r requirements.txt
RUN pip install fastapi uvicorn

COPY . .

EXPOSE 8000

CMD ["python", "api.py"]
