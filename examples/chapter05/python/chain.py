from dotenv import load_dotenv
load_dotenv()
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate

# Create LLM and prompt template
llm = ChatOpenAI(model_name="gpt-4o-mini", temperature=0)
prompt = ChatPromptTemplate.from_template("Answer this question: {input}")

# LCEL chain using pipes:
chain = prompt | llm

# Invoke the chain
result = chain.invoke({"input": "What is the capital of France?"})
print(result.content)