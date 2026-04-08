# ───────────────────────────────────────────────────────────────────────────────
# 1) Builder: install build‑tools, download & wheel everything
# ───────────────────────────────────────────────────────────────────────────────
FROM python:3.10-slim AS builder

WORKDIR /wheels
ENV DNNL_DISABLE_JIT=1 \
    PYTHONUNBUFFERED=1

# install any build deps you need (e.g. for C extensions)
RUN apt-get update \
 && apt-get install -y --no-install-recommends build-essential \
 && rm -rf /var/lib/apt/lists/*

# copy only requirements to leverage layer caching
COPY requirements.txt .

# build wheels for all your deps (including the extra)
RUN pip wheel --no-cache-dir --wheel-dir . \
        -r requirements.txt \
    && pip wheel --no-cache-dir --wheel-dir . \
        "huggingface_hub[hf_xet]" hf_xet


# ───────────────────────────────────────────────────────────────────────────────
# 2) Final: pull in just Python & the pre‑built wheels
# ───────────────────────────────────────────────────────────────────────────────
FROM python:3.10-slim

WORKDIR /app
ENV DNNL_DISABLE_JIT=1 \
    PYTHONUNBUFFERED=1

# copy the built wheels from the builder stage
COPY --from=builder /wheels /wheels
COPY requirements.txt .

# install from our local wheel cache, never hitting pypi
RUN pip install --no-cache-dir \
        --no-index \
        --find-links /wheels \
        -r requirements.txt \
    && pip install --no-cache-dir \
        --no-index \
        --find-links /wheels \
        "huggingface_hub[hf_xet]" hf_xet

# copy your app code last so code changes don’t bust deps cache
COPY . .

EXPOSE 50051
CMD ["python", "-u", "AI_grpc.py"]

