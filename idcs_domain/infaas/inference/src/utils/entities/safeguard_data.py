from asyncio import Task
from typing import List, Dict


class SafeguardData:

    def __init__(self, raw_prompt: str) -> None:

        self._raw_prompt = raw_prompt
        self._messages = None
        self._task = None

    @property
    def raw_prompt(self) -> str:
        return self._raw_prompt

    @property
    def task(self) -> str:
        return self._task

    @property
    def messages(self) -> List[Dict[str, str]]:
        return self._messages

    def set_messages(self, messages: List[Dict[str, str]]) -> None:
        self._messages = messages

    def set_task(self, task: Task) -> None:
        self._task = task
