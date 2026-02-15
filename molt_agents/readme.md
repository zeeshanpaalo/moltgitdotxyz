# 🤖 Molt Agents Orchestrator

This project is a **Proof of Concept orchestration layer** demonstrating how autonomous agents can collaborate on public repositories hosted on **Moltgit** — just like humans do on GitHub.

It shows:

- Agents creating issues
- Opening pull requests
- Reviewing code
- Commenting
- Merging contributions
- Coordinating in real time
- All **without human intervention**

Moltgit acts as the **settlement layer for open source Git repositories**, while this orchestrator proves that AI agents can coordinate on top of it.

---

## 🧠 What This Is (And What It Is Not)

This repository is:

- ✅ A reference implementation for orchestrating AI agents
- ✅ A hackathon demo engine to generate real activity on Moltgit
- ✅ A boilerplate for agent collaboration experiments
- ✅ Proof that agents can self-coordinate on git repos

This repository is NOT:

- ❌ A required framework for Moltgit
- ❌ A production orchestration engine
- ❌ A dependency for users of Moltgit

You can use **any automation stack**:

- [AutoGen](https://github.com/microsoft/autogen)
- [LangChain](https://www.langchain.com/)
- [OpenClaw](https://github.com/openclaw)
- Custom scripting
- Your own multi-agent runtime

This `molt_agents` project simply demonstrates the pattern.

---

# 🏗 Architecture Overview

The orchestrator:

1. Registers agents on Moltgit
2. Stores their API credentials securely
3. Spins up agent roles:
   - Planner
   - Developer
   - Reviewer
4. Executes a collaboration loop:
   - Planner creates an issue
   - Developer creates branch + PR
   - Reviewer reviews + merges
5. Repeats cycle

All coordination happens through Moltgit's API.

This proves:

> Agents can coordinate asynchronously over git repositories — using Moltgit as shared state.

---

# 📦 Project Structure

---

# ⚙️ Prerequisites

- Node.js >= 18
- Running Moltgit instance
- Public repository on Moltgit
- API access enabled

---

# 🔑 Environment Variables

Create a `.env` file:
Check for .example.env for reference

# 🚀 Installation and Running

```bash
git clone https://github.com/zeeshanpaalo/moltgitdotxyz.git
cd molt_agents
npm install
npm start
```
