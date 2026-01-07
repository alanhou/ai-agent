
import argparse
import copy
import json
import os
import random
import uuid
import sys
from collections import namedtuple
from concurrent.futures import ThreadPoolExecutor
from openai import OpenAI

try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass # Assume env vars are set if dotenv is missing

# Initialize OpenAI Client
client = OpenAI()

# Global config
SEARCHING_MODE = True
PRINT_LLM_DEBUG = False

# Helper Types
Info = namedtuple('Info', ['name', 'author', 'content', 'iteration_idx'])

# --- Utility Functions ---

def random_id():
    return str(uuid.uuid4())[:8]

def bootstrap_confidence_interval(data, num_bootstrap_samples=1000, confidence_level=0.95):
    """Calculates simple mean as fitness (simplified from bootstrap for standalone usage)."""
    if not data:
        return 0.0
    return sum(data) / len(data)

def get_json_response_from_gpt(msg, model, system_message, temperature=0.5):
    try:
        response = client.chat.completions.create(
            model=model,
            messages=[
                {"role": "system", "content": system_message},
                {"role": "user", "content": msg},
            ],
            temperature=temperature,
            response_format={"type": "json_object"}
        )
        content = response.choices[0].message.content
        return json.loads(content)
    except Exception as e:
        print(f"Error in GPT call: {e}")
        return {}

def get_json_response_from_gpt_reflect(msg_list, model, temperature=0.8):
    try:
        response = client.chat.completions.create(
            model=model,
            messages=msg_list,
            temperature=temperature,
            response_format={"type": "json_object"}
        )
        content = response.choices[0].message.content
        return json.loads(content)
    except Exception as e:
        print(f"Error in GPT reflect: {e}")
        return {}

def FORMAT_INST(request_keys):
    return f"""Reply EXACTLY with the following JSON format.\n{json.dumps(request_keys)}\nDO NOT MISS ANY REQUEST FIELDS!"""

def ROLE_DESC(role):
    return f"You are a {role}."

# --- Core ADAS Agent Classes ---

class LLMAgentBase:
    def __init__(self, output_fields: list, agent_name: str,
                 role='helpful assistant', model='gpt-4o', temperature=0.5):
        self.output_fields = output_fields
        self.agent_name = agent_name
        self.role = role
        self.model = model
        self.temperature = temperature
        self.id = random_id()

    def generate_prompt(self, input_infos, instruction, output_description):
        output_desc_str = {k: output_description.get(k, f"Your {k}") for k in self.output_fields}
        system_prompt = ROLE_DESC(self.role) + "\n\n" + FORMAT_INST(output_desc_str)
        
        input_text = ""
        for info in input_infos:
            if isinstance(info, Info):
                input_text += f"### {info.name}:\n{info.content}\n\n"
        
        prompt = input_text + instruction
        return system_prompt, prompt

    def query(self, input_infos: list, instruction, output_description, iteration_idx=-1):
        system_prompt, prompt = self.generate_prompt(input_infos, instruction, output_description)
        response_json = get_json_response_from_gpt(prompt, self.model, system_prompt, self.temperature)
        
        # Ensure fields exist
        for k in self.output_fields:
            if k not in response_json:
                response_json[k] = ""
                
        return [Info(k, self.agent_name + " " + self.id, v, iteration_idx) for k, v in response_json.items()]

    def __call__(self, input_infos, instruction, output_description, iteration_idx=-1):
        return self.query(input_infos, instruction, output_description, iteration_idx)

class AgentSystem:
    def __init__(self):
        pass
    def forward(self, task_info):
        raise NotImplementedError

# --- Supply Chain Task Definition ---

