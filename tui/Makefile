.PHONY: install install-dev run test clean screenshot screenshots help

help:
	@echo "Bearing TUI - Development Commands"
	@echo ""
	@echo "  make install      Install the package"
	@echo "  make install-dev  Install with dev dependencies"
	@echo "  make run          Run the TUI"
	@echo "  make test         Run tests"
	@echo "  make screenshot   Generate single screenshot.svg"
	@echo "  make screenshots  Generate all visual regression screenshots"
	@echo "  make clean        Remove build artifacts"

install:
	pip install -e .

install-dev:
	pip install -e ".[dev]"

run:
	bearing-tui

test:
	pytest -v

screenshot:
	BEARING_WORKSPACE=~/Projects bearing-tui --screenshot assets/screenshot.svg
	@echo "Screenshot saved to assets/screenshot.svg"

screenshots:
	@echo "Generating visual regression screenshots..."
	python3 scripts/generate_screenshots.py
	@echo "Screenshots saved to screenshots/"

screenshots-check:
	@echo "Checking screenshots are up to date..."
	python3 scripts/generate_screenshots.py
	@if git diff --exit-code screenshots/; then \
		echo "✓ Screenshots are up to date"; \
	else \
		echo "✗ Screenshots changed! Run 'make screenshots' and commit."; \
		exit 1; \
	fi

clean:
	rm -rf build/ dist/ *.egg-info/
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
