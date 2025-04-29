#!/usr/bin/env python3
"""
visualize_clusters.py
---------------------

Projects Sentence-BERT embeddings from your MongoDB `articles` collection
into 2-D (UMAP or t-SNE) and colours the points by `clusterId`.

Dependencies
------------
pip install pymongo numpy umap-learn matplotlib seaborn plotly pandas scikit-learn
"""
import os
import argparse
import sys
from typing import List

import numpy as np
import pymongo
from bson import ObjectId

# --- plotting libraries ------------------------------------------------------
import matplotlib.pyplot as plt
import seaborn as sns

import plotly.express as px
import pandas as pd

# --- dimensionality-reduction choices ---------------------------------------
import umap.umap_ as umap
from sklearn.manifold import TSNE


# -----------------------------------------------------------------------------


def fetch_docs(host: str, db: str, coll: str, limit: int = 0) -> List[dict]:
    client = pymongo.MongoClient(host)
    cursor = (
        client[db][coll]
        .find({}, {"embed": 1, "clusterId": 1, "title": 1, "date": 1})
        .limit(limit)  # 0 ⇒ no limit
    )
    docs = [d for d in cursor if "embed" in d and "clusterId" in d]
    if not docs:
        sys.exit("No documents found that contain both `embed` and `clusterId`")
    return docs


def project(
    embeds: np.ndarray, method: str = "umap", random_state: int = 42
) -> np.ndarray:
    if method == "tsne":
        return TSNE(
            n_components=2,
            metric="cosine",
            perplexity=min(40, (len(embeds) - 1) // 3),
            learning_rate="auto",
            init="pca",
            random_state=random_state,
        ).fit_transform(embeds)

    # default = UMAP
    return umap.UMAP(
        n_neighbors=15,
        min_dist=0.1,
        metric="cosine",
        random_state=random_state,
    ).fit_transform(embeds)


def plot_matplotlib(xy: np.ndarray, labels: np.ndarray, out: str = None) -> None:
    uniq = sorted(set(labels) - {-1})
    palette = sns.color_palette("hls", n_colors=len(uniq))
    colour = {lbl: palette[i] for i, lbl in enumerate(uniq)}
    colour[-1] = (0.6, 0.6, 0.6)  # grey for noise

    fig, ax = plt.subplots(figsize=(9, 7), dpi=110)
    ax.scatter(
        xy[:, 0],
        xy[:, 1],
        s=18,
        c=[colour[lbl] for lbl in labels],
        alpha=0.75,
        linewidths=0,
    )
    ax.set_xticks([]), ax.set_yticks([])
    ax.set_title("Article clusters")

    if out:
        fig.savefig(out, bbox_inches="tight")
        print(f"saved {out}")
    else:
        plt.show()


def plot_plotly(
    xy: np.ndarray, labels: np.ndarray, titles: List[str], dates: List[str], out_html: str
) -> None:
    df = pd.DataFrame(
        {
            "x": xy[:, 0],
            "y": xy[:, 1],
            "cluster": labels,
            "title": titles,
            "date": dates,
        }
    )
    fig = px.scatter(
        df,
        x="x",
        y="y",
        color="cluster",
        hover_data=["title", "date", "cluster"],
        title="Interactive view of article clusters",
        height=750,
    )
    fig.update_traces(marker=dict(size=6, opacity=0.8, line=dict(width=0)))
    fig.write_html(out_html, include_plotlyjs="cdn")
    print(f"wrote interactive plot to {out_html}")




def main() -> None:
    MONGODB_URI = os.getenv("MONGODB_URI")
    p = argparse.ArgumentParser(description="Visualise article clusters.")
    p.add_argument("--mongo", default=MONGODB_URI, help="Mongo URI")
    p.add_argument("--db", default="sleuth", help="database name")
    p.add_argument("--coll", default="articles", help="collection name")
    p.add_argument("--limit", type=int, default=0, help="max docs (0 = all)")
    p.add_argument(
        "--tsne",
        action="store_true",
        help="use t-SNE instead of UMAP for dimensionality reduction",
    )
    p.add_argument("--out", help="write static PNG instead of showing window")
    p.add_argument("--html", help="write interactive Plotly HTML to this file")
    args = p.parse_args()

    docs = fetch_docs(args.mongo, args.db, args.coll, args.limit)
    embeds = np.asarray([d["embed"] for d in docs], dtype=np.float32)
    labels = np.asarray([d["clusterId"] for d in docs])
    titles = [d.get("title", "—") for d in docs]
    dates = [d.get("date", "") for d in docs]

    xy = project(embeds, method="tsne" if args.tsne else "umap")

    if args.html:
        plot_plotly(xy, labels, titles, dates, args.html)
    else:
        plot_matplotlib(xy, labels, out=args.out)


if __name__ == "__main__":
    main()