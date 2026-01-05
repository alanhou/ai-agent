from dotenv import load_dotenv
load_dotenv()
from typing import Annotated
from typing_extensions import TypedDict
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_core.vectorstores import InMemoryVectorStore
from langchain_core.documents import Document
from langgraph.graph import StateGraph, MessagesState, START

llm = ChatOpenAI(model="gpt-5")

def call_model(state: MessagesState):
   response = llm.invoke(state["messages"])
   return {"messages": response}

# Use LangChain's standard InMemoryVectorStore
# This avoids the "mutex lock failed" crash on macOS associated with 'vectordb'
embeddings = OpenAIEmbeddings()
memory = InMemoryVectorStore(embeddings)

text = """Machine learning is a method of data analysis that automates analytical model building. It is a branch of artificial intelligence based on the idea that systems can learn from data, identify patterns and make decisions with minimal human intervention. Machine learning algorithms are trained on datasets that contain examples of the desired output. For example, a machine learning algorithm that is used to classify images might be trained on a dataset that contains images of cats and dogs. Once an algorithm is trained, it can be used to make predictions on new data. For example, the machine learning algorithm that is used to classify images could be used to predict whether a new image contains a cat or a dog."""

metadata = {"title": "Introduction to Machine Learning", "url": "https://learn.microsoft.com/en-us/training/modules/" + 
    "introduction-to-machine-learning"}

memory.add_documents([Document(page_content=text, metadata=metadata)])

text2 = """Artificial intelligence (AI) is the simulation of human intelligence in machinesthat are programmed to think like humans and mimic their actions.The term may also be applied to any machine that exhibits traits associated witha human mind such as learning and problem-solving.AI research has been highly successful in developing effective techniques for solving a wide range of problems, from game playing to medical diagnosis."""

metadata2 = {"title": "Artificial Intelligence for Beginners", "url": "https://microsoft.github.io/AI-for-Beginners"}

memory.add_documents([Document(page_content=text2, metadata=metadata2)])

query = "What is the relationship between AI and machine learning?"
results = memory.similarity_search(query, k=3)

builder = StateGraph(MessagesState)
builder.add_node("call_model", call_model)
builder.add_edge(START, "call_model")
graph = builder.compile()

input_message = {"type": "user", "content": "hi! I'm bob"}
for chunk in graph.stream({"messages": [input_message]}, stream_mode="values"):
   chunk["messages"][-1].pretty_print()

print("\nSearch Results:")
for doc in results:
    print(f"- {doc.page_content[:100]}... (Metadata: {doc.metadata})")