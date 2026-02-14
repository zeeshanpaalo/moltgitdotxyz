import { askLLM } from "./llm";
import { createIssue, createRepo } from "./gitea";
import { config } from "./config";

export async function plannerLoop() {
  // const existingIssues = await getIssues().catch(() => []);

  // If repo already active, don't spawn new one
  // if (existingIssues.length > 0) return;

//   const repos = await listRepos();
// const existing = repos.map(r => r.name).join(", ");

  // Ask LLM what to build
  const system = "You are an autonomous software founder.";
  const user = `
Propose ONE small but real web app that can be built in 3-5 issues. Don't repeat the same idea
Return JSON:

{
  "repo_name": "",
  "description": "",
  "issues": [
    { "title": "", "description": "" }
  ]
}
`;

  const result = await askLLM(system, user);

  const parsed = JSON.parse(result);

  console.log(parsed);

  // 2️⃣ Create repo
  const repo = await createRepo(parsed.repo_name, parsed.description);

  console.log("📦 Created repo:", repo.full_name);
  // good

  // console.log(parsed.issues);

  // 3️⃣ Create issues in that repo
  for (const issue of parsed.issues) {
    await createIssue(issue.title, issue.description, "planner-1", repo.name);
  }

  console.log("📝 Issues created");
}
