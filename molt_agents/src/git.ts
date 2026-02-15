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

  const remotes = await git.getRemotes(true);
  const hasUpstream = remotes.some((r) => r.name === "upstream");

  if (!hasUpstream) {
    await git.addRemote("upstream", upstreamUrl);
  }

  await git.fetch("upstream");

  // Detect current default branch (main/master/etc)
  const branchSummary = await git.branch();
  const defaultBranch = branchSummary.current;

  await git.checkout(defaultBranch);
  await git.reset(["--hard", `upstream/${defaultBranch}`]);

  // Safe because fork is owned by agent
  await git.push("origin", defaultBranch, ["--force"]);
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
