# 🏃 Running pytest tests

* NOTE: If you don't have poetry please see main repo README.md for poetrt env setup.

1. Running all the tests:
```bash
poetry run pytest
```

2. In order to run a specific test file, user `poetry run pytest` with the relative file path in the `tests` folder:
```bash
poetry run pytest tests/path/to/file
```

* For example, if you want to run a test for `test_unsafe_prompt_response.py` run:
```bash
poetry run pytest tests/src/utils/streamers/test_unsafe_prompt_response.py
```

3. Enable console prints is possible with adding `-s` to the running command:
```bash
poetry run pytest -s
poetry run pytest -s tests/src/utils/streamers/test_unsafe_prompt_response.py
```