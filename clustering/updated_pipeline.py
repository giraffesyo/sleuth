import os
import pymongo
import torch
import hdbscan
import numpy as np
from sentence_transformers import SentenceTransformer
import collections

# --- CONFIG ---
CONF_THRESHOLD = 0.9

# --- DB SETUP ---
MONGODB_URI = os.getenv("MONGODB_URI")
client = pymongo.MongoClient(MONGODB_URI)
coll = client.sleuth.articles


# --- UTILS ---
def get_device():
    if torch.cuda.is_available():
        return "cuda"
    if torch.backends.mps.is_available():
        return "mps"
    return "cpu"


def build_text(doc):
    """
    Build a single text string for embedding:
      - title, description, date
      - concatenated snippets from relevantTimestamps
    """
    parts = [doc.get("title", ""), doc.get("description", ""), doc.get("date", "")]
    events = doc.get("relevantTimestamps", [])
    for e in events:
        snippet = e.get("text_snippet", "").strip()
        loc = e.get("location", "").strip()
        time_d = e.get("time_detail", "").strip()
        # append only non-empty components
        if snippet:
            parts.append(snippet)
        if loc:
            parts.append(loc)
        if time_d:
            parts.append(time_d)
    return " ".join(parts)


# --- LOAD DATA ---
docs = list(
    coll.find(
        {},
        {
            "_id": 1,
            "title": 1,
            "description": 1,
            "date": 1,
            "bodyDiscoveryEvents.text_snippet": 1,
            "bodyDiscoveryEvents.location": 1,
            "bodyDiscoveryEvents.time_detail": 1,
        },
    )
)
ids = [d["_id"] for d in docs]
texts = [build_text(d) for d in docs]

# --- EMBEDDING ---
print("Encoding embeddings…")
model = SentenceTransformer("all-mpnet-base-v2", device=get_device())
embeds = model.encode(texts, batch_size=256, normalize_embeddings=True)
# HDBSCAN’s linkage expects float64
embeds = np.vstack(embeds).astype(np.float64)

# --- CLUSTERING ---
print("Clustering with HDBSCAN (cosine)…")
clusterer = hdbscan.HDBSCAN(
    metric="cosine",
    algorithm="generic",
    min_cluster_size=2,
    min_samples=2,
    cluster_selection_method="leaf",
    cluster_selection_epsilon=0.02,
).fit(embeds)

# Quick stats
print("Cluster label counts:", collections.Counter(clusterer.labels_))

# --- UPDATE DB ---
print("Writing cluster results to MongoDB…")
ops = []
for idx, doc_id in enumerate(ids):
    lbl = int(clusterer.labels_[idx])
    conf = float(clusterer.probabilities_[idx])
    # decide caseId: if noise or low-confidence, mark as -1
    case_id = lbl if (lbl != -1 and conf >= CONF_THRESHOLD) else -1
    ops.append(
        pymongo.UpdateOne(
            {"_id": doc_id},
            {
                "$set": {
                    "clusterId": lbl,
                    "clusterConf": conf,
                    "caseId": case_id,
                    "embed": embeds[idx].tolist(),
                }
            },
        )
    )

if ops:
    coll.bulk_write(ops, ordered=False)
    print(f"Updated {len(ops)} documents with clusterId/clusterConf/caseId/embed.")
else:
    print("No updates to write.")
