FROM python:3.13.6-slim

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/
COPY pyproject.toml uv.lock ./
RUN uv sync --locked
COPY . .
ENV PYTHONUNBUFFERED=1
CMD [".venv/bin/python", "main.py"]
