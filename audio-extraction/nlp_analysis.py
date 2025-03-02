import spacy
from sentence_transformers import SentenceTransformer, util

### Testing only

# Load NLP models
nlp = spacy.load('en_core_web_sm') # Load English tokenizer, tagger, parser, NER and word vectors
transformer = SentenceTransformer('paraphrase-MiniLM-L6-v2') # Load Sentence Transformer model for semantic similarity


def is_relevant(txt, keywords, threshold=0.5):
    """Check if a given text is relevant to a list of keywords

    Args:
        txt (str): Text to check
        keywords (list): List of keywords
        threshold (float): Threshold for similarity score
    """
    # doc = nlp(txt)
    # embeddings = transformer.encode([txt])
    # scores = util.cos_sim(embeddings, transformer.encode(keywords))[0]
    # return any(score >= threshold for score in scores)
    pass