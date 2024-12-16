import requests
from langchain_community.vectorstores import FAISS
from sentence_transformers import SentenceTransformer
from langchain.embeddings.base import Embeddings
import time


def model_name_configuration(model: str):
        if model == 'RobertaBS2':
            model_name = "deepset/roberta-base-squad2"
            embedding_model_name = "sentence-transformers/all-MiniLM-L6-v2"
            delay = 10

        return model_name, embedding_model_name, delay


def main(model, question, chat_id):

    def model_name_configuration(model: str):
        if model == 'RobertaBS2':
            model_name = "deepset/roberta-base-squad2"
            embedding_model_name = "sentence-transformers/all-MiniLM-L6-v2"
            delay = 10

        return model_name, embedding_model_name, delay

    model_name, embedding_model_name, delay = model_name_configuration(model)

    API_URL = f'https://api-inference.huggingface.co/models/{model_name}'
    headers = {"Authorization": "Bearer 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"}

    # Настройка модели эмбеддингов
    class SentenceTransformerEmbeddings(Embeddings):
        """
        Обёртка для модели SentenceTransformer для интеграции с FAISS.
        """
        def __init__(self, model_name):
            self.model = SentenceTransformer(model_name)

        def embed_documents(self, texts):
            return self.model.encode(texts, convert_to_tensor=False)

        def embed_query(self, text):
            return self.embed_documents([text])[0]

    embedding_model = SentenceTransformerEmbeddings(embedding_model_name)


    def query_huggingface_api(context, question, delay):
        """
        Отправка запроса к API Hugging Face для задач QA с задержкой.
        
        Args:
            context (str): Контекст для ответа на вопрос.
            question (str): Вопрос для ответа.
            delay (int): Время задержки в секундах перед отправкой запроса.
        """
        payload = {
            "inputs": {
                "question": question,
                "context": context
            },
            "parameters": {
                "max_length": 1024,
                "temperature": 1.0,
                "top_p": 0.9,
                "do_sample": True
            }
        }

        print(f"Ожидание {delay} секунд перед отправкой запроса...")
        time.sleep(delay)  # Задержка перед отправкой запроса

        print("Отправка запроса к API Hugging Face...")
        response = requests.post(API_URL, headers=headers, json=payload)
        print(f"HTTP статус: {response.status_code}")

        if response.status_code != 200:
            raise ValueError(f"API Error {response.status_code}: {response.text}")
        
        full_response = response.json()
        print("Полный ответ от API:", full_response)
        
        if "answer" in full_response:
            return full_response["answer"]
        else:
            raise ValueError("Ответ от API не содержит ключа 'answer'.")


    def faiss_founder(chat_id: str, embedding_model):# 
        faiss_index_path = f'faiss/faiss_{chat_id}'
        print("Попытка загрузить существующий индекс...")
        vectorstore = FAISS.load_local(faiss_index_path, embedding_model, allow_dangerous_deserialization=True)
        print(f"Индекс загружен.")
        retriever = vectorstore.as_retriever()

        return retriever

    # RAG Fusion Pipeline
    def rag_fusion_pipeline(question, chat_id: str, max_context_tokens=1024):
        """
        Извлечение релевантного контекста из базы знаний.
        """
        retriever = faiss_founder(chat_id, embedding_model)
        relevant_documents = retriever.invoke(question, k=3)
        relevant_documents_content = [doc.page_content for doc in relevant_documents]
        combined_context = "\n".join(relevant_documents_content)

        if len(combined_context.split()) > max_context_tokens:
            combined_context = " ".join(combined_context.split()[:max_context_tokens])
            
        return combined_context

    def rag_chain(question, chat_id):
        """
        Основная цепочка RAG: извлечение контекста и генерация ответа.
        """
        context = rag_fusion_pipeline(question, chat_id, max_context_tokens=1024)

        # Отправка запроса в API QA
        print(f"Длина контекста в токенах: {len(context.split())}")
        answer = query_huggingface_api(context, question, delay)
        print(f"Ответ от модели Hugging Face:\n{answer}")
        return answer
    
    return rag_chain(question, chat_id)
