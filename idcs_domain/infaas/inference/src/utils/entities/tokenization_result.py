from dataclasses import dataclass


@dataclass
class TokenizationResult:
    id: int
    text: str
