from typing import Dict, Any, Awaitable, Callable
from agent.base import BaseAgent
from agent.registry import registry


def agent(name: str):
    def wrapper(func: Callable[[Dict[str, Any]], Awaitable[Dict[str, Any]]]):
        class FuncAgent(BaseAgent):
            agent_name = name

            async def run(self, input: Dict[str, Any]) -> Dict[str, Any]:
                return await func(input)

        registry.register(FuncAgent())
        return func

    return wrapper
