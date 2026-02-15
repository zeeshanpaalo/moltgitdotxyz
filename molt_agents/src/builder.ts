import * as fs from "fs";
import * as path from "path";
import { askLLM } from "./llm";
import {
  searchReposWithIssues,
  getIssuesForRepo,
  createPR,
  forkRepo,
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

export async function builderLoop() {
  console.log("🔍 Searching public repos with open issues...");

  const repos: any[] = await searchReposWithIssues();

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

  const issue: any = pickRandom(issues);
  console.log(`🛠 Working on issue #${issue.number}: ${issue.title}`);

  // Deterministic branch naming: {proposerName}-issue-{issueNumber}
  // This allows the reviewer to extract the issue number for closing
  const branch = `${forkOwner}-issue-${issue.number}`;
  const dir = `./workspace-${repoName}`;

  console.log(`🌿 Creating branch: ${branch}`);

  // Clone fork
  await freshClone(dir, builderUsername, forkOwner, repoName);

  // Sync fork with upstream
  await syncForkWithUpstream(dir, repoOwner, repoName);

  // Create feature branch
  await createBranch(dir, branch);

  // Read updated codebase
  const codebase = readCodebase(dir);
  const isEmpty = codebase.trim().length === 0;

  console.log(
    isEmpty
      ? "📁 Repository is empty. Will create boilerplate."
      : "📚 Codebase read successfully.",
  );

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
