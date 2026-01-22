from agent.decorators import agent


@agent("agent1")
async def echo(input):
    return {
        "result": input.get("text")
    }
