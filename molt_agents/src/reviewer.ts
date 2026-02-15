import { askLLM } from "./llm";
import {
  getAllReposForReviewer,
  getPRs,
  commentPR,
  mergePR,
  closeIssue,
} from "./gitea";
import axios from "axios";
import { config } from "./config";
import { getToken } from "./tokenStore";

/**
 * Extract issue number from branch name.
 * Expected format: ${proposerName}-issue-${issueNumber}
 * Example: builder-agent-issue-3
 */
function extractIssueNumber(branch: string): number | null {
  const match = branch.match(/-issue-(\d+)$/);
  return match ? parseInt(match[1], 10) : null;
}

async function getPRDetails(owner: string, repo: string, prNumber: number) {
  const token = getToken(config.reviewerAgentName);

  const api = axios.create({
    baseURL: `${config.API_BASE_URL}/api/v1`,
    headers: { Authorization: `Bearer ${token}` },
  });

  const [prRes, filesRes] = await Promise.all([
    api.get(`/repos/${owner}/${repo}/pulls/${prNumber}`),
    api.get(`/repos/${owner}/${repo}/pulls/${prNumber}/files`),
  ]);

  return {
    pr: prRes.data,
    files: filesRes.data,
  };
}

export async function reviewerLoop() {
  console.log("🔎 Reviewer scanning all repos...");

  const repos = await getAllReposForReviewer();
  if (!repos.length) {
    console.log("📭 No repos found for reviewer");
    return;
  }

  // Collect all PRs across all repos
  let allPRs: any[] = [];
  for (const repo of repos) {
    const prs = await getPRs(repo.owner.login, repo.name);

    for (const pr of prs) {
      allPRs.push({
        ...pr,
        owner: repo.owner.login,
        repo: repo.name,
      });
    }
  }

  if (!allPRs.length) {
    console.log("📭 No open PRs found.");
    return;
  }

  // Sort PRs by issue number extracted from branch name
  allPRs.sort((a, b) => {
    const aNum = extractIssueNumber(a.head.ref);
    const bNum = extractIssueNumber(b.head.ref);
    
    // Handle null values - push them to the end
    if (aNum === null && bNum === null) return 0;
    if (aNum === null) return 1;
    if (bNum === null) return -1;
    
    return aNum - bNum;
  });

  // Pick a PR randomly
  const randomIndex = Math.floor(Math.random() * allPRs.length);
  const pr = allPRs[randomIndex];

  console.log(
    `📝 Reviewing PR #${pr.number} from ${pr.owner}/${pr.repo}, branch ${pr.head.ref}`,
  );

  // Extract issue number from branch name
  const issueNumber = extractIssueNumber(pr.head.ref);
  if (issueNumber) {
    console.log(`🔗 This PR is linked to issue #${issueNumber}`);
  } else {
    console.log(`⚠️ Could not extract issue number from branch: ${pr.head.ref}`);
  }

  // Fetch PR details and changed files
  const { pr: prDetails, files } = await getPRDetails(
    pr.owner,
    pr.repo,
    pr.number,
  );

  const diffSummary = files
    .map(
      (f: any) =>
        `File: ${f.filename}\nAdditions: ${f.additions}\nDeletions: ${f.deletions}\nPatch:\n${f.patch || "No patch available"}`,
    )
    .join("\n\n");

  // Ask LLM for review with confidence score
  const reviewResponse = await askLLM(
    `You are a senior software architect and reviewer.
Return STRICT JSON:
{
  "approve": boolean,
  "confidence": number, // 0.0 = reject, 1.0 = fully confident to approve
  "comment": "detailed review feedback"
}

Instructions:
- Provide a confidence score for your decision to merge this PR.
- Approve only if confident; otherwise, reject.
- Always provide a meaningful comment explaining your reasoning.
- Confidence should reflect how safe and correct the code looks.
- Don't worry about test cases and super edge cases, just focus on overall code quality, correctness, and whether it addresses the issue.
`,
    `
Repo: ${pr.owner}/${pr.repo}
PR Title: ${prDetails.title}
PR Body: ${prDetails.body}

Changed Files:
${diffSummary}

Review carefully:
- Ensure it addresses the issue from branch name.
- Check architecture, code quality, and correctness.
- Provide confidence score (0.0-1.0) indicating how confident you are to merge.
- Approve if safe and mostly correct.
- Reject if there is a significant problem or unsafe code.
`,
  );

  let parsed;
  try {
    parsed = JSON.parse(reviewResponse);
  } catch {
    console.error("❌ Failed to parse LLM response. Leaving safe comment.");
    await commentPR(
      pr.owner,
      pr.repo,
      pr.number,
      "Reviewer failed to parse analysis. Manual review required.",
    );
    return;
  }

  // Comment PR
  await commentPR(pr.owner, pr.repo, pr.number, parsed.comment);
  
  const CONFIDENCE_CUTOFF = 0.5;
  
  // Decide merge based on confidence
  if (parsed.confidence >= CONFIDENCE_CUTOFF) {
    console.log(
      `✅ Confidence ${parsed.confidence} >= ${CONFIDENCE_CUTOFF}. Merging PR #${pr.number}`,
    );
    
    try {
      await mergePR(pr.owner, pr.repo, pr.number);
      console.log(`✅ PR #${pr.number} merged successfully`);
      
      // Close the associated issue if we have the issue number
      if (issueNumber) {
        console.log(`🔒 Closing issue #${issueNumber}...`);
        await closeIssue(pr.owner, pr.repo, issueNumber);
      }
    } catch (err: any) {
      console.error(
        `❌ Failed to merge PR #${pr.number}:`,
        err.response?.data || err.message,
      );
    }
  } else {
    console.log(
      `🛑 PR #${pr.number} not merged. Confidence ${parsed.confidence} < ${CONFIDENCE_CUTOFF}`,
    );
  }
}