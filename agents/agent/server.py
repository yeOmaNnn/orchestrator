from fastapi import FastAPI, HTTPException
from agent.registry import registry

app = FastAPI()


@app.get("/")
def health():
    return {
        "status": "ok", 
        "agents": registry.list(), 
            }


@app.post("/agents/{agent_name}")
async def call_agent(agent_name: str, payload: dict):
    try:
        agent = registry.get(agent_name)
    except KeyError:
        raise HTTPException(status_code=404, detail="Agent not found")

    try:
        result = await agent.run(payload.get("input", {}))
        return {"output": result}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
