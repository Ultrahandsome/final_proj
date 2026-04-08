#!/usr/bin/env python3
import asyncio
import json
import os
import threading
import traceback
from concurrent import futures

import comment_pb2
import comment_pb2_grpc
import faiss
import google.generativeai as genai
import grpc
import numpy as np
from google.generativeai import GenerativeModel
from sentence_transformers import SentenceTransformer

# Configure Gemini model
genai.configure(api_key="AIzaSyAWq584lfThByief9ZaH49YiXx2tmi1at0")
# Note: Creating the model instance here isn't strictly necessary; we recreate it in the async function

# Load FAISS index and embedding model
index_file = "faiss_index.bin"
document_file = "data/comments_labeled.json"

with open(document_file, 'r', encoding='utf-8') as f:
    data = json.load(f)

comments = [item["comment"] for item in data]
global_ids = [item.get("global_id", "") for item in data]
comment_labels = [item["label"] for item in data]
embedder = SentenceTransformer("all-MiniLM-L6-v2")

# Build or load FAISS index
if os.path.exists(index_file):
    index = faiss.read_index(index_file)
else:
    embeddings = embedder.encode(comments, normalize_embeddings=True)
    embeddings = np.array(embeddings).astype(np.float32)
    dim = embeddings.shape[1]
    index = faiss.IndexFlatIP(dim)
    index.add(embeddings)
    faiss.write_index(index, index_file)


def retrieve_similar_comments(query_texts, k=3):
    """
    Given a list of query strings, return for each a list of up to k similar examples:
    [ {"comment": str, "label": str}, ... ]
    """
    q_embeds = embedder.encode(query_texts, normalize_embeddings=True)
    q_embeds = np.array(q_embeds).astype(np.float32)
    distances, indices = index.search(q_embeds, k)
    results = []
    for qi, idxs in enumerate(indices):
        hits = []
        for rank, idx in enumerate(idxs):
            score = float(distances[qi][rank])
            if idx >= 0 and idx < len(comments) and score >= 0.55:
                hits.append({
                    "comment": comments[idx],
                    "label": comment_labels[idx]
                })
        results.append(hits)
    return results


def build_prompt_with_similar(comments_batch, contexts):
    sections = []
    for i, (comment, examples) in enumerate(zip(comments_batch, contexts), start=1):
        if examples:
            ex_lines = [f'- "{ex["comment"]}" → Label: {ex["label"]}' for ex in examples]
            block = "\n".join(ex_lines)
        else:
            block = "No similar examples were retrieved; please assess the risk level on your own"
        sections.append(
            f"Comment {i}:\n\"{comment}\"\n\n" +
            f"Similar Examples:\n{block}"
        )
    body = "\n\n---\n\n".join(sections)
    prompt = f"""
You are a professional content‑moderation expert.
Classify the following student comments **using similar examples as guidance**
Labels reflect tone/safety, not topic. Use exactly one of:
- Safe,
- Needs Review,
- Harassment,
- Swearing,
- Hate Speech,
- Sarcasm,
- Aggressive,
- Complaint,
- Constructive Feedback,
- Other

2. Do NOT use topic‑based labels.
3. If confidence < 0.50, label as Needs Review.
4. For each comment, output exactly one JSON object with keys:
   - "id": string
   - "label": string
   - "confidence": float
   - "keywords": []
Return a valid JSON array only.

{body}
""".strip()
    return prompt


async def classify_comment_batch_async(comments_batch):
    # Retrieve similar contexts
    contexts = retrieve_similar_comments(comments_batch)
    prompt = build_prompt_with_similar(comments_batch, contexts)
    print("Calling Gemini API...")
    # create a fresh model instance per call
    model = GenerativeModel("gemini-2.0-flash")
    response = await asyncio.to_thread(
        model.generate_content,
        contents=prompt,
        generation_config={"temperature": 0.2, "max_output_tokens": 9000}
    )
    print("Gemini API returned.")
    text = response.candidates[0].content.parts[0].text
    # strip markdown
    json_str = text.strip().lstrip("```json").rstrip("```")
    parsed = json.loads(json_str)
    # attach contexts
    for i, itm in enumerate(parsed):
        itm["similar_examples"] = contexts[i]
    return parsed

# Start asyncio event loop in background thread
loop = asyncio.new_event_loop()
def _start_loop(loop):
    asyncio.set_event_loop(loop)
    loop.run_forever()
threading.Thread(target=_start_loop, args=(loop,), daemon=True).start()

class CommentClassifier(comment_pb2_grpc.CommentClassifierServicer):
    def ClassifyComments(self, request, context):
        print("Received classification request...")
        raw_comments = request.comments
        batch_size = 50
        total = len(raw_comments)
        print(f"[DEBUG] Total {total} comments, batch_size={batch_size}")
        for batch_idx, start in enumerate(range(0, total, batch_size), start=1):
            end = min(start + batch_size, total)
            batch = raw_comments[start:end]
            ids = [c.id for c in batch]
            comments_batch = [c.rawComment for c in batch]
            print(f"[DEBUG] Processing batch {batch_idx} (items {start} to {end})")
            try:
                future = asyncio.run_coroutine_threadsafe(
                    classify_comment_batch_async(comments_batch), loop
                )
                parsed_data = future.result()
                print(f"[DEBUG] Completed batch {batch_idx}")
            except Exception as e:
                print(f"Error during classification: {e}")
                traceback.print_exc()
                # yield error responses
                for cid in ids:
                    yield comment_pb2.ClassifiedComments(
                        comments=[
                            comment_pb2.ClassifiedComment(
                                id=cid,
                                label="Error",
                                score=0.0,
                                similarComment=[],
                                keywords=[]
                            )
                        ]
                    )
                continue
            # serialize each result
            for cid, item in zip(ids, parsed_data):
                # convert similar_examples to list[str]
                sims = [ex.get("comment", "") for ex in item.get("similar_examples", [])]
                kws  = list(item.get("keywords", []))
                cc = comment_pb2.ClassifiedComment(
                    id=cid,
                    label=item.get("label", "Unknown"),
                    score=round(item.get("confidence", 0.0), 2),
                    similarComment=sims,
                    keywords=kws,
                )
                yield comment_pb2.ClassifiedComments(comments=[cc])


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    comment_pb2_grpc.add_CommentClassifierServicer_to_server(CommentClassifier(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    print("gRPC server running on port 50051...")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
