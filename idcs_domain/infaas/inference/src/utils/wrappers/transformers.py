from src.utils.entities.tokenization_result import TokenizationResult

from transformers import AutoTokenizer, TensorType
from typing import List, Dict, Any


class HfTokenizer:

    def __init__(self, model_id: str) -> None:
        self._is_ready = False
        self._model_id = model_id
        self._tokenizer = AutoTokenizer.from_pretrained(model_id)
        self._is_ready = True

    @property
    def model_id(self) -> str:
        return self._model_id

    @property
    def is_ready(self) -> bool:
        return self._is_ready

    @property
    def eos_token(self) -> str:
        return self._tokenizer.eos_token

    @property
    def eos_token_id(self) -> int:
        return self._tokenizer.eos_token_id

    def tokenize(self, text: str) -> List[TokenizationResult]:

        tokens = self._tokenizer.tokenize(text)
        ids = self._tokenizer.convert_tokens_to_ids(tokens)

        # The original tokes list has tokenizer-specific tokens
        # for example:
        #   'Ċ' for '\n', 'Ġ' for ' ' --> llama
        #   '▁' for ' ', ['▁', '<0x0A>'] for '\n' --> mistral
        # by decoding the id directly we're getting the token
        # "print version" with user-ready representation
        tokenization_result = [
            TokenizationResult(
                id=id,
                text=self._tokenizer.decode(id)
            )
            for id in ids
        ]
        return tokenization_result

    def apply_chat_template(
        self,
        conversation: List[Dict[str, str]],
        chat_template: str | None = None,
        add_generation_prompt: bool = False,
        tokenize: bool = True,
        padding: bool = False,
        truncation: bool = False,
        max_length: int | None = None,
        return_tensors: str | TensorType | None = None,
        return_dict: bool = False,
        **tokenizer_kwargs: Any
    ) -> (str | List[int]):

        return self._tokenizer.apply_chat_template(
            conversation=conversation,
            chat_template=chat_template,
            add_generation_prompt=add_generation_prompt,
            tokenize=tokenize,
            padding=padding,
            truncation=truncation,
            max_length=max_length,
            return_tensors=return_tensors,
            return_dict=return_dict,
            **tokenizer_kwargs
        )

    def count_tokens(self, messages: List[Dict[str, str]]) -> int:
        tokenized_messages = self.apply_chat_template(messages)
        return len(tokenized_messages)
