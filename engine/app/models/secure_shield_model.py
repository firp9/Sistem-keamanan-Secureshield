import joblib
import numpy as np

class SecureShieldModel:
    def __init__(self):
        self.ddos_model = None
        self.sqli_model = None
        self.webshell_model = None
        self.webshell_vectorizer = None

    def load_ddos_model(self, model_path):
        self.ddos_model = joblib.load(model_path)

    def load_sqli_model(self, model_path):
        self.sqli_model = joblib.load(model_path)

    def load_webshell_model(self, model_path, vectorizer_path):
        self.webshell_model = joblib.load(model_path)
        self.webshell_vectorizer = joblib.load(vectorizer_path)

    def predict_ddos(self, df):
        if self.ddos_model is None:
            raise Exception("DDOS model not loaded")
        prediction = self.ddos_model.predict(df)
        proba = self.ddos_model.predict_proba(df)
        return prediction, proba

    def predict_sqli(self, df):
        if self.sqli_model is None:
            raise Exception("SQLi model not loaded")
        prediction = self.sqli_model.predict(df)
        proba = self.sqli_model.predict_proba(df)
        return prediction, proba

    def predict_webshell(self, texts):
        if self.webshell_model is None or self.webshell_vectorizer is None:
            raise Exception("Webshell model or vectorizer not loaded")
        X = self.webshell_vectorizer.transform(texts)
        prediction = self.webshell_model.predict(X)
        proba = self.webshell_model.predict_proba(X)
        return prediction, proba
