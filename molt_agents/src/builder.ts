import * as fs from "fs";
import * as path from "path";
import { v4 as uuid } from "uuid";
import { askLLM } from "./llm";
import { getIssues, createPR } from "./gitea";
import { freshClone, commitAndPush } from "./git";
import { config } from "./config";

export async function builderLoop() {
  const issues = await getIssues();
  if (!issues.length) return;

  const issue = issues[0];
  const branch = `issue-${issue.number}-${uuid().slice(0, 4)}`;
  const dir = "./workspace";

  await freshClone(dir, config.tokens.builder);

  const system = "You are an autonomous Node.js developer.";
  const user = `
Build code for this issue:

${issue.title}
${issue.body}

Generate JSON:
{
  branch_name: "...",
  commit_message: "...",
  files: [{ path: "...", content: "..." }]
}
`;

  const output = await askLLM(system, user);
  const parsed = JSON.parse(output);

  for (const file of parsed.files) {
    const fullPath = path.join(dir, file.path);
    fs.mkdirSync(path.dirname(fullPath), { recursive: true });
    fs.writeFileSync(fullPath, file.content);
  }

  await commitAndPush(dir, branch, parsed.commit_message);
  await createPR(config.tokens.builder, issue.title);
}
