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
  const system = `You are a visionary software founder with a knack for identifying unique, practical problems that need solving. You think outside the box and create tools that people didn't know they needed until they saw them.`;

  // Include existing repo names in prompt to prevent duplicates
  const user = `
You need to propose ONE unique and creative web application idea that solves a real problem or provides genuine value.

AVOID these overdone categories:
- Generic todo/task managers
- Basic chat applications
- Simple blog platforms
- Standard e-commerce stores
- Basic CRUD apps
- Weather apps
- Calculator apps
- Note-taking apps

INSTEAD, think about:
- Niche tools for specific communities (musicians, teachers, designers, etc.)
- Automation tools that save time on repetitive tasks
- Creative utilities (ASCII art generators, color palette creators, sound visualizers)
- Educational tools with a unique twist
- Data visualization or analysis tools
- Fun experimental projects (multiplayer games, collaborative art, virtual pets)
- Developer tools that solve specific pain points
- Social experiments or interactive experiences
- Tools that combine unexpected technologies
- Micro-SaaS ideas that solve one thing really well

Requirements:
- Must be buildable in 3-5 well-defined issues
- Must be a complete, functional application (not just a skeleton)
- Must have a clear unique value proposition
- Name should be memorable and relevant to what it does

Projects that already exist (DON'T REPEAT THESE):
[${existingRepoNames.join(", ")}]

Return STRICT JSON:
{
  "repo_name": "",
  "description": "",
  "issues": [
    { 
      "title": "",
      "description": ""
    }
  ]
}

Be creative and think of something that would make developers say "Oh, that's actually cool and useful!"
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
