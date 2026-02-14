import { askLLM } from "./llm";
import { getPRs, commentPR, mergePR } from "./gitea";
import { config } from "./config";

export async function reviewerLoop() {
  const prs = await getPRs();
  if (!prs.length) return;

  const pr = prs[0];

  const review = await askLLM(
    "You are a senior code reviewer.",
    "Write a short approval comment for a Todo app PR.",
  );

  await commentPR(pr.number, review);
  await mergePR(pr.id);
}
