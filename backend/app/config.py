from pydantic import BaseSettings

class Settings(BaseSettings):
    MONGO_URI: str = "mongodb://mongo:27017"
    MONGO_DB_NAME: str = "testdb"
    
    class Config:
        env_file = ".env"

settings = Settings()
