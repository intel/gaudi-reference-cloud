from src.utils.factories.backends import(
    BackendFactory, BackendOptions
)
from src.utils.wrappers.backends.openai import OpenAiWrapper
from src.utils.wrappers.backends.tgi import TgiWrapper
from tests.consts import TestConsts


def test_tgi_backend_factory(server_config) -> None:
    """Tests backend factory for TGI."""

    llm_engine, sg_engine = BackendFactory.backend_wrapper(
        config=server_config,
        backend=BackendOptions.TGI,
        base_url=TestConsts.BASE_URL,
        safeguard_url=TestConsts.SAFEGUARD_URL,
        headers=None,
        cookies=None,
        timeout=None
    )
    
    assert isinstance(llm_engine, TgiWrapper)
    assert isinstance(sg_engine, TgiWrapper)

def test_openai_backend_factory(server_config) -> None:
    """Tests backend factory for vLLM (using the OpenAI sdk)."""

    llm_engine, sg_engine = BackendFactory.backend_wrapper(
        config=server_config,
        backend=BackendOptions.VLLM,
        base_url=TestConsts.BASE_URL,
        safeguard_url=TestConsts.SAFEGUARD_URL,
        headers=None,
        cookies=None,
        timeout=None
    )

    assert isinstance(llm_engine, OpenAiWrapper)
    assert isinstance(sg_engine, OpenAiWrapper)
