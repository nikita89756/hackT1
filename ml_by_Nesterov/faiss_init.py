from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.schema import Document
from langchain_community.vectorstores import FAISS
from langchain.embeddings.base import Embeddings
from sentence_transformers import SentenceTransformer
from langchain.document_loaders import TextLoader


class SentenceTransformerEmbeddings(Embeddings):
    def __init__(self, model_name):
        self.model = SentenceTransformer(model_name)

    def embed_documents(self, texts):
        return self.model.encode(texts, convert_to_tensor=False)

    def embed_query(self, text):
        return self.embed_documents([text])[0]


def faiss_init_method(embedding_model_name, chat_id, file_path):
    # Загрузка данных
    text_loader = TextLoader(file_path, encoding='utf-8')  
    data = text_loader.load()

    # Настройка модели эмбеддингов
    embeddings = SentenceTransformerEmbeddings(embedding_model_name)

    # Разбиение текстов на фрагменты
    text_splitter = RecursiveCharacterTextSplitter(chunk_size=1000, chunk_overlap=200)
    texts = text_splitter.split_documents(data)

    # Создание векторного хранилища FAISS
    vectorstore = FAISS.from_documents(texts, embeddings)

    # Сохранение хранилища на диск
    faiss_index_path = f'faiss/faiss_{chat_id}'
    vectorstore.save_local(faiss_index_path)

    print("Векторное хранилище FAISS успешно создано и сохранено!")

