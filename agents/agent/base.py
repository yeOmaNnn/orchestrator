from abc import ABC, abstractmethod
from typing import Dict, Any


class BaseAgent(ABC):
    agent_name: str

    @abstractmethod
    async def run(self, input: Dict[str, Any]) -> Dict[str, Any]:
        pass
