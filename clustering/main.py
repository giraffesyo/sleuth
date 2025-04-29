#!/usr/bin/env python
import pymongo, torch, hdbscan, numpy as np
from sentence_transformers import SentenceTransformer
from tqdm import tqdm
import os
# get MONGODB_URI from environment
MONGODB_URI = os.getenv("MONGODB_URI")

client = pymongo.MongoClient(MONGODB_URI)
coll   = client.sleuth.articles

def text_of(doc):
    bits = [
        doc.get("title", ""),
        doc.get("description", ""),
        doc.get("date", ""),
        " ".join(doc.get("victimNames", []))
    ]
    return " ".join(filter(None, bits))
docs   = list(coll.find({}, {"_id": 1, "title": 1, "description": 1, "victimNames": 1, "date": 1}))
ids    = [d["_id"] for d in docs]
txts   = [text_of(d) for d in docs]

model  = SentenceTransformer("all-mpnet-base-v2",
                             device="cuda" if torch.cuda.is_available() else "cpu")
embeds = model.encode(txts, batch_size=256, normalize_embeddings=True)
embeds = np.vstack(embeds).astype("float32")

clusterer = hdbscan.HDBSCAN(
    metric                  = "euclidean",      # OK → unit vectors
    min_cluster_size        = 2,                # ↓ catch 2-article bursts
    min_samples             = 1,                # less strict core rule
    cluster_selection_method= "leaf",           # work at leaf level
    cluster_selection_epsilon = 0.02            # split clusters whose
                                                # centroids differ by ≥0.02
).fit(embeds)

import collections
print(collections.Counter(clusterer.labels_))
# e.g. Counter({-1: 138, 5: 12, 8: 7, 1: 6, 9: 6, 3: 3, ...})

bulk = [
    pymongo.UpdateOne({"_id": _id},
                      {"$set": {"clusterId": int(lbl),
                                "clusterConf": float(prob),
                                "embed": embeds[i].tolist()}})
    for i, (_id, lbl, prob) in enumerate(zip(ids, clusterer.labels_, clusterer.probabilities_))
]
coll.bulk_write(bulk, ordered=False)
print(f"Inserted clusterIds for {len(bulk)} docs")