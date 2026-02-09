.PHONY: install-deps install-pjsua2 check-pjsua2 install-dev venv test lint clean

VENV := .venv
PYTHON := $(VENV)/bin/python
PIP := $(VENV)/bin/pip

# ── System deps (Ubuntu/Debian) ─────────────────────────────────
install-deps:
	sudo apt-get update -qq
	sudo apt-get install -y build-essential python3-dev swig \
		libasound2-dev libopus-dev libssl-dev

# ── pjsua2 Python bindings ──────────────────────────────────────
install-pjsua2: | $(VENV)
	$(PIP) install pjsua2

check-pjsua2: | $(VENV)
	$(PYTHON) scripts/check_pjsua2.py

# ── Dev environment ─────────────────────────────────────────────
venv:
	test -d $(VENV) || python3 -m venv $(VENV)
	$(PIP) install --upgrade pip

install-dev: | $(VENV)
	$(PIP) install -e '.[test,dev]'

# ── Quality ─────────────────────────────────────────────────────
test: | $(VENV)
	$(PYTHON) -m pytest tests/

lint: | $(VENV)
	$(PYTHON) -m ruff check src/

# ── Cleanup ─────────────────────────────────────────────────────
clean:
	rm -rf build/ dist/ *.egg-info src/*.egg-info
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name '*.pyc' -delete 2>/dev/null || true
