from typing import Dict
from agent.base import BaseAgent


class AgentRegistry:
    def __init__(self):
        self._agents: Dict[str, BaseAgent] = {}

    def register(self, agent: BaseAgent):
        if "." not in agent.agent_name:
            raise RuntimeError(f"Agent '{agent.agent_name}' already registered")
        
        if agent.agent_name in self._agents:
            raise RuntimeError(f"Agent {agent.agent_name}")
        
        self._agents[agent.agent_name] = agent

    def get(self, full_name: str) -> BaseAgent:
        if full_name not in self._agents:
            raise KeyError(f"Agent '{full_name}' not found")
        return self._agents[full_name]
    
    def list(self):
        return list(self._agents.keys())


registry = AgentRegistry()
