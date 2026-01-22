.PHONY: install install-dev run test clean screenshot help

help:
	@echo "Bearing TUI - Development Commands"
	@echo ""
	@echo "  make install      Install the package"
	@echo "  make install-dev  Install with dev dependencies"
	@echo "  make run          Run the TUI"
	@echo "  make test         Run tests"
	@echo "  make screenshot   Generate screenshot.svg"
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

clean:
	rm -rf build/ dist/ *.egg-info/
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
