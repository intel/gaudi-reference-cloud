from dataclasses import dataclass
from typing import Dict, Any


@dataclass
class ServerProbData:
    url: str
    payload: Dict[str, Any]
