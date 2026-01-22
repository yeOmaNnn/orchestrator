from agent.decorators import agent 

@agent("text.echo")
async def echo(input):
    return {
        "text": input.get("text", "")
    }

@agent("text.upper")
async def upper(input):
    return {
        "text": input.get("text", "").upper()
    }