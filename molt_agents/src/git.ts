import simpleGit from "simple-git";
import * as fs from "fs";
import { config } from "./config";
import { getPersonalAccessToken } from "./tokenStore";

/**
 * Clone repository using PAT authentication
 */
export async function freshClone(
  dir: string,
  agentName: string,
  owner: string,
  repo: string,
) {
  const token = getPersonalAccessToken(agentName);

  const repoUrl = `${config.API_BASE_URL}/${owner}/${repo}.git`;
  const authUrl = repoUrl.replace("http://", `http://${token}@`);

  if (fs.existsSync(dir)) {
    fs.rmSync(dir, { recursive: true, force: true });
  }

  await simpleGit().clone(authUrl, dir);
}

/**
 * Sync fork with upstream repository
 */
export async function syncForkWithUpstream(
  dir: string,
  repoOwner: string,
  repoName: string,
) {
  const git = simpleGit(dir);

  const upstreamUrl = `${config.API_BASE_URL}/${repoOwner}/${repoName}.git`;

  // Check if upstream remote exists
  const remotes = await git.getRemotes(true);
  const hasUpstream = remotes.some((r) => r.name === "upstream");

  if (!hasUpstream) {
    await git.addRemote("upstream", upstreamUrl);
  }

  // Get the current local branch
  const branchSummary = await git.branch();
  const currentBranch = branchSummary.current;

  // Fetch all branches from upstream
  try {
    await git.fetch(["upstream"]);
  } catch (err) {
    console.log("⚠️ Failed to fetch upstream:", err);
    return;
  }

  // List all remote branches to find the default branch
  let upstreamBranch = null;
  try {
    const branches = await git.branch(["-r"]);

    // Check for upstream/main first
    if (branches.all.includes("upstream/main")) {
      upstreamBranch = "main";
    }
    // Then check for upstream/master
    else if (branches.all.includes("upstream/master")) {
      upstreamBranch = "master";
    }
    // Use current branch as fallback
    else if (branches.all.includes(`upstream/${currentBranch}`)) {
      upstreamBranch = currentBranch;
    }
  } catch (err) {
    console.log("⚠️ Failed to list upstream branches:", err);
  }

  // If we couldn't find an upstream branch, skip sync
  if (!upstreamBranch) {
    console.log("⚠️ No matching upstream branch found, skipping sync");
    return;
  }

  console.log(`🔄 Syncing with upstream/${upstreamBranch}`);

  // Ensure we're on the correct local branch
  if (currentBranch !== upstreamBranch) {
    try {
      await git.checkout(upstreamBranch);
    } catch (err) {
      // Branch might not exist locally, create it
      await git.checkoutLocalBranch(upstreamBranch);
    }
  }

  // Reset to upstream
  try {
    await git.reset(["--hard", `upstream/${upstreamBranch}`]);

    // Force push to origin (safe because fork is owned by agent)
    await git.push(["origin", upstreamBranch, "--force"]);

    console.log(`✅ Fork synced with upstream/${upstreamBranch}`);
  } catch (err) {
    console.log("⚠️ Failed to sync with upstream:", err);
  }
}

/**
 * Create a new local branch
 */
export async function createBranch(dir: string, branch: string) {
  const git = simpleGit(dir);
  await git.checkoutLocalBranch(branch);
}

/**
 * Commit and push changes
 */
export async function commitAndPush(
  dir: string,
  branch: string,
  message: string,
) {
  const git = simpleGit(dir);

  await git.add(".");
  await git.commit(message);

  await git.push(["-u", "origin", branch]);
}
