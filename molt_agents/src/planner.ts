import { askLLM } from "./llm";
import { createIssue, createRepo, searchReposWithIssues } from "./gitea";
import { config } from "./config";

export async function plannerLoop() {
  const plannerName = config.plannerAgentName;
  const reviewerName = config.reviewerAgentName;

  // 1️⃣ Fetch existing repos that already have issues
  const existingRepos = await searchReposWithIssues();
  const existingRepoNames = existingRepos.map((r: any) => r.name);

  console.log("📚 Existing repos with issues:", existingRepoNames.join(", "));

  // 2️⃣ Ask LLM what to build
  const system = "You are an autonomous software founder.";

  // Include existing repo names in prompt to prevent duplicates
  const user = `
Propose ONE small but real web app that can be built in 3-5 issues. 
Skip any projects that already exist: [${existingRepoNames.join(", ")}]
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

  let parsed: {
    repo_name: string;
    description: string;
    issues: { title: string; description: string }[];
  };

  try {
    parsed = JSON.parse(result);
  } catch (err) {
    console.error("❌ Failed to parse LLM output:", result, err);
    return;
  }

  if (existingRepoNames.includes(parsed.repo_name)) {
    console.log(
      `⚠️ Repo "${parsed.repo_name}" already exists. Skipping creation.`,
    );
    return;
  }

  // 3️⃣ Create repo
  const repo = await createRepo(
    parsed.repo_name,
    parsed.description,
    plannerName,
    reviewerName,
  );
  console.log("📦 Created repo:", repo.full_name);

  // 4️⃣ Create issues in that repo
  for (const issue of parsed.issues) {
    await createIssue(issue.title, issue.description, plannerName, repo.name);
  }

  console.log("📝 Issues created");
}
