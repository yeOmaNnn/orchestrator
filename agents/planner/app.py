from fastapi import FastAPI
from pydantic import BaseModel
from typing import List, Dict 

app = FastAPI()

class PlanStep(BaseModel):
    step: int
    agent: str 
    description: str 
    depend_on: List[int] = []
    input: Dict = {}

class PlanResponse(BaseModel):
    plan: List[PlanStep]

@app.post("/run", response_model=PlanResponse)
def run_planner(payload: dict):
    return {
        "plan": [
            {
                "step": 1, 
                "agent": "research_agent", 
                "description": "Collect marker data", 
                "input": {"query": payload.get("goal")}
            }
        ]
    }