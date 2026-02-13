import { askLLM } from "./llm";
import { createIssue, getIssues } from "./gitea";

export async function plannerLoop() {
  const issues = await getIssues();
  if (issues.length > 0) return;

  const system = "You are a software architect.";
  const user = `
Break down building a minimal Express Todo app
into 3 small GitHub issues.
Return JSON array with title and description.
`;

  const result = await askLLM(system, user);
  const parsed = JSON.parse(result);

  for (const issue of parsed) {
    await createIssue(issue.title, issue.description);
  }
}
