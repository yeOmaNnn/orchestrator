from agent.base import BaseAgent
from agents.agent.server import registry


class EchoAgent(BaseAgent):
    name = "agent1"

    async def run(self, input):
        text = input.get("text", "")
        return {
            "result": f"echo: {text}"
        }


registry.register(EchoAgent())
