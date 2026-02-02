.PHONY: test test-unit test-integration install clean

install:
	pip install -r requirements.txt
	pip install pytest pytest-asyncio pytest-cov httpx

test: test-unit

test-unit:
	pytest test_linear_client.py -v --tb=short

test-integration:
	@if [ -z "$$LINEAR_API_KEY" ]; then \
		echo "❌ ERROR: LINEAR_API_KEY not set"; \
		echo "   Run: export LINEAR_API_KEY=lin_api_xxxxxxxxxxxxx"; \
		exit 1; \
	fi
	python test_integration.py

test-quick:
	@if [ -z "$$LINEAR_API_KEY" ]; then \
		echo "❌ ERROR: LINEAR_API_KEY not set"; \
		echo "   Run: export LINEAR_API_KEY=lin_api_xxxxxxxxxxxxx"; \
		exit 1; \
	fi
	python test_integration.py --quick

test-coverage:
	pytest test_linear_client.py --cov=linear_client --cov-report=html --cov-report=term
	@echo "📊 Coverage report generated in htmlcov/index.html"

clean:
	rm -rf __pycache__ .pytest_cache htmlcov .coverage *.pyc

help:
	@echo "Available targets:"
	@echo "  make install          - Install dependencies"
	@echo "  make test            - Run unit tests (mocked)"
	@echo "  make test-integration - Run integration tests (requires LINEAR_API_KEY)"
	@echo "  make test-quick       - Quick API connection test"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make clean            - Clean cache files"

