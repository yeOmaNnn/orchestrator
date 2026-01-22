from agent.decorators import agent


@agent("math.add")
async def add(input):
    return {
        "result": input.get("a", 0) + input.get("b", 0)
    }
