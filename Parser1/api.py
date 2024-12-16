"""
FastAPI сервер для сервиса парсинга документов.
Поддерживает парсинг веб-страниц, PDF и DOCX файлов.
"""

from fastapi import FastAPI, HTTPException, UploadFile, File
from pydantic import BaseModel
from Parser import DocumentHandle
import os
from typing import Optional
import PyPDF2
from io import BytesIO
import tempfile
import shutil

app = FastAPI(
    title="Document Parser API",
    description="API для парсинга веб-страниц, PDF и DOCX файлов",
    version="1.0.0"
)

class SourceRequest(BaseModel):
    """Модель запроса для парсинга документа."""
    source: str
    output_path: Optional[str] = None

class ParseResponse(BaseModel):
    """Модель ответа с результатом парсинга."""
    content: str
    output_path: Optional[str] = None

@app.post("/parse", 
         response_model=ParseResponse,
         description="Парсинг контента из URL или локального файла")
def parse_document(request: SourceRequest):
    """
    Парсинг документа из URL или пути к файлу.
    
    Аргументы:
        request: Объект SourceRequest, содержащий источник и опциональный путь для сохранения
        
    Возвращает:
        Объект ParseResponse с распарсенным контентом и путем сохранения
        
    Вызывает:
        HTTPException: Если парсинг не удался
    """
    try:
        handle = DocumentHandle(request.source)
        content = handle.get_content()
        
        output_path = None
        if request.output_path:
            output_path = handle.save_as_txt(request.output_path)
        else:
            with tempfile.NamedTemporaryFile(delete=False, suffix='.txt') as tmp:
                output_path = tmp.name
                handle.save_as_txt(output_path)
            os.remove(output_path)
        
        return ParseResponse(
            content=content,
            output_path=output_path
        )
    except Exception as e:
        raise HTTPException(
            status_code=500, 
            detail=f"Ошибка обработки документа: {str(e)}"
        )

@app.post("/parse_pdf", 
         response_model=ParseResponse,
         description="Парсинг контента из загруженного PDF файла")
async def parse_pdf_file(file: UploadFile = File(...)):
    """
    Парсинг загруженного PDF файла.
    
    Аргументы:
        file: Загруженный PDF файл
        
    Возвращает:
        ParseResponse: Объект с извлеченным текстом
        
    Вызывает:
        HTTPException: Если парсинг не удался или файл не PDF
    """
    if not file.filename.lower().endswith('.pdf'):
        raise HTTPException(
            status_code=400,
            detail="Загруженный файл должен быть PDF"
        )
        
    try:
        content = await file.read()
        handle = DocumentHandle("temp.pdf")  # имя не важно, файл не будет сохранен
        text = handle._convert_pdf_to_text(content)
        
        if not text:
            raise HTTPException(
                status_code=500,
                detail="Не удалось извлечь текст из PDF"
            )
            
        return ParseResponse(content=text)
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Ошибка обработки PDF: {str(e)}"
        )

@app.post("/convert_pdf_to_text",
         response_model=ParseResponse,
         description="Конвертация PDF файла в текст без сохранения")
async def convert_pdf_to_text(file: UploadFile = File(...)):
    """
    Конвертация загруженного PDF файла в текст без сохранения файла.
    
    Аргументы:
        file: Загруженный PDF файл
        
    Возвращает:
        ParseResponse: Объект с извлеченным текстом
        
    Вызывает:
        HTTPException: Если конвертация не удалась или файл не PDF
    """
    if not file.filename.lower().endswith('.pdf'):
        raise HTTPException(
            status_code=400,
            detail="Загруженный файл должен быть PDF"
        )
        
    try:
        content = await file.read()
        handle = DocumentHandle("temp.pdf")  # имя не важно, файл не будет сохранен
        text = handle._convert_pdf_to_text(content)
        
        if not text:
            raise HTTPException(
                status_code=500,
                detail="Не удалось извлечь текст из PDF"
            )
            
        return ParseResponse(content=text)
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Ошибка конвертации PDF: {str(e)}"
        )

@app.post("/parse_docx", 
         response_model=ParseResponse,
         description="Парсинг контента из загруженного DOCX файла")
async def parse_docx_file(file: UploadFile = File(...)):
    """
    Парсинг загруженного DOCX файла.
    
    Аргументы:
        file: Загруженный DOCX файл
        
    Возвращает:
        ParseResponse: Объект с извлеченным текстом
        
    Вызывает:
        HTTPException: Если парсинг не удался или файл не DOCX
    """
    if not file.filename.lower().endswith('.docx'):
        raise HTTPException(
            status_code=400,
            detail="Загруженный файл должен быть DOCX"
        )
        
    try:
        content = await file.read()
        handle = DocumentHandle("temp.docx")  # имя не важно, файл не будет сохранен
        text = handle._convert_docx_to_text(content)
        
        if not text:
            raise HTTPException(
                status_code=500,
                detail="Не удалось извлечь текст из DOCX"
            )
            
        return ParseResponse(content=text)
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Ошибка обработки DOCX: {str(e)}"
        )

@app.get("/health")
def health_check():
    """Проверка работоспособности сервиса."""
    return {"status": "ok"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
