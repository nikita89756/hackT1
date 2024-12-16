from fastapi import FastAPI, UploadFile, Form, HTTPException, Body
from pathlib import Path
from pydantic import BaseModel
import model
import uvicorn
import os
import faiss_init

headers = {"Authorization": "Bearer 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"}

# Инициализация FastAPI
app = FastAPI(
    title="FAISS Text Processing API",
    description="API для обработки текстовых файлов с помощью FAISS",
    version="1.0.0"
)


@app.get("/")
def start():
    return {"message": "API работает"}


@app.post("/create_faiss_index")
async def create_faiss_index(
    model_name: str = Form(...),
    chat_id: str = Form(...),
    file: UploadFile = Form(...)
):
    """
    Эндпоинт для создания FAISS индекса с использованием faiss_init.
    """

    _, embedding_model_name, _ = model.model_name_configuration(model_name)

    try:
        # Создание временной папки для файла
        temp_dir = Path("temp_files")
        temp_dir.mkdir(parents=True, exist_ok=True)

        # Путь для сохранения временного файла
        file_path = temp_dir / file.filename

        # Сохранение файла
        file_content = await file.read()
        if not file_content.strip():
            raise HTTPException(status_code=400, detail="Файл пуст")

        with open(file_path, "wb") as temp_file:
            temp_file.write(file_content)

        # Вызов faiss_init с временным файлом
        faiss_init.faiss_init_method(embedding_model_name, chat_id, str(file_path))

        # Удаление временного файла
        os.remove(file_path)

        return {"message": f"FAISS индекс успешно создан для chat_id {chat_id}!"}
    
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Ошибка при создании FAISS индекса: {str(e)}")

class QRequest(BaseModel):
    question: str
    chat_id: int
    model_name: str

@app.post("/answering")
async def answering(request: QRequest = Body(...)):
    """ 
    Эндпоинт для ответа на вопрос
    """
    try:
        answer = model.main(request.model_name, request.question, str(request.chat_id))
        return {"answer": answer}
    except Exception as e:
        return {"error": str(e)}


# Запуск приложения
if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8089)