class SupplyChainTask:
    def __init__(self):
        self.scenarios = self._generate_data(100) # 100 scenarios
    
    def _generate_data(self, n):
        data = []
        for _ in range(n):
            # Simple scenario: reorder point calculation
            # Ground truth policy: Order = max(0, Target - (Stock + Incoming))
            target = 100
            stock = random.randint(0, 120)
            incoming = random.randint(0, 50)
            ground_truth = max(0, target - (stock + incoming))
            
            data.append({
                "target": target,
                "stock": stock,
                "incoming": incoming,
                "ground_truth": ground_truth
            })
        return data

    def get_init_archive(self):
        # Initial simple solution (prompt/code)
        return [{
            "name": "manual_heuristic",
            "code": """
def forward(task_info):
    # Extracts data from task info string
    # Data format: "Target: 100, Stock: 50, Incoming: 10"
    import re
    try:
        text = task_info.content
        target = int(re.search(r'Target: (\d+)', text).group(1))
        stock = int(re.search(r'Stock: (\d+)', text).group(1))
        incoming = int(re.search(r'Incoming: (\d+)', text).group(1))
        
        # Simple heuristic: reorder if stock < 50% of target
        if stock < target * 0.5:
            return target - stock
        return 0
    except:
        return 0
""",
        }]

    def get_prompt(self, archive):
        system = "You are an expert AI Engineer."
        prompt = """You need to write a Python function `forward(task_info)` that helps a supply chain manager decide the reorder quantity.
        
The input `task_info` is an object with a `.content` attribute containing a string like:
"Target: 100, Stock: 50, Incoming: 10"

Your function should return a single integer: the reorder quantity.

Goal: Minimize proper stockouts and overstocking. The "Ground Truth" follows a specific logical formula you need to discover or approximate.

Return your solution in JSON:
{
    "thought": "your reasoning",
    "name": "agent_vX",
    "code": "def forward(task_info):\\n    ..."
}
"""
        return system, prompt

    def get_reflexion_prompt(self, prev_solution):
        return (
            f"Review your previous solution which had poor performance.\nSolution: {prev_solution}\nAnalyze why it failed.",
            "Now propose a new, improved solution code."
        )

    def format_task(self, data):
        return f"Target: {data['target']}, Stock: {data['stock']}, Incoming: {data['incoming']}"

    def evaluate_prediction(self, prediction, ground_truth):
        # 1.0 if exactly correct, linearly decreasing with error
        try:
            pred = int(prediction)
            err = abs(pred - ground_truth)
            return max(0, 1 - (err / 50.0)) # Score 0 if off by 50+
        except:
            return 0

    def load_data(self, searching=True):
        if searching:
            return self.scenarios[:20] # Train on 20
        return self.scenarios[20:] # Test on rest

    def prepare_task_queue(self, data):
        return [Info('task', 'User', self.format_task(d), -1) for d in data]

    def parse_prediction(self, res):
        return res

    def get_ground_truth(self, data):
        return data['ground_truth']

# --- Evolution & Search ---

def evaluate_forward_fn(args, forward_str, task):
    namespace = {}
    try:
        exec(forward_str, globals(), namespace)
        func = namespace.get('forward')
    except Exception as e:
        print(f"Syntax/Import Check Failed: {e}")
        return [0]

    if not callable(func):
        return [0]

    # Create a dynamic agent class structure for execution if needed, 
    # but here we just call the function directly for simplicity
    
    data = task.load_data(SEARCHING_MODE)
    task_queue = task.prepare_task_queue(data)
    
    scores = []
    for i, item in enumerate(task_queue):
        try:
            res = func(item)
            gt = task.get_ground_truth(data[i])
            score = task.evaluate_prediction(res, gt)
            scores.append(score)
        except Exception as e:
            scores.append(0)
            
    return scores

def search(args, task):
    print("=== ADAS Supply Chain Optimizer ===")
    archive = task.get_init_archive()
    
    # Eval initial
    print(f"Evaluating Initial Agent: {archive[0]['name']}")
    acc = evaluate_forward_fn(args, archive[0]['code'], task)
    archive[0]['fitness'] = bootstrap_confidence_interval(acc)
    print(f"Initial Fitness: {archive[0]['fitness']:.4f}")

    for i in range(args.n_generation):
        print(f"\n--- Generation {i+1} ---")
        
        # 1. Generate new solution
        sys_p, u_p = task.get_prompt(archive)
        
        # Simple prompt composition
        msg = [{"role": "system", "content": sys_p}, 
               {"role": "user", "content": u_p + f"\n\nPrevious Best Code:\n{archive[-1]['code']}"}]
        
        try:
            response = get_json_response_from_gpt_reflect(msg, args.model)
            code = response.get("code", "")
            
            # 2. Evaluate
            acc = evaluate_forward_fn(args, code, task)
            fitness = bootstrap_confidence_interval(acc)
            
            print(f"Generated: {response.get('name', 'unnamed')}")
            print(f"Fitness: {fitness:.4f}")
            
            # 3. Add to archive if good
            response['fitness'] = fitness
            if fitness > archive[-1]['fitness']:
                print(">>> Improved Solution Found!")
                archive.append(response)
            else:
                print("Solution discarded (no improvement).")
                
        except Exception as e:
            print(f"Generation failed: {e}")

    print("\n=== Best Solution found ===")
    print(archive[-1]['code'])

# --- Main Entry ---

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--model', default='gpt-4o')
    parser.add_argument('--n_generation', type=int, default=3)
    args = parser.parse_args()
    
    task = SupplyChainTask()
    search(args, task)