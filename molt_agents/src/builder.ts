import * as fs from "fs";
import * as path from "path";
import { askLLM } from "./llm";
import {
  getIssuesForRepo,
  createPR,
  forkRepo,
  searchReposWithIssuesForPlanner,
} from "./gitea";
import {
  freshClone,
  syncForkWithUpstream,
  createBranch,
  commitAndPush,
} from "./git";

/**
 * 📖 Recursively read full repository codebase
 */
function readCodebase(dir: string): string {
  const walk = (dirPath: string): string[] =>
    fs.readdirSync(dirPath).flatMap((file) => {
      const fullPath = path.join(dirPath, file);

      if (fs.statSync(fullPath).isDirectory()) {
        if (file === ".git") return [];
        return walk(fullPath);
      }

      return [fullPath];
    });

  const files = walk(dir);

  return files
    .map((filePath) => {
      const rel = path.relative(dir, filePath);
      const content = fs.readFileSync(filePath, "utf8");
      return `FILE: ${rel}\n${content}`;
    })
    .join("\n\n");
}

/**
 * 🎲 Safe random picker
 */
function pickRandom<T>(array: T[]): T {
  return array[Math.floor(Math.random() * array.length)];
}

/**
 * 🧠 Intelligently select which issue to work on based on dependencies and priority
 */
async function selectBestIssue(
  issues: any[],
  codebase: string,
  repoOwner: string,
  repoName: string,
): Promise<any> {
  if (issues.length === 1) {
    return issues[0];
  }

  const issuesList = issues
    .map(
      (issue) =>
        `Issue #${issue.number}: ${issue.title}\nDescription: ${issue.body || "No description"}`,
    )
    .join("\n\n---\n\n");

  const system = `You are an expert software project manager and architect who understands technical dependencies.`;

  const user = `
Repository: ${repoOwner}/${repoName}

Current Codebase Status:
${codebase.trim().length === 0 ? "EMPTY REPOSITORY - No code exists yet" : `Repository has ${codebase.split("\n").length} lines of code`}

${
  codebase.trim().length > 0
    ? `Current Files:\n${codebase
        .split("FILE:")
        .slice(1)
        .map((f) => "- " + f.split("\n")[0].trim())
        .join("\n")}`
    : ""
}

Open Issues:
${issuesList}

Your task: Select the BEST issue to work on RIGHT NOW based on:
1. **Logical dependencies**: Foundation must come before features (e.g., setup before login, database before API endpoints)
2. **Current state**: What makes sense given what exists in the codebase?
3. **Blocking relationships**: Which issues block others?
4. **Natural order**: What would a smart developer tackle first?

Examples of good prioritization:
- "Project setup" BEFORE "Add authentication"
- "Database schema" BEFORE "User CRUD endpoints"
- "Basic UI structure" BEFORE "Advanced animations"
- "Core functionality" BEFORE "Optional features"

Return STRICT JSON with your reasoning:
{
  "selected_issue_number": <number>,
  "reasoning": "Brief explanation of why this issue should be tackled first"
}
`;

  try {
    const output = await askLLM(system, user);
    const parsed = JSON.parse(output);

    const selectedIssue = issues.find(
      (issue) => issue.number === parsed.selected_issue_number,
    );

    if (selectedIssue) {
      console.log(`🧠 LLM selected issue #${parsed.selected_issue_number}`);
      console.log(`💡 Reasoning: ${parsed.reasoning}`);
      return selectedIssue;
    } else {
      console.log(
        `⚠️ LLM selected invalid issue number, falling back to first issue`,
      );
      return issues[0];
    }
  } catch (err) {
    console.log(
      `❌ Failed to intelligently select issue, falling back to first issue:`,
      err,
    );
    return issues[0];
  }
}

export async function builderLoop() {
  console.log("🔍 Searching public repos with open issues...");

  const repos: any[] = await searchReposWithIssuesForPlanner();

  if (!repos?.length) {
    console.log("😴 No repos found.");
    return;
  }

  const reposWithOpenIssues = repos.filter(
    (repo: any) => repo.open_issues_count > 0,
  );

  if (!reposWithOpenIssues.length) {
    console.log("😴 No repos with open issues.");
    return;
  }

  const selectedRepo = pickRandom(reposWithOpenIssues);
  const repoOwner = selectedRepo.owner.login;
  const repoName = selectedRepo.name;

  console.log(`📦 Selected repo: ${repoOwner}/${repoName}`);

  const builderUsername = process.env.builderAgentName!;
  let forkOwner = builderUsername;

  if (repoOwner !== builderUsername) {
    console.log(`🍴 Forking repo into ${builderUsername} namespace...`);
    const fork = await forkRepo(repoOwner, repoName);
    forkOwner = fork.owner.login;
    console.log(`✅ Fork ready: ${forkOwner}/${repoName}`);
  }

  const issues = await getIssuesForRepo(repoOwner, repoName, builderUsername);

  if (!issues?.length) {
    console.log("⚠ Repo reported open issues but none returned.");
    return;
  }

  const dir = `./workspace-${repoName}`;

  // Clone fork first to read the codebase
  await freshClone(dir, builderUsername, forkOwner, repoName);

  // Sync fork with upstream
  await syncForkWithUpstream(dir, repoOwner, repoName);

  // Read current codebase to help with issue selection
  const codebase = readCodebase(dir);
  const isEmpty = codebase.trim().length === 0;

  console.log(
    isEmpty
      ? "📁 Repository is empty."
      : `📚 Repository has code (${codebase.split("\n").length} lines).`,
  );

  // Intelligently select the best issue to work on
  console.log(`🤔 Analyzing ${issues.length} open issues...`);
  const issue: any = await selectBestIssue(
    issues,
    codebase,
    repoOwner,
    repoName,
  );

  console.log(`🛠 Working on issue #${issue.number}: ${issue.title}`);

  // Deterministic branch naming: {proposerName}-issue-{issueNumber}
  // This allows the reviewer to extract the issue number for closing
  const branch = `${forkOwner}-issue-${issue.number}`;

  console.log(`🌿 Creating branch: ${branch}`);

  // Create feature branch
  await createBranch(dir, branch);

  const system = "You are a senior autonomous software engineer.";

  const user = `
Repository: ${forkOwner}/${repoName} (forked)

Original Repository: ${repoOwner}/${repoName}

Issue:
${issue.title}
${issue.body}

Current Codebase:
${codebase || "EMPTY REPOSITORY"}

If repository does not contain actual source code:
- Choose appropriate language and framework.
- Create full project boilerplate.

You must fully implement the issue.

Return STRICT JSON:
{
  "commit_message": "",
  "files": [
    { "path": "", "content": "" }
  ]
}
`;

  const output = await askLLM(system, user);

  let parsed;
  try {
    parsed = JSON.parse(output);
  } catch {
    console.log("❌ LLM returned invalid JSON.");
    return;
  }

  if (!parsed.files?.length) {
    console.log("⚠ No files returned by LLM.");
    return;
  }

  for (const file of parsed.files) {
    const fullPath = path.join(dir, file.path);
    fs.mkdirSync(path.dirname(fullPath), { recursive: true });
    fs.writeFileSync(fullPath, file.content);
  }

  await commitAndPush(dir, branch, parsed.commit_message);

  await createPR(repoOwner, repoName, issue.title, branch, forkOwner);

  console.log(
    `✅ PR created from ${forkOwner}/${repoName}:${branch} -> ${repoOwner}/${repoName}`,
  );
}
