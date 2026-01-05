from rank_bm25 import BM25Okapi
from typing import List

corpus: List[List[str]] = [
    "Agent J is the fresh recruit with attitude".split(),
    "Agent K has years of MIB experience and a cool neuralyzer".split(),
    "The galaxy is saved by two Agents in black suits".split(),
]

# 1. 构建 BM25 索引
bm25 = BM25Okapi(corpus)

# 2. 对一个有趣的查询执行检索
query = "Who is a recruit?".split()
top_n = bm25.get_top_n(query, corpus, n=2)
print("Query:", " ".join(query))
print("Top matching lines:")
for line in top_n:
    print(" •", " ".join(line))