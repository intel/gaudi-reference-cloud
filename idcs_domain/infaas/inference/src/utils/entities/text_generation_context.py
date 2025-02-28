from dataclasses import dataclass
from typing import Callable, Dict, Any


@dataclass
class TextGenerationContext:
    infer: Callable
    cast_response: Callable
    cast_response_payload: Dict[str, Any]
    unsafe_prompt_streamer: Callable
