FROM python:3.13-slim AS base

ENV DEBIAN_FRONTEND=noninteractive \
    PATH="/.venv/bin:/root/.cargo/bin:$PATH" \
    UV_COMPILE_BYTECODE=1 \
    UV_PYTHON_DOWNLOADS=never \
    UV_PYTHON=3.13

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

FROM base AS python-deps

ARG INSTALL_DEV_DEPENDENCIES=false

COPY services/producer/pyproject.toml services/producer/uv.lock ./

RUN pip install uv==0.7.18 && \
    if [ "$INSTALL_DEV_DEPENDENCIES" = "true" ]; then \
        uv sync --frozen; \
    else \
        uv sync --frozen --no-dev; \
    fi

FROM base AS runtime

COPY --from=python-deps ./.venv ./.venv
ENV PATH="/.venv/bin:$PATH"

RUN useradd --create-home appuser
USER appuser
WORKDIR /home/appuser

COPY services/producer/main.py ./

EXPOSE 8080

CMD ["python3", "-m", "main"]