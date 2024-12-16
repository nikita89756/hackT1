FROM python:3.10-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /ml

COPY reqirements.txt .

RUN pip install -r reqirements.txt
RUN pip install fastapi uvicorn
RUN pip install accelerate>=0.26.0
RUN pip install python-multipart

COPY . .

EXPOSE 8089

CMD ["python", "api.py"]